package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrincipalIsAnonymousTrueForAnonymousPrincipal(t *testing.T) {
	p := NewAnonymousPrincipal()

	assert.True(t, p.IsAnonymous(), "Should have been anonymous.")
}

func TestPrincipalIsAnonymousFalseForSpecificPrincipal(t *testing.T) {
	p := NewPrincipal("u1", "alice", "org123", true)

	assert.False(t, p.IsAnonymous(), "Should NOT have been anonymous.")
}
