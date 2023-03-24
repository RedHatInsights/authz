package model

type License struct {
	OrgID       string
	ServiceID   string
	MaxSeats    uint
	assignedIDs map[string]struct{}
}

func NewLicense(orgID string, serviceID string, maxSeats uint, assignedIDs []string) *License {
	l := &License{
		OrgID:       orgID,
		ServiceID:   serviceID,
		MaxSeats:    maxSeats,
		assignedIDs: map[string]struct{}{}, //idiomatic set
	}

	for _, id := range assignedIDs {
		l.assignedIDs[id] = struct{}{}
	}

	return l
}

func (l *License) InUse() uint {
	return uint(len(l.assignedIDs))
}

func (l *License) IsAssigned(principalId string) bool {
	_, ok := l.assignedIDs[principalId]
	return ok
}

func (l *License) IsCompliant() bool {
	return len(l.assignedIDs) <= int(l.MaxSeats)
}

func (l *License) Assign(principalId string) {
	l.assignedIDs[principalId] = struct{}{}
}

func (l *License) UnAssign(principalId string) {
	delete(l.assignedIDs, principalId)
}

func (l *License) GetAssigned() []string {
	assigned := make([]string, len(l.assignedIDs))

	i := 0
	for id, _ := range l.assignedIDs {
		assigned[i] = id
		i++
	}

	return assigned
}
