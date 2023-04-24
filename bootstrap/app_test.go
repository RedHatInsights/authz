package bootstrap

import (
	"authz/infrastructure/repository/authzed"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/golang/glog"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/mendsley/gojwk"

	"github.com/bradhe/stopwatch"
	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

// container to get the port when re-initializing the service
var container *authzed.LocalSpiceDbContainer

const (
	testKID           = "test-kid"
	testIssuer        = "http://localhost:8180/idp"
	testAudience      = "cloud-services"
	testRequiredScope = "openid"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	t.SkipNow() //Skip until meta-authz is in place
	setupService()
	defer teardownService()

	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "bad",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckAccess(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestCheckReturnsTrueWhenUserAuthorized(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckReturnsFalseWhenUserNotAuthorized(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "not_authorized", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)
}

func TestAssignLicenseReturnsSuccess(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "system",
		`{
			"assign": [
			  "okay"
			]
		  }`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestUnassignLicenseReturnsSuccess(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "system",
		`{
			"unassign": [
			  "u1"
			]
		}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestGrantedLicenseAllowsUse(t *testing.T) {
	setupService()
	defer teardownService()
	//The user isn't licensed initially, use is denied
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)

	//Grant a license
	resp, err = http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "okay",
		`{
		"assign": [
			"u2"
			]
			}`))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{}`)
	container.WaitForQuantizationInterval()

	//Should be allowed now
	resp, err = http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestGrantedLicenseAffectsCountsAndDetails(t *testing.T) {
	setupService()
	defer teardownService()

	resp, err := http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts", "system"))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":9, "seatsTotal": 10}`)

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts/seats", "system"))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"users": [{"assigned":true,"displayName":"O1 User 1","id":"u1"}]}`)

	//Grant a license
	resp, err = http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "okay",
		`{
			"assign": [
			  "okay"
			]
		  }`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	container.WaitForQuantizationInterval()

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts", "system"))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":8, "seatsTotal": 10}`)

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts/seats", "system"))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"users": "<<PRESENCE>>"}`)

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts/seats?filter=assignable", "token"))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"users":"<<PRESENCE>>"}`)
}

func TestOverAssigningLicensesFails(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "okay",
		`{
		"assign": [
			"user1",
			"user2",
			"user3",
			"user4",
			"user5",
			"user6",
			"user7",
			"user8",
			"user9",
			"user10"
		]
	}`))

	assert.NoError(t, err)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestCors_NotImplementedMethod(t *testing.T) {
	setupService()
	defer teardownService()
	body := `{
			"assign": [
			  "okay"
			]
		  }`

	req := createRequest(http.MethodTrace, "/v1alpha/orgs/o1/licenses/smarts", "okay", body)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, 501)
}

func TestCors_AllowAllOrigins(t *testing.T) {
	setupService()
	defer teardownService()
	body := `{
			"assign": [
			  "okay"
			]
		  }`

	req := post("/v1alpha/orgs/o1/licenses/smarts", "okay", body)
	req.Header.Set("AllowAllOrigins", "true")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, resp.Header.Get("Vary"), "Origin")
	assertJSONResponse(t, resp, 200, `{}`)
}

func assertJSONResponse(t *testing.T, resp *http.Response, statusCode int, template string, args ...interface{}) {
	if assert.NotNil(t, resp) {
		assert.Equal(t, statusCode, resp.StatusCode)

		payload := new(strings.Builder)
		_, err := io.Copy(payload, resp.Body)
		assert.NoError(t, err)

		ja := jsonassert.New(t)
		ja.Assertf(payload.String(), template, args...)
	}
}

func get(relativeURI string, subject string) *http.Request {
	return createRequest(http.MethodGet, relativeURI, subjectIDToToken(subject), "")
}

func post(relativeURI string, subject string, body string) *http.Request {
	return createRequest(http.MethodPost, relativeURI, subjectIDToToken(subject), body)
}

func createRequest(method string, relativeURI string, authToken string, body string) *http.Request {
	req, err := http.NewRequest(method, fmt.Sprintf("http://localhost:8081%s", relativeURI), strings.NewReader(body))
	if err != nil {
		panic(err)
	}

	if body != "" {
		req.Header.Add("Content-Type", "application/json")
	}

	if authToken != "" {
		req.Header.Add("Authorization", authToken)
	}

	return req
}

func setupService() {
	spicedbToken, err := container.NewToken()
	if err != nil {
		panic(err)
	}
	writeTestEnvToYaml(spicedbToken)

	go Run("../")
	err = waitForSuccess(func() *http.Request { //Repeat a check permission request until it succeeds or a timeout is reached
		return post("/v1alpha/check", "system",
			`{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`)
	})

	if err != nil {
		log.Printf("Error waiting for gateway to come online: %s", err)
		os.Exit(1)
	}
}

func writeTestEnvToYaml(token string) {
	var data, err = os.ReadFile("../config.yaml")
	if err != nil {
		fmt.Printf("Error reading config.yaml: %s\n", err)
		os.Exit(1)
	}
	y := make(map[string]interface{})
	err = yaml.Unmarshal(data, &y)
	if err != nil {
		fmt.Printf("Error parsing yaml: %s\n", err)
		os.Exit(1)
	}

	storeKey := y["app"].(map[string]interface{})["store"].(map[string]interface{})
	storeKey["token"] = token
	storeKey["endpoint"] = "localhost:" + container.Port()
	res, err := yaml.Marshal(y)
	if err != nil {
		fmt.Printf("Error marshalling yaml in test: %s\n", err)
		os.Exit(1)
	}

	e := os.WriteFile("../config.yaml", res, 0644)
	if e != nil {
		fmt.Printf("Error writing new yaml in test: %s\n", err)
		os.Exit(1)
	}
}

func subjectIDToToken(subject string) string {
	if subject == "" {
		return ""
	}

	data, err := jwt.NewBuilder().
		Issuer(testIssuer).
		IssuedAt(time.Now()).
		Audience([]string{testAudience}).
		Subject(subject).
		Claim("scope", testRequiredScope).
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

func teardownService() {
	Stop()
}

func waitForSuccess(reqFactory func() *http.Request) error {
	watch := stopwatch.Start()
	defer func(w stopwatch.Watch) {
		w.Stop()
		log.Printf("Waited %s for gateway to start.", w.Milliseconds())
	}(watch)
	ch := time.After(10 * time.Second)

	for {
		req := reqFactory()
		resp, err := http.DefaultClient.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}

		select {
		case <-ch:
			return err
		case <-time.After(10 * time.Millisecond):
		}
	}
}

var tokenSigningKey jwk.Key
var tokenVerificationKey crypto.PublicKey

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

var oidcDiscoveryURL string

func hostFakeIdp() {
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

	oidcDiscoveryURL = "http://localhost:8180/idp/.well-known/openid-configuration"
	err := http.ListenAndServe("localhost:8180", mux)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	tokenSigningKey, tokenVerificationKey = generateKeys()
	go hostFakeIdp()
	err := waitForSuccess(func() *http.Request {
		req, err := http.NewRequest(http.MethodGet, oidcDiscoveryURL, nil)
		if err != nil {
			panic(err)
		}

		return req
	})

	if err != nil {
		glog.Errorf("Error waiting for fake idp: %s", err)
		os.Exit(1)
	}

	factory := authzed.NewLocalSpiceDbContainerFactory()
	container, err = factory.CreateContainer()

	if err != nil {
		glog.Errorf("Error initializing SpiceDB container: %s", err)
		os.Exit(1)
	}
	result := m.Run()

	container.Close()
	os.Exit(result)
}
