package messaging

// SubjectEventMessage represents a message from the UMB about a subject's lifecycle
type SubjectEventMessage struct {
	Header struct {
		Operation string `xml:"Operation"`
		Type      string `xml:"Type"`
	} `xml:"Header"`
	Payload struct {
		Sync struct {
			User struct {
				Identifiers struct {
					Identifier struct {
						Text       string `xml:",chardata"`
						EntityName string `xml:"entity-name,attr"`
						Qualifier  string `xml:"qualifier,attr"`
					} `xml:"Identifier"`
					Reference []struct {
						Text       string `xml:",chardata"`
						EntityName string `xml:"entity-name,attr"`
						Qualifier  string `xml:"qualifier,attr"`
					} `xml:"Reference"`
				} `xml:"Identifiers"`
				Status struct {
					Primary bool   `xml:"primary,attr"`
					State   string `xml:"State"`
				} `xml:"Status"`
			} `xml:"User"`
		} `xml:"Sync"`
	} `xml:"Payload"`
}

// IsAdded returns true if the event represents a new subject
func (e SubjectEventMessage) IsAdded() bool {
	return e.Header.Operation == "added"
}

// IsUpdated returns true if the event represents an updated subject
func (e SubjectEventMessage) IsUpdated() bool {
	return e.Header.Operation == "updated"
}

// IsActive returns true if the subject is currently active, else false
func (e SubjectEventMessage) IsActive() bool {
	return e.Payload.Sync.User.Status.State == "Active"
}

// SubjectID returns the id of the referenced subject
func (e SubjectEventMessage) SubjectID() string {
	return e.Payload.Sync.User.Identifiers.Identifier.Text
}

// OrgID returns the organization id of the referenced subject
func (e SubjectEventMessage) OrgID() string {
	for _, ref := range e.Payload.Sync.User.Identifiers.Reference {
		if ref.EntityName == "Customer" && ref.Qualifier == "id" {
			return ref.Text
		}
	}

	return ""
}
