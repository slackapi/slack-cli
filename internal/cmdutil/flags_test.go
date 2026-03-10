package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestIsFlagChanged(t *testing.T) {
	tests := map[string]struct {
		flag     string
		setFlag  bool
		expected bool
	}{
		"returns true when flag is set": {
			flag:     "app-id",
			setFlag:  true,
			expected: true,
		},
		"returns false when flag exists but not set": {
			flag:     "app-id",
			setFlag:  false,
			expected: false,
		},
		"returns false when flag does not exist": {
			flag:     "nonexistent",
			setFlag:  false,
			expected: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("app-id", "", "app ID")
			if tc.setFlag && tc.flag == "app-id" {
				cmd.Flags().Set("app-id", "A12345")
			}
			result := IsFlagChanged(cmd, tc.flag)
			assert.Equal(t, tc.expected, result)
		})
	}
}
