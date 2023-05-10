// Package interceptor contains middleware interceptors for unary and stream. Interceptors are applied to calls from HTTP and GRPC
package interceptor

import (
	"authz/api"
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
)

type providerJSON struct {
	Issuer  string `json:"issuer"`
	JwksURL string `json:"jwks_uri"`
}

// NewAuthnInterceptor constructor
func NewAuthnInterceptor(config api.AuthConfig) (*AuthnInterceptor, error) {
	providerData, err := getProviderData(config.DiscoveryEndpoint) //TODO: not sure about making a web request from a constructor
	if err != nil {
		return nil, err
	}

	keyCache, err := createKeyCache(providerData.JwksURL)
	if err != nil {
		return nil, err
	}

	return newAuthnInterceptorFromData(providerData.Issuer, config.Audience, config.RequiredScope, keyCache), nil
}

func newAuthnInterceptorFromData(issuer string, audience string, minimumScope string, keys jwk.Set) *AuthnInterceptor {
	return &AuthnInterceptor{issuer: issuer, audience: audience, minimumScope: minimumScope, verificationKeys: keys}
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
		err = fmt.Errorf("Unsuccessful response when retrieving provider configuration data from %s: %s", discoveryEndpoint, resp.Status)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	return
}

// Unary impl of the Unary interceptor
func (r *AuthnInterceptor) Unary() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		glog.Info("Hello from AuthnInterceptor %v: ", req)
		token := getBearerTokenFromContext(ctx)

		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
		}

		result, err := r.validateTokenAndExtractSubject(token)
		if err != nil {
			glog.Errorf("Error processing token: %s", err)
			return nil, status.Error(codes.Unauthenticated, "Invalid or expired identity token.")
		}

		return handler(context.WithValue(ctx, RequestorContextKey, result.SubjectID), req)
	})
}

func (r *AuthnInterceptor) validateTokenAndExtractSubject(token string) (result tokenIntrospectionResult, err error) {
	jwtoken, err := jwt.ParseString(token, jwt.WithVerify(false), jwt.WithKeySet(r.verificationKeys), jwt.WithIssuer(r.issuer), jwt.WithAudience(r.audience))
	if err != nil {
		return
	}

	err = ensureRequiredScope(r.minimumScope, jwtoken)
	if err != nil {
		return
	}

	//TODO Validate the existence of the subject ?
	result.SubjectID = jwtoken.Subject()
	if result.SubjectID == "" {
		err = domain.ErrNotAuthenticated
	}
	return
}

func ensureRequiredScope(requiredScope string, token jwt.Token) error {
	scopesClaim, ok := token.Get("scope")
	if !ok {
		return domain.ErrNotAuthenticated //No scopes present
	}

	scopesString, ok := scopesClaim.(string)
	if !ok {
		return domain.ErrNotAuthenticated //Scope(s) present but not a string??
	}

	for _, scope := range strings.Split(scopesString, " ") {
		if scope == requiredScope {
			return nil
		}
	}

	return domain.ErrNotAuthenticated //Scope not present
}

type tokenIntrospectionResult struct {
	SubjectID string
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
