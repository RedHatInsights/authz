package interceptor

import (
	"authz/domain"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"encoding/base64"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/stretchr/testify/assert"
)

const (
	validIssuer    = "example.com/issuer"
	validAudience  = "example.com"
	minimumScope   = "openid"
	defaultSubject = "u1"
	testKID        = "test-kid"
)

func TestInterceptorHoldsValuesFromDiscoveryEndpoint(t *testing.T) {
	interceptor := createInterceptor()

	result, err := interceptor.validateTokenAndExtractSubject(createToken(createDefaultTokenBuilder()))

	assert.NoError(t, err)
	assert.Equal(t, defaultSubject, result.SubjectID)
}

func TestInvalidTokenMissingSubject(t *testing.T) {
	interceptor := createInterceptor()

	builder := jwt.NewBuilder().Audience([]string{validAudience}).IssuedAt(time.Now()).Issuer(validIssuer)
	_, err := interceptor.validateTokenAndExtractSubject(createToken(builder))

	assert.ErrorIs(t, err, domain.ErrNotAuthenticated)
}

func TestInvalidTokenExpired(t *testing.T) {
	interceptor := createInterceptor()

	builder := createDefaultTokenBuilder().
		NotBefore(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		Expiration(time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err := interceptor.validateTokenAndExtractSubject(createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrTokenExpired())
}

func TestInvalidTokenFromTheFuture(t *testing.T) {
	interceptor := createInterceptor()

	builder := createDefaultTokenBuilder().
		NotBefore(time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)).
		Expiration(time.Date(2200, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err := interceptor.validateTokenAndExtractSubject(createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrTokenNotYetValid())
}

func TestInvalidAudience(t *testing.T) {
	interceptor := createInterceptor()

	builder := createDefaultTokenBuilder().
		Audience([]string{"invalid-audience"})
	_, err := interceptor.validateTokenAndExtractSubject(createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrInvalidAudience())
}

func TestInvalidIssuer(t *testing.T) {
	interceptor := createInterceptor()

	builder := createDefaultTokenBuilder().Issuer("example.com/invalidissuer")

	_, err := interceptor.validateTokenAndExtractSubject(createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrInvalidIssuer())
}

func TestInvalidTokenMissingScope(t *testing.T) {
	interceptor := createInterceptor()

	builder := jwt.NewBuilder().Audience([]string{validAudience}).IssuedAt(time.Now()).Issuer(validIssuer).Subject(defaultSubject)

	_, err := interceptor.validateTokenAndExtractSubject(createToken(builder))

	assert.Error(t, err)
}

func TestInvalidTokenWrongSigningKey(t *testing.T) {
	interceptor := createInterceptor()

	data, err := createDefaultTokenBuilder().Build()
	if err != nil {
		panic(err)
	}

	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	maliciousSigning, err := jwk.FromRaw(private)
	if err != nil {
		panic(err)
	}
	maliciousSigning.Set(jwk.KeyIDKey, testKID)

	token, err := jwt.Sign(data, jwt.WithKey(jwa.RS256, maliciousSigning))

	_, err = interceptor.validateTokenAndExtractSubject(string(token))

	assert.Error(t, err) //No specific error for this. See: https://github.com/lestrrat-go/jwx/blob/0121992a0875d2263d99cc90c676276e143580a6/jws/jws.go#L412
}

func TestInvalidTokenTampered(t *testing.T) {
	interceptor := createInterceptor()

	token := createToken(createDefaultTokenBuilder())

	parts := strings.Split(token, ".")
	bodyData, err := base64.RawStdEncoding.DecodeString(parts[1]) //decode body

	bodyJSON := string(bodyData)
	bodyJSON = strings.Replace(bodyJSON, `"u1"`, `"admin"`, 1)

	parts[1] = base64.RawStdEncoding.EncodeToString([]byte(bodyJSON))

	tamperedToken := strings.Join(parts, ".")

	_, err = interceptor.validateTokenAndExtractSubject(tamperedToken)

	assert.Error(t, err) //No specific error for this. See: https://github.com/lestrrat-go/jwx/blob/0121992a0875d2263d99cc90c676276e143580a6/jws/jws.go#L412
}

func createDefaultTokenBuilder() *jwt.Builder {
	return jwt.NewBuilder().
		Subject(defaultSubject).
		IssuedAt(time.Now()).
		Audience([]string{validAudience}).
		Issuer(validIssuer).
		Claim("scope", minimumScope)
}

func createInterceptor() *AuthnInterceptor {
	keyset := jwk.NewSet()

	keyset.AddKey(tokenVerificationKey)

	return newAuthnInterceptorFromData(
		validIssuer,
		validAudience,
		minimumScope,
		keyset)
}

var tokenSigningKey jwk.Key
var tokenVerificationKey jwk.Key
var testIdpHTTPClient *http.Client

func generateKeys() (signing jwk.Key, verification jwk.Key) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	signing, err = jwk.FromRaw(private)
	if err != nil {
		panic(err)
	}
	signing.Set(jwk.KeyIDKey, testKID)

	public := private.Public()
	verification, err = jwk.FromRaw(public)
	if err != nil {
		panic(err)
	}
	verification.Set(jwk.KeyIDKey, testKID)
	verification.Set(jwk.AlgorithmKey, jwa.RS256)

	return
}

func createToken(builder *jwt.Builder) string {
	data, err := builder.Build()

	if err != nil {
		panic(err)
	}
	token, err := jwt.Sign(data, jwt.WithKey(jwa.RS256, tokenSigningKey))

	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s", token)
}

func createFakeIdp() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/idp/.well-known/openid-configuration":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(``))
			break
		case "/idp/certs":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(``))
			break
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte{})
		}
	}))
}

func TestMain(m *testing.M) {
	tokenSigningKey, tokenVerificationKey = generateKeys()

	idp := createFakeIdp()
	testIdpHTTPClient = idp.Client()
	result := m.Run()
	idp.Close()
	os.Exit(result)
}
