package bootstrap

import (
	"authz/domain"
	"authz/infrastructure/repository/authzed"
	"authz/testenv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bradhe/stopwatch"
	"github.com/golang/glog"
	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

const (
	CertDirectory = "../testdata/test-certs/"
)

// container to get the port when re-initializing the service
var container *authzed.LocalSpiceDbContainer

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	t.SkipNow() //Skip until Check meta-authz is in place
	setupService(nil)
	defer teardownService()

	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "bad", "no_particular_org", false, `{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestCheckAccess(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system", "no_particular_org", false, `{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckErrorsWhenTokenMissing(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "", "", false, `{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestCheckReturnsTrueWhenUserAuthorized(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system", "no_particular_org", false, `{"subject": "u1", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestCheckReturnsFalseWhenUserNotAuthorized(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system", "no_particular_org", false, `{"subject": "not_authorized", "operation": "access", "resourcetype": "license", "resourceid": "o1/smarts"}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)
}

func TestAssignSeatReturnsSuccess(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "system", "o1", true, `{
			"assign": [
			  "u7"
			]
		  }`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestAssignSeatReturnsFailureWhenOrgIsUnauthorized(t *testing.T) {
	setupService(nil)
	defer teardownService()
	// OrgID in request path and in the token are different
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "system", "o2", true, `{
			"assign": [
			  "u7"
			]
		  }`))

	assert.NoError(t, err)

	assert.Equal(t, 403, resp.StatusCode)
}

func TestAssignSeatReturnsFailureWhenNotAuthorizedOrgAdmin(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "system", "o1", false, `{
			"assign": [
			  "u7"
			]
		  }`))

	assert.NoError(t, err)

	assert.Equal(t, 403, resp.StatusCode)
}

func TestUnassignLicenseReturnsSuccess(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "system", "o1", true, `{
			"unassign": [
			  "u1"
			]
		}`))

	assert.NoError(t, err)

	assertJSONResponse(t, resp, 200, `{}`)
}

func TestEntitleOrgSucceedsWithNewOrgAndNewServiceLicense(t *testing.T) {
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		},
		{
			SubjectID: "3",
			Enabled:   false,
		},
	}

	expectedOrg := "o3"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()

	_, err := http.DefaultClient.Do(post("/v1alpha/orgs/o3/entitlements/foobar", "system", "o3", true, `{
			"maxSeats": 25
		}`))

	assert.NoError(t, err)

	resp2, err := http.DefaultClient.Do(get("/v1alpha/orgs/o3/licenses/foobar", "system", "o3", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp2, 200, `{"seatsAvailable":25, "seatsTotal": 25}`)

	container.WaitForQuantizationInterval()

	//round trip: check users were imported and are assignable.
	resp3, err := http.DefaultClient.Do(get("/v1alpha/orgs/o3/licenses/foobar/seats?filter=assignable", "system", "o3", true))
	assert.NoError(t, err)
	// 3rd one is disabled, so remove from expected.
	assertJSONResponse(t, resp3, 200, `{"users": ["<<UNORDERED>>", {"assigned":false,"displayName":"User 1","id":"1"}, {"displayName":"User 2","id":"2","assigned":false}]}`)
}

