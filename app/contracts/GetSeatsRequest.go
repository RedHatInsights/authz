package contracts

import "authz/app"

// GetSeatsRequest - Representation of the request
type GetSeatsRequest struct {
	Request
	Org                app.Organization
	Service            app.Service
	IncludeUsers       bool
	IncludeLicenseInfo bool
	Filter             string
}
