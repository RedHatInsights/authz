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

// CreateFakeUserServiceAPI creates a faked userservice API to call in tests. Add a list of expected subjects the API should return, a list of statusses it should return (default: 200) and the relative path to the certDir from your test.
func CreateFakeUserServiceAPI(t *testing.T, subjects []domain.Subject, explicitStatus map[int]int, certDir string) *httptest.Server {
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

				paging, err := ExtractPagingParameters(requestBody)
				assert.NoError(t, err)

				validateRequestJSON(ja, string(requestBody))

				results := make([]domain.Subject, 0, paging.Take)
				for resultIndex, subjIndex := 0, paging.Skip; resultIndex < paging.Take && subjIndex < len(subjects); resultIndex, subjIndex = resultIndex+1, subjIndex+1 {
					results = append(results, subjects[subjIndex])
				}
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(CreateResponseJSON(results)))
				if err != nil {
					t.Logf("Error sending response: %s", err)
				}
			}
		}

		requestNo++
	}))
	certFile, err := os.ReadFile(certDir + "/client-ca.crt")
	if err != nil {
		glog.Fatalf("Could not find ca cert! err: %v", err)
	}
	srv.TLS = &tls.Config{
		ClientCAs: x509.NewCertPool(),
	}
	srv.TLS.ClientCAs.AppendCertsFromPEM(certFile)
	srv.StartTLS()

	return srv
}

// ExtractPagingParameters unmarshals paging parameters from a request json.
func ExtractPagingParameters(reqBody []byte) (p UserServicePagingParameters, err error) {
	req := UserServiceRequest{}
	err = json.Unmarshal(reqBody, &req)

	p = req.By.PagingParameters
	return
}

func validateRequestJSON(ja *jsonassert.Asserter, json string) {
	ja.Assertf(json, `{
		"by": {
		  "accountId": "123",
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
