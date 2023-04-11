package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrincipalIsAnonymousTrueForAnonymousPrincipal(t *testing.T) {
	p := NewAnonymousPrincipal()

	assert.True(t, p.IsAnonymous(), "Should have been anonymous.")
}

func TestPrincipalIsAnonymousFalseForSpecificPrincipal(t *testing.T) {
	p := NewPrincipal("alice", "Alice", "aspian")

	assert.False(t, p.IsAnonymous(), "Should NOT have been anonymous.")
}
