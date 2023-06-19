package userservice

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain"
	"authz/domain/contracts"
	"authz/testenv"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/stretchr/testify/assert"
)

const (
	OrgID         = "123"
	CertDirectory = "../../../testdata/test-certs/"
)

func TestUserServiceSubjectRepository_get_single_page(t *testing.T) {
	//Given
	expectedSubjects := []domain.Subject{{
		SubjectID: "1",
		Enabled:   true,
	}}

	srv := testenv.CreateFakeUserServiceAPI(t, expectedSubjects, map[int]int{}, CertDirectory)
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

	srv := testenv.CreateFakeUserServiceAPI(t, expectedSubjects, map[int]int{}, CertDirectory)
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

	srv := testenv.CreateFakeUserServiceAPI(t, expectedSubjects, map[int]int{}, CertDirectory)
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

	srv := testenv.CreateFakeUserServiceAPI(t, expectedSubjects, map[int]int{}, CertDirectory)
	defer srv.Close()

	repo := createSubjectRepository(srv)

	//When
	subjects, errors := repo.GetByOrgID(OrgID)

	//Then
	assertSuccessfulRequest(t, subjects, errors, expectedSubjects)
}

func TestUserServiceSubjectRepository_error_on_first_request(t *testing.T) {
	expectedSubjects := []domain.Subject{}
	srv := testenv.CreateFakeUserServiceAPI(t, expectedSubjects, map[int]int{0: http.StatusBadRequest}, CertDirectory)
	defer srv.Close()

	repo := createSubjectRepository(srv)

	_, errors := repo.GetByOrgID(OrgID)

	err := <-errors

	assert.Error(t, err)
}

func createSubjectRepository(srv *httptest.Server) contracts.SubjectRepository {
	config := serviceconfig.UserServiceConfig{
		URL:                       fmt.Sprintf("%s/v2/findUsers", srv.URL),
		UserServiceClientCertFile: CertDirectory + "client.crt",
		UserServiceClientKeyFile:  CertDirectory + "client.key",
	}

	cacerts := x509.NewCertPool()
	cacerts.AddCert(srv.Certificate())

	repo, err := NewUserServiceSubjectRepositoryFromConfig(config, cacerts)

	if err != nil {
		panic(err)
	}

	return repo
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

	respJSON := testenv.CreateResponseJSON(expectedSubjects)
	srv := testenv.CreateFakeUserServiceAPI(t, allsubjects, map[int]int{}, CertDirectory)

	defer srv.Close()

	cert, err := tls.LoadX509KeyPair(CertDirectory+"/client.crt", CertDirectory+"client.key")
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
	paging, err := testenv.ExtractPagingParameters([]byte(`{
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

func TestCreateResponseJSON(t *testing.T) {
	json := testenv.CreateResponseJSON([]domain.Subject{{SubjectID: "1", Enabled: false}, {SubjectID: "2", Enabled: true}})
	ja := jsonassert.New(t)

	ja.Assertf(json, `[{"id":"1","status":"disabled"}, {"id":"2","status":"enabled"}]`)
}
