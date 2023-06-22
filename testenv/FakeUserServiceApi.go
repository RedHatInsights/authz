//go:build !release

// Package testenv contains helpers for integration tests, such as a Fake UserService API, ...
package testenv

import (
	"authz/domain"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/glog"
	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

// UserServicePagingParameters Struct for Paging Parameters on the userservice call. Part of the request.
type UserServicePagingParameters struct {
	Skip int `json:"firstResultIndex"`
	Take int `json:"maxResults"`
}

// UserServiceRequest - struct for the request to userservice
type UserServiceRequest struct {
	By struct {
		PagingParameters UserServicePagingParameters `json:"withPaging"`
	} `json:"by"`
}

type userServiceUserDataRequest struct {
	By struct {
		UserIds []string `json:"userIds"`
	} `json:"by"`
	Include struct {
		AllOf []string `json:"allOf"`
	} `json:"include"`
}

type userServiceUserDataResponse []struct {
	ID                  string `json:"id"`
	PersonalInformation struct {
		FirstName string `json:"firstName"`
		LastNames string `json:"lastNames"`
	} `json:"personalInformation"`
}

// FakeUserServiceAPI Struct to use in tests.
type FakeUserServiceAPI struct {
	Server *httptest.Server
	// URI of the api
	URI string
	// ServerRootCa Path for optional rootCa to add for the test
	ServerRootCa string
	// CertFile Path for mTLS crt file
	CertFile string
	// CertKey path for mTLS key file
	CertKey string
}

// HostFakeUserServiceAPI creates a faked userservice API to call in tests. Add a list of expected subjects the API should return, a list of statusses it should return (default: 200) and the relative path to the certDir from your test.
func HostFakeUserServiceAPI(t *testing.T, subjects []domain.Subject, org string, explicitStatus map[int]int, certDir string) *FakeUserServiceAPI {
	ja := jsonassert.New(t)

	requestNo := 0
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/findUsers":
			assert.Equal(t, http.MethodPost, r.Method)

			requestBody, err := io.ReadAll(r.Body)
			if assert.NoError(t, err) {
				if explicitStatus[requestNo] > 0 {
					w.WriteHeader(explicitStatus[requestNo])
					return
				}

				// Two Requests use the same URL path: one is for user data enrichment and get subjects/users for a given org
				// Inorder to determine which request is that, the below check is the least simple way of determining what request is that
				// accountID is present in get users for a given orgID, and not present in the request body of user data enrichment
				isGetByOrg := strings.Contains(strings.ToLower(string(requestBody)), "accountid")

				if isGetByOrg {
					paging, err := ExtractPagingParameters(requestBody)
					assert.NoError(t, err)

					validateRequestJSON(ja, string(requestBody), org)

					results := make([]domain.Subject, 0, paging.Take)
					for resultIndex, subjIndex := 0, paging.Skip; resultIndex < paging.Take && subjIndex < len(subjects); resultIndex, subjIndex = resultIndex+1, subjIndex+1 {
						results = append(results, subjects[subjIndex])
					}
					w.WriteHeader(http.StatusOK)
					_, err = w.Write([]byte(CreateResponseJSON(results)))
					if err != nil {
						t.Logf("Error sending response: %s", err)
					}
				} else {
					req := userServiceUserDataRequest{}
					err := json.Unmarshal(requestBody, &req)

					if err != nil {
						t.Logf("Error unmarshalling request: %v", err)
					}

					resp := make(userServiceUserDataResponse, 0)
					for _, uid := range req.By.UserIds {
						// find subjectID
						for _, subject := range subjects {
							if string(subject.SubjectID) == uid {
								principal := struct {
									ID                  string `json:"id"`
									PersonalInformation struct {
										FirstName string `json:"firstName"`
										LastNames string `json:"lastNames"`
									} `json:"personalInformation"`
								}{
									ID: uid,
									PersonalInformation: struct {
										FirstName string `json:"firstName"`
										LastNames string `json:"lastNames"`
									}{
										FirstName: "User",
										LastNames: uid,
									},
								}
								resp = append(resp, principal)

								break
							}
						}

					}
					w.WriteHeader(http.StatusOK)

					bytes, err := json.Marshal(resp)
					if err != nil {
						t.Logf("Error marshalling response: %s", err)
					}

					_, err = w.Write(bytes)
					if err != nil {
						t.Logf("Error sending response: %s", err)
					}
				}

			}
		}

		requestNo++
	}))
	clientRootCa, err := os.ReadFile(certDir + "/client-ca.crt")
	if err != nil {
		glog.Fatalf("Could not load client ca cert! err: %v", err)
	}

	serverCert := fmt.Sprintf("%sserver.crt", certDir)
	serverKey := fmt.Sprintf("%sserver.key", certDir)
	serverPair, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		glog.Fatalf("Failed to load server TLS cert! err: %v", err)
	}

	srv.TLS = &tls.Config{
		ClientCAs:    x509.NewCertPool(),
		Certificates: []tls.Certificate{serverPair},
	}

	srv.TLS.ClientCAs.AppendCertsFromPEM(clientRootCa)

	srv.StartTLS()

	result := &FakeUserServiceAPI{
		Server:       srv,
		URI:          fmt.Sprintf("%s/v2/findUsers", srv.URL),
		CertFile:     fmt.Sprintf("%sclient.crt", certDir),
		CertKey:      fmt.Sprintf("%sclient.key", certDir),
		ServerRootCa: fmt.Sprintf("%sserver-ca.crt", certDir),
	}
	return result
}

// ExtractPagingParameters unmarshals paging parameters from a request json.
func ExtractPagingParameters(reqBody []byte) (p UserServicePagingParameters, err error) {
	req := UserServiceRequest{}
	err = json.Unmarshal(reqBody, &req)

	p = req.By.PagingParameters
	return
}

func validateRequestJSON(ja *jsonassert.Asserter, json string, org string) {
	ja.Assertf(json, `{
		"by": {
		  "accountId": "`+org+`",
		  "withPaging": {
			"firstResultIndex" : "<<PRESENCE>>",
			"maxResults": "<<PRESENCE>>",
			"sortBy": "principal",
			"ascending": true
		  }
		},
		"include": {
		  "allOf": [
			"status"
		  ]
		}
	  }`)
}

// CreateResponseJSON creates expected response json for a given list of subjects
func CreateResponseJSON(subjects []domain.Subject) string {
	/*
		Example response:
		[{"id":"1","status":"disabled"}, {"id":"2","status":"enabled"}]
	*/

	var status string
	var s strings.Builder
	lastIndex := len(subjects) - 1

	s.WriteString("[")
	for i, subject := range subjects {
		if subject.Enabled {
			status = "enabled"
		} else {
			status = "disabled"
		}

		s.WriteString(fmt.Sprintf(`{"id":"%s", "status":"%s"}`, subject.SubjectID, status))
		if i < lastIndex {
			s.WriteString(", ")
		}
	}
	s.WriteString("]")

	return s.String()
}
