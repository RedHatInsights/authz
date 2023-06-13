package userservice

import (
	"authz/domain"
	"authz/domain/contracts"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

func TestUserServiceSubjectRepository_get_single_page(t *testing.T) {
	//Given
	expectedSubjects := []domain.Subject{{
		SubjectID: "1",
		Enabled:   true,
	}}

	srv := createTestServer(t, []requestAndResponse{
		{
			RequestJSON:  createRequestJSON("123", 0, 2),
			ResponseJSON: createResponseJSON(expectedSubjects),
		},
	})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID("123")

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

	srv := createTestServer(t, []requestAndResponse{
		{
			RequestJSON:  createRequestJSON("123", 0, 2),
			ResponseJSON: createResponseJSON(expectedSubjects),
		},
	})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID("123")

	//Then
	assertSuccessfulRequest(t, subjects, errors, expectedSubjects)
}

func TestUserServiceSubjectRepository_get_two_pages_one_item_on_second(t *testing.T) {
	//Given
	expectedSubjects1 := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		},
	}
	expectedSubjects2 := []domain.Subject{
		{
			SubjectID: "3",
			Enabled:   true,
		},
	}

	srv := createTestServer(t, []requestAndResponse{
		{
			RequestJSON:  createRequestJSON("123", 0, 2),
			ResponseJSON: createResponseJSON(expectedSubjects1),
		},
		{
			RequestJSON:  createRequestJSON("123", 2, 2),
			ResponseJSON: createResponseJSON(expectedSubjects2),
		},
	})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID("123")

	//Then
	assertSuccessfulRequest(t, subjects, errors, append(expectedSubjects1, expectedSubjects2...))
}

func TestUserServiceSubjectRepository_get_two_full_pages(t *testing.T) {
	//Given
	expectedSubjects1 := []domain.Subject{
		{
			SubjectID: "1",
			Enabled:   true,
		},
		{
			SubjectID: "2",
			Enabled:   true,
		},
	}

	expectedSubjects2 := []domain.Subject{
		{
			SubjectID: "3",
			Enabled:   true,
		},
		{
			SubjectID: "4",
			Enabled:   true,
		},
	}

	srv := createTestServer(t, []requestAndResponse{
		{
			RequestJSON:  createRequestJSON("123", 0, 2),
			ResponseJSON: createResponseJSON(expectedSubjects1),
		},
		{
			RequestJSON:  createRequestJSON("123", 2, 2),
			ResponseJSON: createResponseJSON(expectedSubjects2),
		},
	})
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID("123")

	//Then
	assertSuccessfulRequest(t, subjects, errors, append(expectedSubjects1, expectedSubjects2...))
}

type requestAndResponse struct {
	RequestJSON  string
	ResponseJSON string
}

func createSubjectRepository(srv *httptest.Server) contracts.SubjectRepository {
	return nil
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

	reqJSON := createRequestJSON("123", 0, 2)
	respJSON := createResponseJSON(expectedSubjects)
	srv := createTestServer(t, []requestAndResponse{
		{
			RequestJSON:  reqJSON,
			ResponseJSON: respJSON,
		},
	})
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

func createTestServer(t *testing.T, setups []requestAndResponse) *httptest.Server {
	ja := jsonassert.New(t)
	i := 0

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/findUsers":
			if !assert.Less(t, i, len(setups)) {
				return
			}

			assert.Equal(t, http.MethodPost, r.Method)

			requestBody, err := io.ReadAll(r.Body)
			if assert.NoError(t, err) {
				ja.Assertf(string(requestBody), setups[i].RequestJSON)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(setups[i].ResponseJSON))
			}

			i++
		}
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

func TestCreateRequestJSON(t *testing.T) {
	json := createRequestJSON("123", 0, 5)
	ja := jsonassert.New(t)

	ja.Assertf(json, `{
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

func createRequestJSON(orgID string, firstElementIndex int, pageSize int) string {
	return fmt.Sprintf(`{
		"by": {
		  "accountId": "%s",
		  "withPaging": {
			"firstResultIndex" : %d,
			"maxResults": %d,
			"sortBy": "principal",
			"ascending": true
		  }
		},
		"include": {
		  "allOf": [
			"status"
		  ]
		}
	  }`, orgID, firstElementIndex, pageSize)
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
