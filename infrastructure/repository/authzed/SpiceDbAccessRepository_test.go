package authzed

import (
	"authz/domain"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var container *LocalSpiceDbContainer

func TestMain(m *testing.M) {
	factory := NewLocalSpiceDbContainerFactory()
	var err error
	container, err = factory.CreateContainer()

	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		os.Exit(-1)
	}

	result := m.Run()

	container.Close()
	os.Exit(result)
}

func TestCheckAccess(t *testing.T) {
	t.Parallel()
	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	cases := []struct {
		sub       domain.SubjectID
		operation string
		resource  domain.Resource
		expected  domain.AccessDecision
	}{
		{sub: "u1", operation: "access", resource: domain.Resource{Type: "license", ID: "o1/smarts"}, expected: true},
		{sub: "u1", operation: "access", resource: domain.Resource{Type: "license", ID: "o1/doesnotexist"}, expected: false},
		{sub: "doesnotexist", operation: "access", resource: domain.Resource{Type: "license", ID: "o1/smarts"}, expected: false},
	}

	for _, testcase := range cases {
		actual, err := client.CheckAccess(testcase.sub, testcase.operation, testcase.resource)
		assert.NoError(t, err, fmt.Sprintf("Error in case (subject: %s, operation: %s, resource: [%s, %s])", testcase.sub, testcase.operation, testcase.resource.Type, testcase.resource.ID))
		assert.Equal(t, testcase.expected, actual, "Unexpected result for case (subject: %s, operation: %s, resource: [%s, %s])", testcase.sub, testcase.operation, testcase.resource.Type, testcase.resource.ID)
	}
}

func TestGetLicense(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	lic, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, "o1", lic.OrgID)
	assert.Equal(t, "smarts", lic.ServiceID)
	assert.Equal(t, 10, lic.MaxSeats)
	assert.Equal(t, 2, lic.InUse) //u1, u3
}

func TestGetAssignable(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	assignable, err := client.GetAssignable("o1", "smarts")
	assert.NoError(t, err)
	initialAssignableUsers := []domain.SubjectID{"u2", "u5", "u6", "u7", "u8", "u9", "u10", "u11", "u12", "u13", "u14", "u15", "u16", "u17", "u18", "u19", "u20"}

	assert.ElementsMatch(t, initialAssignableUsers, assignable)
}

func TestGetAssigned(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	assigned, err := client.GetAssigned("o1", "smarts")
	assert.NoError(t, err)

	assert.ElementsMatch(t, []domain.SubjectID{"u1", "u3"}, assigned)
}

func TestAssignBatch(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	// given
	subs := []domain.SubjectID{
		"u6", "u7",
	}

	oldLic, e1 := client.GetLicense("o1", "smarts")
	assert.NoError(t, e1)
	assert.Equal(t, 2, oldLic.InUse)

	// when
	err = client.ModifySeats(subs, []domain.SubjectID{}, oldLic, "o1", domain.Service{ID: "smarts"})

	// then
	assert.NoError(t, err)
	newLic, e2 := client.GetLicense("o1", "smarts")
	assert.NoError(t, e2)
	assert.Equal(t, oldLic.InUse+len(subs), newLic.InUse)
}

func TestFailAssignBatchIfOneDisabled(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	// given
	subs := []domain.SubjectID{
		"u4", "u101",
	}
	oldLic, e1 := client.GetLicense("o1", "smarts")
	assert.NoError(t, e1)
	assert.Equal(t, 2, oldLic.InUse) // u1, u3

	// when
	err = client.ModifySeats(subs, []domain.SubjectID{}, oldLic, "o1", domain.Service{ID: "smarts"})

	// then
	assert.Error(t, err)

	expectedSameLicense, e2 := client.GetLicense("o1", "smarts")
	assert.NoError(t, e2)
	assert.Equal(t, 2, expectedSameLicense.InUse) // still u1, u3, so if error in batch nothing gets applied.
}

func TestUnassignBatch(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	// given
	subs := []domain.SubjectID{
		"u1", "u3",
	}

	oldLic, e1 := client.GetLicense("o1", "smarts")
	assert.NoError(t, e1)
	assert.Equal(t, 2, oldLic.InUse) //u1, u3

	// when
	err = client.ModifySeats([]domain.SubjectID{}, subs, oldLic, "o1", domain.Service{ID: "smarts"})

	// then
	assert.NoError(t, err)
	newLic, e2 := client.GetLicense("o1", "smarts")
	assert.NoError(t, e2)
	assert.Equal(t, oldLic.InUse-len(subs), newLic.InUse)
}

