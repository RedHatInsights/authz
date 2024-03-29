// Package userservice is for the userservice repository and related components
package userservice

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain"
	"authz/domain/contracts"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/golang/glog"
)

const (
	sortBy           = "principal"
	defaultPageSize  = 20
	defaultSortOrder = true

	assumeNextPageAvailableByDefaultIfError = true // when retrieving a page of users and there is an error, should we still assume another page exists
)

// SubjectRepository defines a repository that queries a user service using json requests of the type defined in userServiceSubjectByOrgRequest
type SubjectRepository struct {
	URL        url.URL
	HTTPClient http.Client
	Paging     struct {
		PageSize  int
		SortOrder bool
	}
}

// NewUserServiceSubjectRepositoryFromConfig creates a new UserServiceRepository instance from a config struct and certpool
func NewUserServiceSubjectRepositoryFromConfig(config serviceconfig.UserServiceConfig, cacerts *x509.CertPool) (contracts.SubjectRepository, error) {
	url, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(config.UserServiceClientCertFile, config.UserServiceClientKeyFile)
	if err != nil {
		return nil, err
	}

	if len(config.OptionalRootCA) > 0 {
		glog.Infof("Adding optional root CA: %s", config.OptionalRootCA)
		rootCa, err := os.ReadFile(config.OptionalRootCA)
		if err != nil {
			return nil, err
		}

		ok := cacerts.AppendCertsFromPEM(rootCa)
		if !ok {
			glog.Errorf("Error adding optional ca cert. Could not append certificate to cert pool.")
		}
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.DisableCAVerification,
				RootCAs:            cacerts,
				Certificates:       []tls.Certificate{cert},
			},
		},
	}

	return NewUserServiceSubjectRepository(*url, client), nil
}

// NewUserServiceSubjectRepository creates a new UserServiceSubjectRepository
func NewUserServiceSubjectRepository(url url.URL, client http.Client) *SubjectRepository {

	return &SubjectRepository{
		URL:        url,
		HTTPClient: client,
		Paging: struct {
			PageSize  int
			SortOrder bool
		}{PageSize: defaultPageSize, SortOrder: defaultSortOrder},
	}
}

type userServiceSubjectByOrgRequest struct {
	By struct {
		AccountID  string `json:"accountId"`
		WithPaging struct {
			FirstResultIndex int    `json:"firstResultIndex"`
			MaxResults       int    `json:"maxResults"`
			SortBy           string `json:"sortBy"`
			Ascending        bool   `json:"ascending"`
		} `json:"withPaging"`
	} `json:"by"`
	Include struct {
		AllOf []string `json:"allOf"`
	} `json:"include"`
}

type userServiceSubjectByOrgResponse []struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type userServiceUserDataRequest struct {
	By struct {
		UserIds []string `json:"userIds"`
	} `json:"by"`
	Include struct {
		AllOf []string `json:"allOf"`
	} `json:"include"`
}

type authenticationData struct {
	Principal    string `json:"principal"`
	ProviderName string `json:"providerName"`
}

type userServiceUserDataResponse []struct {
	ID                  string               `json:"id"`
	Authentications     []authenticationData `json:"authentications"`
	PersonalInformation struct {
		FirstName   string `json:"firstName"`
		MiddleNames string `json:"middleNames"`
		LastNames   string `json:"lastNames"`
		Prefix      string `json:"prefix"`
	} `json:"personalInformation"`
	Status string `json:"status"`
}

