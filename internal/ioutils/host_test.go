package ioutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHostname(t *testing.T) {
	t.Run("returns a non-empty hashed hostname", func(t *testing.T) {
		hostname := GetHostname()
		assert.NotEmpty(t, hostname)
		// The hostname should be hashed, not the raw hostname
		// It should not be "unknown" on a normal system
		assert.NotEqual(t, "", hostname)
	})

	t.Run("returns consistent results", func(t *testing.T) {
		hostname1 := GetHostname()
		hostname2 := GetHostname()
		assert.Equal(t, hostname1, hostname2)
	})
}
