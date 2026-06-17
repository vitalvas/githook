package hooks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vitalvas/githook/internal/config"
)

func TestDispatch(t *testing.T) {
	t.Run("rejects unknown hook", func(t *testing.T) {
		err := Dispatch(&Context{Hook: "not-a-hook", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
		assert.ErrorContains(t, err, "unknown hook")
	})

	t.Run("runs built-in handler and fails the hook", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		path := filepath.Join(dir, "msg")
		require.NoError(t, os.WriteFile(path, []byte("# only comment\n"), 0o600))

		err := Dispatch(&Context{Hook: "commit-msg", Args: []string{path}, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
		assert.ErrorContains(t, err, "empty")
	})

	t.Run("runs config rule after passing built-in", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		marker := filepath.Join(dir, "ran")
		writeConfig(t, dir, fmt.Sprintf("hooks:\n  post-merge:\n    - run: touch %s\n", marker))

		var stderr bytes.Buffer
		err := Dispatch(&Context{Hook: "post-merge", Stdout: &bytes.Buffer{}, Stderr: &stderr})
		require.NoError(t, err)
		assert.FileExists(t, marker)
	})

	t.Run("failing rule aborts the hook", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		writeConfig(t, dir, "hooks:\n  post-merge:\n    - name: fail\n      run: exit 3\n")

		err := Dispatch(&Context{Hook: "post-merge", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
		assert.ErrorContains(t, err, "fail")
	})

	t.Run("allow_failure lets the hook pass", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		writeConfig(t, dir, "hooks:\n  post-merge:\n    - run: exit 1\n      allow_failure: true\n")

		var stderr bytes.Buffer
		err := Dispatch(&Context{Hook: "post-merge", Stdout: &bytes.Buffer{}, Stderr: &stderr})
		require.NoError(t, err)
		assert.Contains(t, stderr.String(), "ignored")
	})

	t.Run("passes hook args to rule as positional params", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		out := filepath.Join(dir, "out")
		writeConfig(t, dir, fmt.Sprintf("hooks:\n  post-merge:\n    - run: printf '%%s' \"$1\" > %s\n", out))

		err := Dispatch(&Context{Hook: "post-merge", Args: []string{"origin"}, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
		require.NoError(t, err)
		data, err := os.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "origin", string(data))
	})

	t.Run("no config runs only built-in", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)

		err := Dispatch(&Context{Hook: "post-merge", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
		assert.NoError(t, err)
	})
}

func TestRunRule(t *testing.T) {
	t.Run("empty run is skipped", func(t *testing.T) {
		err := runRule(&Context{Stderr: &bytes.Buffer{}}, config.Rule{})
		assert.NoError(t, err)
	})

	t.Run("stdin is forwarded", func(t *testing.T) {
		dir := t.TempDir()
		out := filepath.Join(dir, "stdin")
		ctx := &Context{
			Hook:   "pre-receive",
			Stdin:  strings.NewReader("payload"),
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}
		err := runRule(ctx, config.Rule{Run: fmt.Sprintf("cat > %s", out)})
		require.NoError(t, err)
		data, err := os.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "payload", string(data))
	})
}

func TestRuleLabel(t *testing.T) {
	assert.Equal(t, "named", ruleLabel(config.Rule{Name: "named", Run: "cmd"}))
	assert.Equal(t, "cmd", ruleLabel(config.Rule{Run: "cmd"}))
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".githook.yaml"), []byte(content), 0o600))
}
