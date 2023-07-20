// Package interceptor contains middleware interceptors for unary and stream. Interceptors are applied to calls from HTTP and GRPC
package interceptor

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthnInterceptor - Middleware to validate incoming bearer tokens
type AuthnInterceptor struct {
	providers []*authnProvider
}

type authnProvider struct {
	issuer           string
	audience         string
	minimumScope     string
	verificationKeys jwk.Set
}

// ContextKey Type to hold Keys that are applied to the request context
type ContextKey string

const (
	// RequestorContextKey Key for the Requestor value
	RequestorContextKey ContextKey = ContextKey("Requestor")
	// RequestorOrgContextKey Key for Requestor org
	RequestorOrgContextKey ContextKey = ContextKey("RequestorOrgContextKey")
	// IsRequestorOrgAdminContextKey Key for asserting whether requestor is admin of the above org
	IsRequestorOrgAdminContextKey ContextKey = ContextKey("IsOrgAdmin")
)

type providerJSON struct {
	Issuer  string `json:"issuer"`
	JwksURL string `json:"jwks_uri"`
}

// NewAuthnInterceptor creates a new AuthnInterceptor
func NewAuthnInterceptor(configs []serviceconfig.AuthConfig) (*AuthnInterceptor, error) {
	if err := validateAuthConfigs(configs); err != nil {
		return nil, err
	}

	var providers []*authnProvider
	for _, config := range configs {
		provider, err := configureAuthnProvider(config)
		if err != nil {
			return nil, fmt.Errorf("error creating authn provider")
		}

		providers = append(providers, provider)
	}

	return &AuthnInterceptor{providers}, nil
}

func validateAuthConfigs(configs []serviceconfig.AuthConfig) (err error) {
	if configs == nil {
		return fmt.Errorf("configs should not be nil")
	}

	checked := make(map[serviceconfig.AuthConfig]bool)

	for _, config := range configs {
		if config == (serviceconfig.AuthConfig{}) {
			return fmt.Errorf("authnconfig should not be empty")
		}

		if checked[config] {
			return fmt.Errorf("authnconfigs should not have duplicates: %v", config)
		}

		checked[config] = true
	}

	return
}

// configureAuthnProvider constructor
func configureAuthnProvider(config serviceconfig.AuthConfig) (*authnProvider, error) {
	providerData, err := getProviderData(config.DiscoveryEndpoint)
	if err != nil {
		return nil, err
	}

	keyCache, err := createKeyCache(providerData.JwksURL)
	if err != nil {
		return nil, err
	}

	return newAuthnProviderFromData(providerData.Issuer, config.Audience, config.RequiredScope, keyCache), nil
}

func newAuthnProviderFromData(issuer string, audience string, minimumScope string, keys jwk.Set) *authnProvider {
	return &authnProvider{issuer: issuer, audience: audience, minimumScope: minimumScope, verificationKeys: keys}
}

func createKeyCache(jwksURL string) (jwk.Set, error) {
	cache := jwk.NewCache(context.Background())

	err := cache.Register(jwksURL)
	if err != nil {
		return nil, err
	}

	return jwk.NewCachedSet(cache, jwksURL), nil
}

func getProviderData(discoveryEndpoint string) (data providerJSON, err error) {
	req, err := http.NewRequest(http.MethodGet, discoveryEndpoint, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unsuccessful response when retrieving provider configuration data from %s: %s", discoveryEndpoint, resp.Status)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	return
}

// Unary impl of the Unary interceptor
func (authnInterceptor *AuthnInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		if info.FullMethod == "/api.v1alpha.HealthCheckService/HealthCheck" {
			return handler(ctx, req)
		}
		token := getBearerTokenFromContext(ctx)

		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
		}

		result, err := authnInterceptor.validateTokenAndExtractData(token)

		if err != nil {
			glog.Errorf("Error processing token: %s", err)
			return nil, status.Error(codes.Unauthenticated, "Invalid or expired identity token.")
		}

		// Context with multiple values - derive a context from context
		ctx = context.WithValue(ctx, RequestorContextKey, result.SubjectID)
		ctx = context.WithValue(ctx, RequestorOrgContextKey, result.Org)
		ctx = context.WithValue(ctx, IsRequestorOrgAdminContextKey, result.IsOrgAdmin)

		return handler(ctx, req)
	}
}

func (authnInterceptor *AuthnInterceptor) validateTokenAndExtractData(token string) (result tokenIntrospectionResult, err error) {
	jwtoken, err := jwt.ParseString(token, jwt.WithVerify(false), jwt.WithValidate(false)) //Parse without any validation to peek issuer

	if err != nil {
		return
	}

	issuer := jwtoken.Issuer()

	for _, provider := range authnInterceptor.providers {
		if issuer == provider.issuer {
			result, err = validateTokenAndExtractData(provider, token)

			return
		}
	}

	err = jwt.ErrInvalidIssuer()

	return
}

func validateTokenAndExtractData(p *authnProvider, token string) (result tokenIntrospectionResult, err error) {
	//Parse with signature verification and token validation. Second parse is necessary because WithKeySet cannot be passed to jwt.Validate
	jwtToken, err := jwt.ParseString(token, jwt.WithKeySet(p.verificationKeys), jwt.WithIssuer(p.issuer), jwt.WithAudience(p.audience))

	if err != nil {
		return
	}

	err = ensureRequiredScope(p.minimumScope, jwtToken)
	if err != nil {
		return
	}

	result.SubjectID = jwtToken.Subject()
	result.Org, result.IsOrgAdmin = requestorOrg(jwtToken)

	if result.SubjectID == "" {
		err = fmt.Errorf("%w: no sub claim found", domain.ErrNotAuthenticated)
	}

	return
}

func ensureRequiredScope(requiredScope string, token jwt.Token) error {
	scopesClaim, ok := token.Get("scope")
	if !ok {
		return fmt.Errorf("%w: no scopes present", domain.ErrNotAuthenticated) //No scopes present
	}

	scopesString, ok := scopesClaim.(string)
	if !ok {
		return fmt.Errorf("%w: unable to decode scopes", domain.ErrNotAuthenticated) //Scope(s) present but not a string??
	}

	for _, scope := range strings.Split(scopesString, " ") {
		if scope == requiredScope {
			return nil
		}
	}

	return fmt.Errorf("%w: required scope %s not present", domain.ErrNotAuthenticated, requiredScope) //Scope not present
}

type tokenIntrospectionResult struct {
	SubjectID  string
	Org        string
	IsOrgAdmin bool
}

func getBearerTokenFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, name := range []string{"grpcgateway-authorization", "authorization"} {
			headers := md.Get(name)
			if len(headers) > 0 {
				value := headers[0]
				parts := strings.Split(value, " ")

				if len(parts) > 1 {
					return parts[1]
				}

				return parts[0]
			}
		}
	}
	return ""
}

func requestorOrg(token jwt.Token) (orgID string, isOrgAdmin bool) {
	if claims := token.PrivateClaims(); claims != nil {
		if claims["org_id"] != nil {
			orgID = claims["org_id"].(string)
		}
		if claims["is_org_admin"] != nil {
			isOrgAdmin = claims["is_org_admin"].(bool)
		}
	}
	return orgID, isOrgAdmin
}
