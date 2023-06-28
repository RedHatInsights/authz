//go:build !release

package testenv

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/mendsley/gojwk"
)

const (
	testKID    = "test-kid"
	testIssuer = "http://localhost:8180/idp"

	testAudience      = "cloud-services"
	testRequiredScope = "openid"
)

// OidcDiscoveryURL OIDC Discovery endpoints URL of the fake IDP
var OidcDiscoveryURL string

var tokenSigningKey, tokenVerificationKey = generateKeys()

// HostFakeIdp hosts a fake OIDC identity provider that has an OIDC discovery endpoint and is able to validate certs.
func HostFakeIdp() {
	mux := http.NewServeMux()

	mux.Handle("/idp/.well-known/openid-configuration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf(`{
			"issuer": "%s",
			"authorization_endpoint": "http://localhost:8180/idp/authorize",
			"token_endpoint": "http://localhost:8180/idp/token",
			"userinfo_endpoint": "http://localhost:8180/idp/userinfo",
			"jwks_uri": "http://localhost:8180/idp/certs",
			"scopes_supported": [
				"openid"
			],
			"response_types_supported": [
				"code",
				"id_token",
				"token id_token"
			],
			"token_endpoint_auth_methods_supported": [
				"client_secret_basic"
			]
		}`, testIssuer))) //Modified from an example OIDC discovery document: https://swagger.io/docs/specification/authentication/openid-connect-discovery/
		if err != nil {
			panic(err)
		}
	}))

	mux.Handle("/idp/certs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		pubjwk, err := gojwk.PublicKey(tokenVerificationKey)
		if err != nil {
			panic(err)
		}

		pubjwk.Alg = "RS256"
		pubjwk.Kid = testKID
		serializedKey, err := gojwk.Marshal(pubjwk)
		if err != nil {
			panic(err)
		}

		response := fmt.Sprintf(`{"keys": [%s]}`, string(serializedKey))

		_, err = w.Write([]byte(response))
		if err != nil {
			panic(err)
		}
	}))

	OidcDiscoveryURL = "http://localhost:8180/idp/.well-known/openid-configuration"
	err := http.ListenAndServe("localhost:8180", mux)
	if err != nil {
		panic(err)
	}
}

// SubjectIDToToken creates a new jwt token for a given SubjectID
func CreateToken(subject string, orgID string, isOrgAdmin bool) string {
	if subject == "" {
		return ""
	}

	data, err := jwt.NewBuilder().
		Issuer(testIssuer).
		IssuedAt(time.Now()).
		Audience([]string{testAudience}).
		Subject(subject).
		Claim("scope", testRequiredScope).
		Claim("org_id", orgID).
		Claim("is_org_admin", isOrgAdmin).
		Build()

	if err != nil {
		panic(err)
	}

	token, err := jwt.Sign(data, jwt.WithKey(jwa.RS256, tokenSigningKey))

	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("bearer %s", token)
}

func generateKeys() (signing jwk.Key, verification crypto.PublicKey) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	signing, err = jwk.FromRaw(private)
	if err != nil {
		panic(err)
	}
	err = signing.Set(jwk.KeyIDKey, testKID)
	if err != nil {
		panic(err)
	}

	verification = private.Public()

	return
}
