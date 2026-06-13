package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHookName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "known client hook", in: "pre-commit", want: true},
		{name: "known server hook", in: "pre-receive", want: true},
		{name: "commit-msg", in: "commit-msg", want: true},
		{name: "unknown name", in: "githook", want: false},
		{name: "empty", in: "", want: false},
		{name: "case sensitive", in: "Pre-Commit", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsHookName(tt.in))
		})
	}
}

func TestNamesCoversAllManagedHooks(t *testing.T) {
	assert.Len(t, Names, 14)

	seen := make(map[string]struct{}, len(Names))
	for _, n := range Names {
		assert.NotContains(t, seen, n, "duplicate hook name %q", n)
		seen[n] = struct{}{}
	}
}
