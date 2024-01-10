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
