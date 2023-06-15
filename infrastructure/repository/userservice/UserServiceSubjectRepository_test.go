package userservice

import (
	"authz/domain"
	"authz/domain/contracts"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

const (
	OrgID = "123"
)

func TestUserServiceSubjectRepository_get_single_page(t *testing.T) {
	//Given
	expectedSubjects := []domain.Subject{{
		SubjectID: "1",
		Enabled:   true,
	}}

	srv := createTestServer(t, expectedSubjects, map[int]int{})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID(OrgID)

	//Then
	assertSuccessfulRequest(t, subjects, errors, expectedSubjects)
}

func TestUserServiceSubjectRepository_get_single_page_exact_pagesize(t *testing.T) {
	//Given
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		}}

	srv := createTestServer(t, expectedSubjects, map[int]int{})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID(OrgID)

	//Then
	assertSuccessfulRequest(t, subjects, errors, expectedSubjects)
}

func TestUserServiceSubjectRepository_get_two_pages_one_item_on_second(t *testing.T) {
	//Given
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
			Enabled:   true,
		},
	}

	srv := createTestServer(t, expectedSubjects, map[int]int{})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID(OrgID)

	//Then
	assertSuccessfulRequest(t, subjects, errors, expectedSubjects)
}

func TestUserServiceSubjectRepository_get_two_full_pages(t *testing.T) {
	//Given
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
			Enabled:   true,
		},
		{
			SubjectID: "4",
			Enabled:   true,
		},
	}

	srv := createTestServer(t, expectedSubjects, map[int]int{})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID(OrgID)

	//Then
	assertSuccessfulRequest(t, subjects, errors, expectedSubjects)
}

func TestUserServiceSubjectRepository_error_on_first_request(t *testing.T) {
	expectedSubjects := []domain.Subject{}
	srv := createTestServer(t, expectedSubjects, map[int]int{0: http.StatusBadRequest})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	_, errors := repo.GetByOrgID(OrgID)

	err := <-errors

	assert.Error(t, err)
}

func createSubjectRepository(srv *httptest.Server) contracts.SubjectRepository {
	serverURL, err := url.Parse(fmt.Sprintf("%s/v2/findUsers", srv.URL)) //This seems like the repository's responsibility?
	if err != nil {
		panic(err)
	}

	cert, err := tls.LoadX509KeyPair("test-certs/client.crt", "test-certs/client.key")
	if err != nil {
		panic(err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      x509.NewCertPool(),      //<--this part (and the RootCAs.AddCert below) is an artifact of the test setup and needs to happen here
			Certificates: []tls.Certificate{cert}, //<--this part is the repository's responsibility
		},
	}
	transport.TLSClientConfig.RootCAs.AddCert(srv.Certificate())

	client := http.Client{
		Transport: transport,
	}

	return NewUserServiceSubjectRepository(*serverURL, client)
}

func TestUserServiceSubjectRepository_temp(t *testing.T) {
	//Given
	expectedSubjects := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		}}

	allsubjects := make([]domain.Subject, 0, len(expectedSubjects)+1)
	allsubjects = append(allsubjects, domain.Subject{
		SubjectID: "0",
		Enabled:   false,
	})
	allsubjects = append(allsubjects, expectedSubjects...)

	reqJSON := `{
		"by": {
		  "accountId": "123",
		  "withPaging": {
			"firstResultIndex" : 1,
			"maxResults": 2,
			"sortBy": "principal",
			"ascending": true
		  }
		},
		"include": {
		  "allOf": [
			"status"
		  ]
		}
	  }`

	respJSON := createResponseJSON(expectedSubjects)
	srv := createTestServer(t, allsubjects, map[int]int{})

	defer srv.Close()

	cert, err := tls.LoadX509KeyPair("test-certs/client.crt", "test-certs/client.key")
	if err != nil {
		panic(err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      x509.NewCertPool(),      //<--this part (and the RootCAs.AddCert below) is an artifact of the test setup and needs to happen here
			Certificates: []tls.Certificate{cert}, //<--this part is the repository's responsibility
		},
	}
	transport.TLSClientConfig.RootCAs.AddCert(srv.Certificate())

	client := http.Client{
		Transport: transport,
	}

	//When
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v2/findUsers", srv.URL), strings.NewReader(reqJSON))
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)

	//Then
	data, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	ja := jsonassert.New(t)
	ja.Assertf(string(data), respJSON)
}

func createTestServer(t *testing.T, subjects []domain.Subject, explicitStatus map[int]int) *httptest.Server {
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

				paging, err := extractPagingParameters(requestBody)
				assert.NoError(t, err)

				validateRequestJSON(ja, string(requestBody))

				results := make([]domain.Subject, 0, paging.Take)
				for resultIndex, subjIndex := 0, paging.Skip; resultIndex < paging.Take && subjIndex < len(subjects); resultIndex, subjIndex = resultIndex+1, subjIndex+1 {
					results = append(results, subjects[subjIndex])
				}
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(createResponseJSON(results)))
				if err != nil {
					t.Logf("Error sending response: %s", err)
				}
			}
		}

		requestNo++
	}))

	certFile, err := os.ReadFile("test-certs/client-ca.crt")
	if err != nil {
		panic(err)
	}
	srv.TLS = &tls.Config{
		ClientCAs: x509.NewCertPool(),
	}
	srv.TLS.ClientCAs.AppendCertsFromPEM(certFile)
	srv.StartTLS()

	return srv
}

func assertSuccessfulRequest(t *testing.T, subjects chan domain.Subject, errors chan error, expectedSubjects []domain.Subject) bool {
	actualSubjects := make([]domain.Subject, 0, len(expectedSubjects))
loop:
	for {
		select {
		case sub, open := <-subjects:
			if !open {
				break loop
			}
			actualSubjects = append(actualSubjects, sub)
		case err, open := <-errors:
			if !open {
				break loop
			}
			assert.NoError(t, err)
		}
	}

	return assert.EqualValues(t, expectedSubjects, actualSubjects)
}

func TestValidateRequestJSON(t *testing.T) {
	ja := jsonassert.New(t)

	validateRequestJSON(ja, `{
		"by": {
		  "accountId": "123",
		  "withPaging": {
			"firstResultIndex" : 0,
			"maxResults": 5,
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

func TestExtractPagingParameters(t *testing.T) {
	paging, err := extractPagingParameters([]byte(`{
		"by": {
		  "accountId": "123",
		  "withPaging": {
			"firstResultIndex" : 1,
			"maxResults": 2,
			"sortBy": "principal",
			"ascending": true
		  }
		},
		"include": {
		  "allOf": [
			"status"
		  ]
		}
	}`))

	assert.NoError(t, err)
	assert.Equal(t, 1, paging.Skip)
	assert.Equal(t, 2, paging.Take)
}

func extractPagingParameters(reqBody []byte) (p pagingParameters, err error) {
	req := request{}
	err = json.Unmarshal(reqBody, &req)

	p = req.By.PagingParameters
	return
}

type pagingParameters struct {
	Skip int `json:"firstResultIndex"`
	Take int `json:"maxResults"`
}

type request struct {
	By struct {
		PagingParameters pagingParameters `json:"withPaging"`
	} `json:"by"`
}

func TestCreateResponseJSON(t *testing.T) {
	json := createResponseJSON([]domain.Subject{{SubjectID: "1", Enabled: false}, {SubjectID: "2", Enabled: true}})
	ja := jsonassert.New(t)

	ja.Assertf(json, `[{"id":"1","status":"disabled"}, {"id":"2","status":"enabled"}]`)
}

func createResponseJSON(subjects []domain.Subject) string {
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
