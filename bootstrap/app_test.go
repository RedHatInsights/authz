package bootstrap

import (
	"authz/api/grpc"
	"authz/infrastructure/repository/authzed"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	t.SkipNow() //Skip until meta-authz is in place
	regenerateSpiceDbToken()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "bad",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckAccess(t *testing.T) {
	regenerateSpiceDbToken()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	regenerateSpiceDbToken()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestCheckReturnsTrueWhenUserAuthorized(t *testing.T) {
	regenerateSpiceDbToken()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckReturnsFalseWhenUserNotAuthorized(t *testing.T) {
	regenerateSpiceDbToken()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "not_authorized", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)
}

func TestAssignLicenseReturnsSuccess(t *testing.T) {
	regenerateSpiceDbToken()
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
	regenerateSpiceDbToken()
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
	//The user isn't licensed initially, use is denied
	regenerateSpiceDbToken()
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
	regenerateSpiceDbToken()
	//No one is licensed initially, expect a fixed count and none in use
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
	regenerateSpiceDbToken()
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
	regenerateSpiceDbToken()
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
	regenerateSpiceDbToken()
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

var container *authzed.LocalSpiceDbContainer

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

func get(relativeURI string, authToken string) *http.Request {
	return createRequest(http.MethodGet, relativeURI, authToken, "")
}

func post(relativeURI string, authToken string, body string) *http.Request {
	return createRequest(http.MethodPost, relativeURI, authToken, body)
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

func regenerateSpiceDbToken() {
	token, err := container.NewToken()
	if err != nil {
		panic(err)
	}
	singletonSpiceDbRepository.NewConnection("localhost:"+container.Port(), token, true, false)
}

func waitForGateway() error {
	ch := time.After(10 * time.Second)

	for {
		resp, err := http.Get("http://localhost:8081/")

		if err == nil && resp.StatusCode == http.StatusNotFound {
			return nil
		}

		select {
		case <-ch:
			return err
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func initializeGrpcServer() *grpc.Server {
	token, err := container.NewToken()
	if err != nil {
		panic(err)
	}

	grpc, _ := initialize("localhost:"+container.Port(), token, "spicedb", false)

	return grpc
}

func TestMain(m *testing.M) {
	factory := authzed.NewLocalSpiceDbContainerFactory()
	var err error
	container, err = factory.CreateContainer()

	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		os.Exit(-1)
	}

	go Run(fmt.Sprintf("localhost:%s", container.Port()), "initial", "spicedb", false)

	err = waitForGateway()

	if err != nil {
		fmt.Printf("Error waiting for gateway to come online: %s", err)
		os.Exit(-1)
	}

	result := m.Run()

	container.Close()
	os.Exit(result)
}