func TestEntitleOrgFailsWithUnAuthorizedRequestor(t *testing.T) {
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		},
		{
			SubjectID: "3",
			Enabled:   false,
		},
	}

	expectedOrg := "o3"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()

	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o3/entitlements/foobar", "unAuthorizedSubject", "o3", true, `{
			"maxSeats": 25
		}`))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

}
func TestImportOrgImportsUsersForNewOrg(t *testing.T) {
	//given
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		},
		{
			SubjectID: "3",
			Enabled:   false,
		},
	}
	expectedOrg := "newOrg"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()
	//when

	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/"+expectedOrg+"/import", "system", "newOrg", true, ""))
	//then
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200,
		`{"importedUsersCount":"3", "notImportedUsersCount":"0"}`)
}

func TestImportOrgFailsWithUnAuthorizedRequestor(t *testing.T) {
	//given
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		},
		{
			SubjectID: "3",
			Enabled:   false,
		},
	}
	expectedOrg := "newOrg"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()
	//when

	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/"+expectedOrg+"/import", "unAuthorizedSubject", "newOrg", true, ""))
	//then
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestImportOrgImportsUsersForExistingOrg(t *testing.T) {
	//given
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "u2", //u2 already exists from spicedb relations, skip it.
			Enabled:   true,
		},
		{
			SubjectID: "3",
			Enabled:   false,
		},
	}
	expectedOrg := "o2"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()
	//when
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/"+expectedOrg+"/import", "system", "o2", true, ""))
	//then
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200,
		`{"importedUsersCount":"2", "notImportedUsersCount":"1"}`)
}

func TestImportOrgImportsNothingWhenNoUsersAreThere(t *testing.T) {
	var expectedSubjects []domain.Subject
	expectedOrg := "o3"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()
	//when
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/"+expectedOrg+"/import", "system", "o3", true, ""))
	//then
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200,
		`{"importedUsersCount":"0", "notImportedUsersCount":"0"}`)
}

func TestImportOrgReturnsErrorWhenUserServiceReturnsError(t *testing.T) {
	var expectedSubjects []domain.Subject
	expectedOrg := "o2"
	notExistingOrg := "fooOrg"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{0: http.StatusBadRequest}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()
	//when
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/"+notExistingOrg+"/import", "system", "o2", true, ""))
	//then
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestEntitleOrgSucceedstWithExistingOrgAndNewLicenses(t *testing.T) {
	setupService(nil)
	defer teardownService()
	_, err := http.DefaultClient.Do(post("/v1alpha/orgs/o2/entitlements/foobar", "system", "o2", true, `{
			"maxSeats": 25
		}`))

	assert.NoError(t, err)
	_, err = http.DefaultClient.Do(get("/v1alpha/orgs/o2/licenses/foobar", "system", "o2", true))
	assert.NoError(t, err)
	_, err = http.DefaultClient.Do(post("/v1alpha/orgs/o2/entitlements/bazbar", "system", "o2", true, `{
			"maxSeats": 20
		}`))

	assert.NoError(t, err)
	resp2, err := http.DefaultClient.Do(get("/v1alpha/orgs/o2/licenses/bazbar", "system", "o2", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp2, 200, `{"seatsAvailable":20, "seatsTotal": 20}`)
}

func TestEntitleOrgTriggersUserImportWhenOrgExistsButHasNoUsersImportedYet(t *testing.T) {
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		},
		{
			SubjectID: "3",
			Enabled:   false,
		},
	}

	expectedOrg := "oNoUsers"
	usSrv := testenv.HostFakeUserServiceAPI(t, expectedSubjects, expectedOrg, map[int]int{}, CertDirectory)
	defer usSrv.Server.Close()

	setupService(usSrv)
	defer teardownService()

	_, err := http.DefaultClient.Do(post("/v1alpha/orgs/"+expectedOrg+"/entitlements/foo", "system", "oNoUsers", true, `{
			"maxSeats": 25
		}`))

	assert.NoError(t, err)

	resp2, err := http.DefaultClient.Do(get("/v1alpha/orgs/"+expectedOrg+"/licenses/foo", "system", "oNoUsers", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp2, 200, `{"seatsAvailable":25, "seatsTotal": 25}`)

	container.WaitForQuantizationInterval()
	//round trip: check users were imported and are assignable.
	resp3, err := http.DefaultClient.Do(get("/v1alpha/orgs/"+expectedOrg+"/licenses/foo/seats?filter=assignable", "system", "oNoUsers", true))
	assert.NoError(t, err)
	// 3rd one is disabled, so remove from expected.
	assertJSONResponse(t, resp3, 200, `{"users": ["<<UNORDERED>>", {"assigned":false,"displayName":"User 1","id":"1"}, {"displayName":"User 2","id":"2","assigned":false}]}`)
}

func TestEntitleOrgTwiceForSameLicenseFailsWithBadRequest(t *testing.T) {
	setupService(nil)
	defer teardownService()

	_, err := http.DefaultClient.Do(post("/v1alpha/orgs/o3/entitlements/foobar", "system", "o3", true, `{
			"maxSeats": 25
		}`))

	assert.NoError(t, err)

	resp2, err := http.DefaultClient.Do(get("/v1alpha/orgs/o3/licenses/foobar", "system", "o3", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp2, 200, `{"seatsAvailable":25, "seatsTotal": 25}`)

	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o3/entitlements/foobar", "system", "o3", true, `{
			"maxSeats": 24
		}`))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	//to make sure the license didn't get messed up we also assert that the seatcount is the expected one
	resp4, err := http.DefaultClient.Do(get("/v1alpha/orgs/o3/licenses/foobar", "system", "o3", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp4, 200, `{"seatsAvailable":25, "seatsTotal": 25}`)
}

func TestEntitleOrgFailsWithNegativeMaxSeatsValue(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/entitlements/wisdom", "system", "o1", true, `{
			"maxSeats": -1
		}`))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // TODO - Check and update the correct http status code: bad request or internal server error
}

func TestEntitleOrgFailsWithEmptyMaxSeatsValue(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/entitlements/wisdom", "system", "o1", true, `{
			"maxSeats":
		}`))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestEntitleOrgFailsWithEmptyBody(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/entitlements/wisdom", "system", "o1", true, ``))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // TODO - Check and update the correct http status code: bad request or internal server error

}

func TestGrantedLicenseAllowsUse(t *testing.T) {
	setupService(nil)
	defer teardownService()
	//The user isn't licensed initially, use is denied
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system", "o1", false, `{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, false)

	//Grant a license
	resp, err = http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "okay", "o1", true, `{
		"assign": [
			"u2"
			]
			}`))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{}`)
	container.WaitForQuantizationInterval()

	//Should be allowed now
	resp, err = http.DefaultClient.Do(post("/v1alpha/check", "system", "o1", false, `{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestGrantedLicenseGivesAccessToServices(t *testing.T) {
	setupService()
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/check", "system",
		`{"subject": "u1", "operation": "access", "resourcetype": "service", "resourceid": "smarts"}`))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"result": %t, "description": ""}`, true)
}

func TestGrantedLicenseAffectsCountsAndDetails(t *testing.T) {
	setupService(nil)
	defer teardownService()

	resp, err := http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts", "system", "o1", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":8, "seatsTotal": 10}`)

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts/seats", "system", "o1", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"users": ["<<UNORDERED>>", {"assigned":true,"displayName":"O1 User 1","id":"u1"}, {"displayName":"O1 User 3","id":"u3","assigned":true}]}`)

	//Grant a license
	resp, err = http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "okay", "o1", true, `{
			"assign": [
			  "u7"
			]
		  }`))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	container.WaitForQuantizationInterval()

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts", "system", "o1", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"seatsAvailable":7, "seatsTotal": 10}`)

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts/seats", "system", "o1", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"users": "<<PRESENCE>>"}`)

	resp, err = http.DefaultClient.Do(get("/v1alpha/orgs/o1/licenses/smarts/seats?filter=assignable", "token", "o1", true))
	assert.NoError(t, err)
	assertJSONResponse(t, resp, 200, `{"users":"<<PRESENCE>>"}`)
}

