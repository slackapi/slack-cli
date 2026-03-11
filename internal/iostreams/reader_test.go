package iostreams

import (
	"strings"
	"testing"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/slackdeps"
	"github.com/stretchr/testify/assert"
)

func TestReadIn(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"returns the configured stdin reader": {
			input:    "test input",
			expected: "test input",
		},
		"returns an empty reader": {
			input:    "",
			expected: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			cfg := config.NewConfig(fsMock, osMock)
			io := NewIOStreams(cfg, fsMock, osMock)
			io.Stdin = strings.NewReader(tc.input)

			reader := io.ReadIn()

			buf := make([]byte, len(tc.input)+1)
			n, _ := reader.Read(buf)
			assert.Equal(t, tc.expected, string(buf[:n]))
		})
	}
}
