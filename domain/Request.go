package domain

// A Request represents the parameters common to all requests
type Request struct {
	//The principal sending the request
	Requestor SubjectID
}