func TestOverAssigningLicensesFails(t *testing.T) {
	setupService(nil)
	defer teardownService()
	resp, err := http.DefaultClient.Do(post("/v1alpha/orgs/o1/licenses/smarts", "okay", "o1", true, `{
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
	setupService(nil)
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
	setupService(nil)
	defer teardownService()
	body := `{
			"assign": [
			  "u7"
			]
		  }`

	req := post("/v1alpha/orgs/o1/licenses/smarts", "okay", "o1", true, body)
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

func get(relativeURI string, subject string, orgID string, isOrgAdmin bool) *http.Request {
	return createRequest(http.MethodGet, relativeURI, testenv.CreateToken(subject, orgID, isOrgAdmin), "")
}

func post(relativeURI string, subject string, orgID string, isOrgAdmin bool, body string) *http.Request {
	return createRequest(http.MethodPost, relativeURI, testenv.CreateToken(subject, orgID, isOrgAdmin), body)
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

var temporaryConfigFile *os.File
var temporarySecretDirectory string

func setupService(fakeUserService *testenv.FakeUserServiceAPI) {
	spicedbToken, err := container.NewToken()
	if err != nil {
		panic(err)
	}
	temporaryConfigFile, err = os.CreateTemp("", "authz-test-config-*.yaml")
	if err != nil {
		panic(err)
	}
	writeTestEnvToYaml(spicedbToken, fakeUserService)

	go Run(temporaryConfigFile.Name())
	err = waitForSuccess(func() *http.Request { //Repeat a check permission request until it succeeds or a timeout is reached
		return post("/v1alpha/check", "system", "o3", true, `{"subject": "u2", "operation": "assigned", "resourcetype": "license_seats", "resourceid": "o1/smarts"}`)
	})

	if err != nil {
		log.Printf("Error waiting for gateway to come online: %s", err)
		os.Exit(1)
	}
}

func writeTestEnvToYaml(token string, userService *testenv.FakeUserServiceAPI) {
	var data, err = os.ReadFile("../config.yaml")
	if err != nil {
		fmt.Printf("Error reading config.yaml: %s\n", err)
		os.Exit(1)
	}
	yml := make(map[string]interface{})
	err = yaml.Unmarshal(data, &yml)
	if err != nil {
		fmt.Printf("Error parsing yaml: %s\n", err)
		os.Exit(1)
	}

	tempSecretFile, err := os.CreateTemp(temporarySecretDirectory, "")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(tempSecretFile.Name(), []byte(token), 0666)
	if err != nil {
		panic(err)
	}

	storeKey := yml["store"].(map[string]interface{})
	storeKey["tokenFile"] = tempSecretFile.Name()
	storeKey["endpoint"] = "localhost:" + container.Port()

	authKey := yml["auth"].([]interface{})[0].(map[string]interface{})

	if storeKey["kind"] == "stub" {
		log.Printf("Enabling spicedb store for tests.")
		storeKey["kind"] = "spicedb"
	}

	if authKey["enabled"] == false {
		log.Printf("Enabling authn middleware for tests.")
		authKey["enabled"] = true
	}

	// Ensure the local discovery endpoint is used instead of any specified in the config
	authKey["discoveryEndpoint"] = "http://localhost:8180/idp/.well-known/openid-configuration"

	// Override any config that is there for AuthZ - allow list with test specific data
	yml["authz"] = make(map[string]interface{}, 0)
	authzKey := yml["authz"].(map[string]interface{})
	authzKey["licenseImportWhitelist"] = []string{"system"}

	if userService != nil {
		userServiceKey := yml["userservice"].(map[string]interface{})
		userServiceKey["url"] = userService.URI
		userServiceKey["userServiceClientCertFile"] = userService.CertFile
		userServiceKey["userServiceClientKeyFile"] = userService.CertKey
		userServiceKey["optionalRootCA"] = userService.ServerRootCa
	}

	res, err := yaml.Marshal(yml)
	if err != nil {
		fmt.Printf("Error marshalling yaml in test: %s\n", err)
		os.Exit(1)
	}

	_, e := temporaryConfigFile.Write(res)
	if e != nil {
		fmt.Printf("Error writing new yaml in test: %s\n", err)
		os.Exit(1)
	}
}

func teardownService() {
	Stop()
	err := os.Remove(temporaryConfigFile.Name())

	if err != nil && !os.IsNotExist(err) {
		glog.Errorf("Error deleting temporary config file %s from temp directory: %v", temporaryConfigFile.Name(), err)
	}

	temporaryConfigFile = nil
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

func TestMain(m *testing.M) {
	go testenv.HostFakeIdp()
	err := waitForSuccess(func() *http.Request {
		req, err := http.NewRequest(http.MethodGet, testenv.OidcDiscoveryURL, nil)
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

	temporarySecretDirectory, err = os.MkdirTemp(os.TempDir(), ".secrets")
	if err != nil {
		glog.Error("Error setting up secret directory: ", err)
		os.Exit(1)
	}

	result := m.Run()

	container.Close()

	err = os.RemoveAll(temporarySecretDirectory)
	if err != nil {
		panic(err)
	}

	os.Exit(result)
}
