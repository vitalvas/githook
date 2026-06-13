package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunHook(t *testing.T) {
	t.Run("succeeds for passing hook", func(t *testing.T) {
		t.Chdir(t.TempDir())

		var stdout, stderr bytes.Buffer
		code := runHook("pre-commit", nil, strings.NewReader(""), &stdout, &stderr)
		assert.Equal(t, 0, code)
		assert.Empty(t, stderr.String())
	})

	t.Run("returns 1 and reports error for failing hook", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		msg := filepath.Join(dir, "msg")
		require.NoError(t, os.WriteFile(msg, []byte("# comment only\n"), 0o600))

		var stdout, stderr bytes.Buffer
		code := runHook("commit-msg", []string{msg}, strings.NewReader(""), &stdout, &stderr)
		assert.Equal(t, 1, code)
		assert.Contains(t, stderr.String(), "githook:")
	})

	t.Run("returns 1 for unknown hook", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := runHook("bogus", nil, strings.NewReader(""), &stdout, &stderr)
		assert.Equal(t, 1, code)
		assert.Contains(t, stderr.String(), "unknown hook")
	})
}

func TestRunAsCLI(t *testing.T) {
	// When not invoked as a hook name, Run() drops into the CLI. Invoke with a
	// help flag so cobra exits zero without side effects.
	t.Chdir(t.TempDir())
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"githook", "--help"}

	assert.Equal(t, 0, Run())
}

func TestRunAsHookName(t *testing.T) {
	// Invoking under a hook name dispatches to the hook rather than the CLI.
	t.Chdir(t.TempDir())
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"/somewhere/pre-commit"}

	assert.Equal(t, 0, Run())
}
