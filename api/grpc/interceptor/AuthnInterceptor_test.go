package interceptor

import (
	"authz/domain"
	"crypto/rand"
	"crypto/rsa"
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
	authnProvider := createAuthnProvider()

	result, err := validateTokenAndExtractSubject(authnProvider, createToken(createDefaultTokenBuilder()))

	assert.NoError(t, err)
	assert.Equal(t, defaultSubject, result.SubjectID)
}

func TestInvalidTokenMissingSubject(t *testing.T) {
	authnProvider := createAuthnProvider()

	builder := jwt.NewBuilder().Audience([]string{validAudience}).IssuedAt(time.Now()).Issuer(validIssuer)
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder))

	assert.ErrorIs(t, err, domain.ErrNotAuthenticated)
}

func TestInvalidTokenExpired(t *testing.T) {
	authnProvider := createAuthnProvider()

	builder := createDefaultTokenBuilder().
		NotBefore(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		Expiration(time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrTokenExpired())
}

func TestInvalidTokenFromTheFuture(t *testing.T) {
	authnProvider := createAuthnProvider()

	builder := createDefaultTokenBuilder().
		NotBefore(time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)).
		Expiration(time.Date(2200, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrTokenNotYetValid())
}

func TestInvalidAudience(t *testing.T) {
	authnProvider := createAuthnProvider()

	builder := createDefaultTokenBuilder().
		Audience([]string{"invalid-audience"})
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrInvalidAudience())
}

func TestInvalidIssuer(t *testing.T) {
	authnProvider := createAuthnProvider()

	builder := createDefaultTokenBuilder().Issuer("example.com/invalidissuer")

	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder))

	assert.ErrorIs(t, err, jwt.ErrInvalidIssuer())
}

func TestInvalidTokenMissingScope(t *testing.T) {
	authnProvider := createAuthnProvider()

	builder := jwt.NewBuilder().Audience([]string{validAudience}).IssuedAt(time.Now()).Issuer(validIssuer).Subject(defaultSubject)

	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder))

	assert.Error(t, err)
}

func TestInvalidTokenWrongSigningKey(t *testing.T) {
	authnProvider := createAuthnProvider()

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
	err = maliciousSigning.Set(jwk.KeyIDKey, testKID)
	assert.NoError(t, err)

	token, err := jwt.Sign(data, jwt.WithKey(jwa.RS256, maliciousSigning))
	assert.NoError(t, err)

	_, err = validateTokenAndExtractSubject(authnProvider, string(token))

	assert.Error(t, err) //No specific error for this. See: https://github.com/lestrrat-go/jwx/blob/0121992a0875d2263d99cc90c676276e143580a6/jws/jws.go#L412
}

func TestInvalidTokenTampered(t *testing.T) {
	authnProvider := createAuthnProvider()

	token := createToken(createDefaultTokenBuilder())

	parts := strings.Split(token, ".")
	bodyData, err := base64.RawStdEncoding.DecodeString(parts[1]) //decode body
	if err != nil {
		panic(err)
	}

	bodyJSON := string(bodyData)
	bodyJSON = strings.Replace(bodyJSON, `"u1"`, `"admin"`, 1)

	parts[1] = base64.RawStdEncoding.EncodeToString([]byte(bodyJSON))

	tamperedToken := strings.Join(parts, ".")

	_, err = validateTokenAndExtractSubject(authnProvider, tamperedToken)

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

func createAuthnProvider() *authnProvider {
	keyset := jwk.NewSet()

	err := keyset.AddKey(tokenVerificationKey)
	if err != nil {
		panic(err)
	}

	return newAuthnProviderFromData(
		validIssuer,
		validAudience,
		minimumScope,
		keyset)
}

var tokenSigningKey jwk.Key
var tokenVerificationKey jwk.Key

func generateKeys() (signing jwk.Key, verification jwk.Key) {
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

	public := private.Public()
	verification, err = jwk.FromRaw(public)
	if err != nil {
		panic(err)
	}
	err = verification.Set(jwk.KeyIDKey, testKID)
	if err != nil {
		panic(err)
	}
	err = verification.Set(jwk.AlgorithmKey, jwa.RS256)
	if err != nil {
		panic(err)
	}

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

	return string(token)
}

func TestMain(m *testing.M) {
	tokenSigningKey, tokenVerificationKey = generateKeys()

	result := m.Run()
	os.Exit(result)
}
