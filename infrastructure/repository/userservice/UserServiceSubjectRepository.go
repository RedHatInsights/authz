// Package userservice is for the userservice repository and related components
package userservice

import (
	"authz/domain"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	sortBy           = "principal"
	defaultPageSize  = 20
	defaultSortOrder = true

	assumeNextPageAvailableByDefaultIfError = true // when retrieving a page of users and there is an error, should we still assume another page exists
)

// UserServiceSubjectRepository defines a repository that queries a user service using json requests of the type defined in userRepositoryRequest
type UserServiceSubjectRepository struct {
	URL        url.URL
	HTTPClient http.Client
	Paging     struct {
		PageSize  int
		SortOrder bool
	}
}

// NewUserServiceSubjectRepository creates a new UserServiceSubjectRepository
func NewUserServiceSubjectRepository(url url.URL, client http.Client) UserServiceSubjectRepository {
	return UserServiceSubjectRepository{
		URL:        url,
		HTTPClient: client,
		Paging: struct {
			PageSize  int
			SortOrder bool
		}{PageSize: defaultPageSize, SortOrder: defaultSortOrder},
	}
}

type userRepositoryRequest struct {
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

type userRepositoryResponse []struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// GetByOrgID retrieves all members of the given organization
func (u *UserServiceSubjectRepository) GetByOrgID(orgID string) (chan domain.Subject, chan error) {
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
				errChan <- fmt.Errorf("GetByOrgID may not have retrieved all subjects due to errors")
			}
		}
	}()

	return subChan, errChan
}

func (u *UserServiceSubjectRepository) validateConfigAndOrg(_ string) bool {
	// TODO: add more validations

	return u.Paging.PageSize > 0
}

func (u *UserServiceSubjectRepository) makeUserRepositoryRequest(orgID string, resultIndex int) userRepositoryRequest {
	req := userRepositoryRequest{}
	req.By.AccountID = orgID
	req.By.WithPaging.FirstResultIndex = resultIndex
	req.By.WithPaging.MaxResults = u.Paging.PageSize
	req.By.WithPaging.SortBy = sortBy
	req.By.WithPaging.Ascending = u.Paging.SortOrder
	req.Include.AllOf = []string{"status"}

	return req
}

func (u *UserServiceSubjectRepository) fetchPageOfUsers(orgID string, currentPage int, subChan chan domain.Subject, errChan chan error) (bool, error, error) {
	req := u.makeUserRepositoryRequest(orgID, currentPage*u.Paging.PageSize)

	resp, nextPageAvailable, serviceCallErr := u.doPagedUserServiceCall(req, errChan)

	var pageProcessingErr error
	if resp != nil {
		pageProcessingErr = processUsersResponsePage(resp, subChan, errChan)
	}

	return nextPageAvailable, serviceCallErr, pageProcessingErr
}

func (u *UserServiceSubjectRepository) doPagedUserServiceCall(req userRepositoryRequest, errChan chan error) (userRepositoryResponse, bool, error) {
	// Step 1: marshall the userRepositoryRequest
	userRepositoryRequestJSON, err := json.Marshal(req)

	if err != nil {
		err = fmt.Errorf("error marshalling userRepositoryRequest: %v: %w", req, err)
		errChan <- err
		return nil, assumeNextPageAvailableByDefaultIfError, err
	}

	// Step 2: POST the request using the configured repository http client and url
	resp, err := u.HTTPClient.Post(u.URL.String(), "application/json", bytes.NewBuffer(userRepositoryRequestJSON))

	if err != nil {
		err = fmt.Errorf("failed to POST to UserService: %v: %w", u.URL, err)
		errChan <- err
		return nil, assumeNextPageAvailableByDefaultIfError, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			errChan <- fmt.Errorf("failed to close response body: %v: %w", u.URL, err)
		}
	}()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected http response status code on request to user repository: %v", resp.Status)
		errChan <- err
		return nil, assumeNextPageAvailableByDefaultIfError, err
	}

	// Step 3: read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
		errChan <- err
		return nil, assumeNextPageAvailableByDefaultIfError, err
	}

	// Step 4: unmarshall the userRepositoryResponse, which is a slice of subjects
	var userResponses userRepositoryResponse
	err = json.Unmarshal(body, &userResponses)

	if err != nil {
		err = fmt.Errorf("failed to unmarshall userRepositoryResponse from body: %v, %w", string(body), err)
		errChan <- err
	}

	// Step 5: try to determine if there is another page that can be requested
	var nextPageAvailable bool
	if userResponses != nil {
		nextPageAvailable = req.By.WithPaging.MaxResults == len(userResponses) // that was a full page, so we know there's another page
	} else {
		nextPageAvailable = assumeNextPageAvailableByDefaultIfError
	}

	return userResponses, nextPageAvailable, err
}

func processUsersResponsePage(resp userRepositoryResponse, subChan chan domain.Subject, errChan chan error) error {
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