// GetByOrgID retrieves all members of the given organization
func (u *SubjectRepository) GetByOrgID(orgID string) (chan domain.Subject, chan error) {
	subChan := make(chan domain.Subject)
	errChan := make(chan error)

	if !u.validateConfigAndOrg(orgID) {
		errChan <- fmt.Errorf("UserServiceSubjectRepository config was not valid: %v", u)
		close(subChan)
		close(errChan)

		return subChan, errChan
	}

	go func() {
		defer func() {
			close(subChan)
			close(errChan)
		}()

		// Users are requested from the UserService in "pages"
		shouldFetchPage := true

		for page := 0; shouldFetchPage; page++ {
			nextPageIsAvailable, serviceCallErr, pageProcessingErr := u.fetchPageOfUsers(orgID, page, subChan, errChan)

			shouldFetchPage = shouldFetchNextPage(nextPageIsAvailable, serviceCallErr, pageProcessingErr)

			if nextPageIsAvailable && !shouldFetchPage {
				errChan <- fmt.Errorf("GetByOrgID has stopped trying to retrieve more users due to errors, but there may be more")
			}
		}
	}()

	return subChan, errChan
}

// GetByID retrieves a principal for the given ID. If no ID is provided (ex: empty string), it returns an anonymous principal. If any error occurs, it's returned.
func (u *SubjectRepository) GetByID(id domain.SubjectID) (principal domain.Principal, err error) {
	principals, err := u.GetByIDs([]domain.SubjectID{id})

	if err == nil && len(principals) > 0 {
		principal = principals[0]
	}

	return
}

// GetByIDs is a bulk version of GetByID to allow the underlying implementation to optimize access to sets of principals and should otherwise have the same behavior.
func (u *SubjectRepository) GetByIDs(ids []domain.SubjectID) (principals []domain.Principal, err error) {
	req := u.makeUserServiceUserDataRequest(ids)

	resp, err := u.doUserServiceUserDataCall(req)

	if err != nil {
		return
	}
	for _, userData := range resp {
		var principal domain.Principal
		principal.ID = domain.SubjectID(userData.ID)
		// TODO - For now display name is constructed as firstname  Lastname - Check with the proper API spec and revisit it
		principal.FirstName = userData.PersonalInformation.FirstName
		principal.LastName = userData.PersonalInformation.LastNames
		principal.Username = getUsername(userData.Authentications)
		principals = append(principals, principal)
	}
	return
}

func getUsername(authns []authenticationData) string {
	if len(authns) > 0 {
		return authns[0].Principal //Look for specific provider?
	}

	return ""
}

func (u *SubjectRepository) validateConfigAndOrg(_ string) bool {
	// TODO: add more validations

	return u.Paging.PageSize > 0
}

func (u *SubjectRepository) makeUserServiceSubjectByOrgRequest(orgID string, resultIndex int) userServiceSubjectByOrgRequest {
	req := userServiceSubjectByOrgRequest{}
	req.By.AccountID = orgID
	req.By.WithPaging.FirstResultIndex = resultIndex
	req.By.WithPaging.MaxResults = u.Paging.PageSize
	req.By.WithPaging.SortBy = sortBy
	req.By.WithPaging.Ascending = u.Paging.SortOrder
	req.Include.AllOf = []string{"status"}

	return req
}

func (u *SubjectRepository) makeUserServiceUserDataRequest(subjectIDs []domain.SubjectID) userServiceUserDataRequest {

	var reqIds []string
	for _, id := range subjectIDs {
		reqIds = append(reqIds, string(id))
	}

	req := userServiceUserDataRequest{}
	req.By.UserIds = reqIds
	req.Include.AllOf = []string{"status", "personal_information", "authentications"}

	return req
}

func (u *SubjectRepository) fetchPageOfUsers(orgID string, currentPage int, subChan chan domain.Subject, errChan chan error) (bool, error, error) {
	req := u.makeUserServiceSubjectByOrgRequest(orgID, currentPage*u.Paging.PageSize)

	resp, nextPageAvailable, serviceCallErr := u.doPagedUserServiceCall(req, errChan)

	var pageProcessingErr error
	if resp != nil {
		pageProcessingErr = processUsersResponsePage(resp, subChan, errChan)
	}

	return nextPageAvailable, serviceCallErr, pageProcessingErr
}

