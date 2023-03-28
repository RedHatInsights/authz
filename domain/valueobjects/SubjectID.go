package valueobjects

// SubjectID represents a reference to a subject on the platform
type SubjectID string

// HasIdentity is a helper method that indicates whether this SubjectID represents an identity
func (p SubjectID) HasIdentity() bool {
	return p != ""
}
