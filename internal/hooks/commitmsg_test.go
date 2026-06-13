package hooks

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommitMessage(t *testing.T) {
	valid := []struct {
		name string
		msg  string
	}{
		{name: "type and subject", msg: "feat: add new thing"},
		{name: "type scope and subject", msg: "fix(parser): handle empty input"},
		{name: "perf type", msg: "perf: speed up lookup"},
		{name: "deps type", msg: "deps: bump testify"},
		{name: "revert type", msg: "revert: undo the change"},
		{name: "docs type", msg: "docs: clarify usage"},
		{name: "chore type", msg: "chore: tidy module"},
		{name: "exclamation in subject is allowed", msg: "fix: handle it now!"},
		{name: "scope with special chars", msg: "fix(commit-msg/v2): something"},
		{name: "exactly 72 chars", msg: fmt.Sprintf("feat: %s", strings.Repeat("a", 66))},
		{name: "trailing newline ignored", msg: "feat: add thing\n"},
	}

	for _, tt := range valid {
		t.Run(fmt.Sprintf("valid/%s", tt.name), func(t *testing.T) {
			assert.NoError(t, validateCommitMessage(tt.msg))
		})
	}

	invalid := []struct {
		name    string
		msg     string
		wantErr string
	}{
		{name: "empty", msg: "", wantErr: "empty"},
		{name: "whitespace only", msg: "   \n\n", wantErr: "empty"},
		{name: "missing colon", msg: "feat add thing", wantErr: "must follow"},
		{name: "missing space after colon", msg: "feat:add thing", wantErr: "must follow"},
		{name: "two spaces after colon", msg: "feat:  add thing", wantErr: "whitespace"},
		{name: "disallowed type", msg: "feature: add thing", wantErr: "not allowed"},
		{name: "build type rejected", msg: "build: compile", wantErr: "not allowed"},
		{name: "uppercase type rejected", msg: "Feat: add thing", wantErr: "must follow"},
		{name: "breaking marker bare", msg: "feat!: add thing", wantErr: "breaking change"},
		{name: "breaking marker with scope", msg: "feat(api)!: add thing", wantErr: "breaking change"},
		{name: "empty scope", msg: "feat(): add thing", wantErr: "must follow"},
		{name: "empty subject", msg: "feat: ", wantErr: "must follow"},
		{name: "trailing period", msg: "feat: add thing.", wantErr: "period"},
		{name: "subject too long", msg: fmt.Sprintf("feat: %s", strings.Repeat("a", 67)), wantErr: "at most 72"},
		{name: "has body", msg: "feat: add thing\n\nthis is a body", wantErr: "only a subject"},
		{name: "multiline subject", msg: "feat: add thing\nmore text", wantErr: "only a subject"},
	}

	for _, tt := range invalid {
		t.Run(fmt.Sprintf("invalid/%s", tt.name), func(t *testing.T) {
			err := validateCommitMessage(tt.msg)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestHasBreakingChangeMarker(t *testing.T) {
	assert.True(t, hasBreakingChangeMarker("feat!: x"))
	assert.True(t, hasBreakingChangeMarker("feat(api)!: x"))
	assert.False(t, hasBreakingChangeMarker("feat: handle it!"))
	assert.False(t, hasBreakingChangeMarker("no colon here"))
}

func TestIsAllowedType(t *testing.T) {
	assert.True(t, isAllowedType("feat"))
	assert.True(t, isAllowedType("chore"))
	assert.False(t, isAllowedType("build"))
	assert.False(t, isAllowedType(""))
}