func (u *SubjectRepository) doPagedUserServiceCall(req userServiceSubjectByOrgRequest, errChan chan error) (userServiceSubjectByOrgResponse, bool, error) {
	// Step 1: marshall the userServiceSubjectByOrgRequest
	userRepositoryRequestJSON, err := json.Marshal(req)

	if err != nil {
		err = fmt.Errorf("error marshalling userServiceSubjectByOrgRequest: %v: %w", req, err)
		errChan <- err
		return nil, assumeNextPageAvailableByDefaultIfError, err
	}

	// Step 2: POST the request using the configured repository http client and url
	body, err := u.doUserServiceCall(userRepositoryRequestJSON, errChan, true)
	if err != nil {
		return nil, assumeNextPageAvailableByDefaultIfError, err
	}

	// Step 3: unmarshall the userServiceSubjectByOrgResponse, which is a slice of subjects
	var userResponses userServiceSubjectByOrgResponse
	err = json.Unmarshal(body, &userResponses)

	if err != nil {
		err = fmt.Errorf("failed to unmarshall userServiceSubjectByOrgResponse from body: %v, %w", string(body), err)
		errChan <- err
	}

	// Step 4: try to determine if there is another page that can be requested
	var nextPageAvailable bool
	if userResponses != nil {
		nextPageAvailable = req.By.WithPaging.MaxResults == len(userResponses) // that was a full page, so we know there's another page
	} else {
		nextPageAvailable = assumeNextPageAvailableByDefaultIfError
	}

	return userResponses, nextPageAvailable, err
}

func (u *SubjectRepository) doUserServiceUserDataCall(req userServiceUserDataRequest) (userServiceUserDataResponse, error) {
	userServiceUserDataRequestJSON, err := json.Marshal(req)
	var userServiceUserDataResponses userServiceUserDataResponse

	if err != nil {
		return nil, fmt.Errorf("error marshalling userServiceUserDataRequest: %v: %w", req, err)
	}
	body, err := u.doUserServiceCall(userServiceUserDataRequestJSON, nil, false)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &userServiceUserDataResponses)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall userServiceUserDataResponse from body: %v, %w", string(body), err)
	}
	return userServiceUserDataResponses, nil
}

func (u *SubjectRepository) doUserServiceCall(reqBody []byte, errChan chan error, useErrChan bool) (respBody []byte, err error) {
	resp, err := u.HTTPClient.Post(u.URL.String(), "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		err = fmt.Errorf("failed to POST to UserService: %v: %w", u.URL, err)
		if useErrChan {
			errChan <- err
		}
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil && useErrChan {
			errChan <- fmt.Errorf("failed to close response body: %v: %w", u.URL, err)
		}
	}()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected http response status code on request to user repository: %v", resp.Status)
		if useErrChan {
			errChan <- err
		}
		return nil, err

	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
		if useErrChan {
			errChan <- err
		}
		return nil, err
	}
	return
}

func processUsersResponsePage(resp userServiceSubjectByOrgResponse, subChan chan domain.Subject, errChan chan error) error {
	for _, user := range resp {
		if user.ID == "" || user.Status == "" {
			err := fmt.Errorf("user ID or user status was empty for importing user %v", user)
			errChan <- err

			if !shouldContinueProcessingUsersPage(err) {
				return err
			}
		}

		var enabled bool
		if strings.EqualFold(user.Status, "enabled") {
			enabled = true
		} else {
			enabled = false
		}

		subject := domain.Subject{
			SubjectID: domain.SubjectID(user.ID),
			Enabled:   enabled,
		}

		subChan <- subject
	}

	return nil
}

func shouldContinueProcessingUsersPage(err error) bool {
	// TODO: Any error causes all processing of this page to cease -- maybe better logic?

	return err != nil
}

func shouldFetchNextPage(anotherPageAvailable bool, serviceCallErr error, pageProcessingErr error) (shouldFetchNext bool) {
	// TODO: Determine whether to keep going assuming there is another page and the error is the "right" type of error

	shouldFetchNext = anotherPageAvailable && serviceCallErr == nil && pageProcessingErr == nil

	return shouldFetchNext
}
