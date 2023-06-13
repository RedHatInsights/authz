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

const sortByPrincipal = "principal"
const defaultPageSize = 20
const sortOrder = true

type UserServiceSubjectRepository struct {
	Url        url.URL
	HttpClient http.Client
	Paging     struct {
		PageSize  int
		SortOrder bool
	}
}

func NewUserServiceSubjectRepository(url url.URL, client http.Client) UserServiceSubjectRepository {
	return UserServiceSubjectRepository{
		Url:        url,
		HttpClient: client,
		Paging: struct {
			PageSize  int
			SortOrder bool
		}{PageSize: defaultPageSize, SortOrder: sortOrder},
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
		shouldFetchPage := true
		currentPage := 0

		for shouldFetchPage {
			req := u.makeUserRepositoryRequest(orgID, currentPage*u.Paging.PageSize)

			resp, nextPageAvailable, serviceCallErr := u.doPagedUserServiceCall(req, errChan)

			var pageProcessingErr error
			if resp != nil {
				pageProcessingErr = processUsersResponsePage(resp, subChan, errChan)
			}

			currentPage += 1
			shouldFetchPage = shouldFetchNextPage(nextPageAvailable, serviceCallErr, pageProcessingErr)
		}

		defer func() {
			close(subChan)
			close(errChan)
		}()
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
	req.By.WithPaging.MaxResults = u.Paging.PageSize + 1 // add 1 to peek into next page to see if there is one
	req.By.WithPaging.SortBy = sortByPrincipal
	req.By.WithPaging.Ascending = u.Paging.SortOrder
	req.Include.AllOf = []string{"status"}

	return req
}

func (u *UserServiceSubjectRepository) doPagedUserServiceCall(req userRepositoryRequest, errChan chan error) (userRepositoryResponse, bool, error) {
	// TODO: put this somewhere better or change it (we only know that another page doesn't exist on a success)
	assumeNextPageAvailableIfError := true

	// Make request with marshalled JSON as the POST body
	userRepositoryRequestJSON, err := json.Marshal(req)

	if err != nil {
		err = fmt.Errorf("error marshalling userRepositoryRequest: %v: %w", req, err)
		errChan <- err
		return nil, assumeNextPageAvailableIfError, err
	}

	resp, err := u.HttpClient.Post(u.Url.String(), "application/json", bytes.NewBuffer(userRepositoryRequestJSON))

	if err != nil {
		err = fmt.Errorf("failed to POST to UserService: %v: %w", u.Url, err)
		errChan <- err
		return nil, assumeNextPageAvailableIfError, err
	}

	if resp != nil && resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected http response status code on request to user repository: %v", resp.Status)
		errChan <- err
		return nil, assumeNextPageAvailableIfError, err
	}

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
		errChan <- err
		return nil, assumeNextPageAvailableIfError, err
	}
	defer resp.Body.Close()

	var userResponses userRepositoryResponse
	err = json.Unmarshal(body, &userResponses)

	if err != nil {
		err = fmt.Errorf("failed to unmarshall userRepositoryResponse from body: %v, %w", string(body), err)
		errChan <- err
	}

	var nextPageAvailable bool
	if userResponses != nil {
		nextPageAvailable = req.By.WithPaging.MaxResults == len(userResponses) // that was a full page + 1, so we know there's another page
	} else {
		nextPageAvailable = assumeNextPageAvailableIfError
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

func shouldFetchNextPage(anotherPageAvailable bool, serviceCallErr error, pageProcessingErr error) bool {
	// TODO: Determine whether to keep going assuming there is another page and the error is the "right" type of error

	return anotherPageAvailable && serviceCallErr == nil && pageProcessingErr == nil
}