func TestAssignUnassign(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	oldLic, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	err = client.ModifySeats([]domain.SubjectID{"u2"}, []domain.SubjectID{}, oldLic, "o1", domain.Service{ID: "smarts"})
	assert.NoError(t, err)

	lic, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, 3, lic.InUse) //u1, u2, u3

	err = client.ModifySeats([]domain.SubjectID{}, []domain.SubjectID{"u2"}, lic, "o1", domain.Service{ID: "smarts"})
	assert.NoError(t, err)

	lic, err = client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, 2, lic.InUse) //u1, u3
}

func TestUnassignNotAssigned(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	licBefore, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	err = client.ModifySeats([]domain.SubjectID{}, []domain.SubjectID{"not_assigned"}, licBefore, "o1", domain.Service{ID: "smarts"})
	assert.Error(t, err)

	licAfter, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, licBefore.InUse, licAfter.InUse)
}

func TestAssignAlreadyAssigned(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	licBefore, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	err = client.ModifySeats([]domain.SubjectID{"u1"}, []domain.SubjectID{}, licBefore, "o1", domain.Service{ID: "smarts"})
	assert.Error(t, err)

	licAfter, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, licBefore.InUse, licAfter.InUse)
}

func TestFailAssignForDisabled(t *testing.T) {
	t.Parallel()

	client, _, err := container.CreateClient()
	assert.NoError(t, err)

	licBefore, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	err = client.ModifySeats([]domain.SubjectID{"u4"}, []domain.SubjectID{}, licBefore, "o1", domain.Service{ID: "smarts"})

	assert.Error(t, err)
}

func TestHasAnyLicenseReturnsTrueForOrgWithLicenseWithoutUsers(t *testing.T) {
	t.Parallel()
	repository, _, err := container.CreateClient()
	assert.NoError(t, err)

	result, err := repository.HasAnyLicense("oNoUsers")
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestHasAnyLicenseReturnsTrueForOrgWithLicenseWithUsers(t *testing.T) {
	t.Parallel()
	repository, _, err := container.CreateClient()
	assert.NoError(t, err)

	result, err := repository.HasAnyLicense("o1")
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestHasAnyLicenseReturnsFalseForUnknownOrgID(t *testing.T) {
	t.Parallel()
	repository, _, err := container.CreateClient()
	assert.NoError(t, err)

	result, err := repository.HasAnyLicense("unknownOrgId")
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestHasAnyLicenseReturnsFalseForOrgWithoutLicense(t *testing.T) {
	t.Parallel()
	repository, _, err := container.CreateClient()
	assert.NoError(t, err)

	result, err := repository.HasAnyLicense("o2")
	assert.NoError(t, err)
	assert.False(t, result)
}

// Six scenarios:
// Scenario 1: If previously enabled, delete nonexistent tombstone (no change)
// Scenario 2: if previously disabled, delete tombstone, now enabled
// Scenario 3: if previously disabled, touch tombstone that's already there (no change)
// Scenario 4: if previously enabled, add tombstone, now disabled
// Scenario 5: new enabled user -> created, no tombstone
// Scenario 6: new disabled user -> created with tombstone
func TestUpsertUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		orgID   string
		subject domain.Subject
	}{
		{
			"o1",
			domain.Subject{
				SubjectID: "u1", //Enabled in seed data
				Enabled:   true,
			},
		},
		{
			"o1",
			domain.Subject{
				SubjectID: "u3", //Disabled in seed data
				Enabled:   true,
			},
		},
		{
			"o1",
			domain.Subject{
				SubjectID: "u4", //Disabled in seed data
				Enabled:   false,
			},
		},
		{
			"o1",
			domain.Subject{
				SubjectID: "u2", //Enabled in seed data
				Enabled:   false,
			},
		},
		{
			"o1",
			domain.Subject{
				SubjectID: "new-enabled",
				Enabled:   true,
			},
		},
		{
			"o1",
			domain.Subject{
				SubjectID: "new-disabled",
				Enabled:   false,
			},
		},
	}

	repository, client, err := container.CreateClient()
	assert.NoError(t, err)

	for _, tt := range tests {
		err = repository.UpsertSubject(tt.orgID, tt.subject)
		assert.NoError(t, err)

		//Assert relationship exists user -member-> org
		userExists := CheckForSubjectRelationship(client, tt.subject.SubjectID, "member", "org", tt.orgID)
		assert.True(t, userExists)
		tombstoned := CheckForSubjectRelationship(client, tt.subject.SubjectID, "disabled", "org", tt.orgID)

		if tt.subject.Enabled {
			assert.False(t, tombstoned)
		} else {
			assert.True(t, tombstoned)
		}
	}
}
