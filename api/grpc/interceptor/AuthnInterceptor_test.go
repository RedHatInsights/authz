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
	validIssuer1    = "example.com/issuer"
	validAudience1  = "example.com"
	defaultSubject1 = "u1"

	validIssuer2    = "classy.com/issuer"
	validAudience2  = "classy.com"
	defaultSubject2 = "u2"

	minimumScope = "openid"
	testKID      = "test-kid"
)

func TestInterceptorHoldsValuesFromDiscoveryEndpoint(t *testing.T) {
	interceptor := AuthnInterceptor{[]*authnProvider{createAuthnProvider1()}}

	result, err := interceptor.validateTokenAndExtractSubject(createToken(createDefaultTokenBuilder1(), tokenSigningKey1))

	assert.NoError(t, err)
	assert.Equal(t, defaultSubject1, result.SubjectID)
}

func TestInterceptorHoldsValuesFromSecondDiscoveryEndpoint(t *testing.T) {
	interceptor := AuthnInterceptor{[]*authnProvider{
		createAuthnProvider1(), createAuthnProvider2(),
	}}

	result, err := interceptor.validateTokenAndExtractSubject(createToken(createDefaultTokenBuilder2(), tokenSigningKey2))

	assert.NoError(t, err)
	assert.Equal(t, defaultSubject2, result.SubjectID)
}

func TestInterceptorHoldsValuesFromFirstDiscoveryEndpoint(t *testing.T) {
	interceptor := AuthnInterceptor{[]*authnProvider{
		createAuthnProvider1(), createAuthnProvider2(),
	}}

	result, err := interceptor.validateTokenAndExtractSubject(createToken(createDefaultTokenBuilder1(), tokenSigningKey1))

	assert.NoError(t, err)
	assert.Equal(t, defaultSubject1, result.SubjectID)
}

func TestAllOkWhen2SameProviders(t *testing.T) {
	interceptor := AuthnInterceptor{[]*authnProvider{
		createAuthnProvider1(), createAuthnProvider1(),
	}}

	result, err := interceptor.validateTokenAndExtractSubject(createToken(createDefaultTokenBuilder1(), tokenSigningKey1))

	assert.NoError(t, err)
	assert.Equal(t, defaultSubject1, result.SubjectID)
}

func TestFailedValidationWhenAuthnProviderAbsent(t *testing.T) {
	interceptor := AuthnInterceptor{[]*authnProvider{
		createAuthnProvider2(), createAuthnProvider2(),
	}}

	_, err := interceptor.validateTokenAndExtractSubject(createToken(createDefaultTokenBuilder1(), tokenSigningKey1))

	assert.Error(t, err) // TODO: Would like to assert domain.ErrNotAuthenticated, but fails on signature verification error
}

func TestAuthnProviderHoldsValuesFromDiscoveryEndpoint(t *testing.T) {
	authnProvider := createAuthnProvider1()

	result, err := validateTokenAndExtractSubject(authnProvider, createToken(createDefaultTokenBuilder1(), tokenSigningKey1))

	assert.NoError(t, err)
	assert.Equal(t, defaultSubject1, result.SubjectID)
}

func TestInvalidTokenMissingSubject(t *testing.T) {
	authnProvider := createAuthnProvider1()

	builder := jwt.NewBuilder().Audience([]string{validAudience1}).IssuedAt(time.Now()).Issuer(validIssuer1)
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder, tokenSigningKey1))

	assert.ErrorIs(t, err, domain.ErrNotAuthenticated)
}

func TestInvalidTokenExpired(t *testing.T) {
	authnProvider := createAuthnProvider1()

	builder := createDefaultTokenBuilder1().
		NotBefore(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		Expiration(time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder, tokenSigningKey1))

	assert.ErrorIs(t, err, jwt.ErrTokenExpired())
}

func TestInvalidTokenFromTheFuture(t *testing.T) {
	authnProvider := createAuthnProvider1()

	builder := createDefaultTokenBuilder1().
		NotBefore(time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC)).
		Expiration(time.Date(2200, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder, tokenSigningKey1))

	assert.ErrorIs(t, err, jwt.ErrTokenNotYetValid())
}

func TestInvalidAudience(t *testing.T) {
	authnProvider := createAuthnProvider1()

	builder := createDefaultTokenBuilder1().
		Audience([]string{"invalid-audience"})
	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder, tokenSigningKey1))

	assert.ErrorIs(t, err, jwt.ErrInvalidAudience())
}

func TestInvalidIssuer(t *testing.T) {
	authnProvider := createAuthnProvider1()

	builder := createDefaultTokenBuilder1().Issuer("example.com/invalidissuer")

	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder, tokenSigningKey1))

	assert.ErrorIs(t, err, jwt.ErrInvalidIssuer())
}

func TestInvalidTokenMissingScope(t *testing.T) {
	authnProvider := createAuthnProvider1()

	builder := jwt.NewBuilder().Audience([]string{validAudience1}).IssuedAt(time.Now()).Issuer(validIssuer1).Subject(defaultSubject1)

	_, err := validateTokenAndExtractSubject(authnProvider, createToken(builder, tokenSigningKey1))

	assert.Error(t, err)
}

func TestInvalidTokenWrongSigningKey(t *testing.T) {
	authnProvider := createAuthnProvider1()

	data, err := createDefaultTokenBuilder1().Build()
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
	authnProvider := createAuthnProvider1()

	token := createToken(createDefaultTokenBuilder1(), tokenSigningKey1)

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

func createDefaultTokenBuilder1() *jwt.Builder {
	return jwt.NewBuilder().
		Subject(defaultSubject1).
		IssuedAt(time.Now()).
		Audience([]string{validAudience1}).
		Issuer(validIssuer1).
		Claim("scope", minimumScope)
}

func createDefaultTokenBuilder2() *jwt.Builder {
	return jwt.NewBuilder().
		Subject(defaultSubject2).
		IssuedAt(time.Now()).
		Audience([]string{validAudience2}).
		Issuer(validIssuer2).
		Claim("scope", minimumScope)
}

func createAuthnProvider1() *authnProvider {
	keyset := jwk.NewSet()

	err := keyset.AddKey(tokenVerificationKey1)
	if err != nil {
		panic(err)
	}

	return newAuthnProviderFromData(
		validIssuer1,
		validAudience1,
		minimumScope,
		keyset)
}

func createAuthnProvider2() *authnProvider {
	keyset2 := jwk.NewSet()

	err := keyset2.AddKey(tokenVerificationKey2)
	if err != nil {
		panic(err)
	}

	return newAuthnProviderFromData(
		validIssuer2,
		validAudience2,
		minimumScope,
		keyset2)
}

var tokenSigningKey1 jwk.Key
var tokenVerificationKey1 jwk.Key

var tokenSigningKey2 jwk.Key
var tokenVerificationKey2 jwk.Key

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

func createToken(builder *jwt.Builder, tokenSigningKey jwk.Key) string {
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
	tokenSigningKey1, tokenVerificationKey1 = generateKeys()
	tokenSigningKey2, tokenVerificationKey2 = generateKeys()

	result := m.Run()
	os.Exit(result)
}
