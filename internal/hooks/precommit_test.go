package hooks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeYake writes an executable named "yake" into dir whose body is the given
// shell script, and points PATH at dir so the handler resolves it.
func fakeYake(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, yakeBinary)
	require.NoError(t, os.WriteFile(path, []byte(fmt.Sprintf("#!/bin/sh\n%s", script)), 0o755))
	t.Setenv("PATH", dir)
}

func TestPreCommitHandler(t *testing.T) {
	newCtx := func() (*Context, *bytes.Buffer, *bytes.Buffer) {
		var stdout, stderr bytes.Buffer
		return &Context{Hook: "pre-commit", Stdout: &stdout, Stderr: &stderr}, &stdout, &stderr
	}

	// Point GIT_DIR at an empty directory so the bypass marker is absent unless
	// a test creates it explicitly.
	t.Setenv("GIT_DIR", t.TempDir())

	t.Run("skips when yake is absent", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		ctx, _, stderr := newCtx()
		assert.NoError(t, preCommitHandler(ctx))
		assert.Contains(t, stderr.String(), "yake not found")
	})

	t.Run("succeeds when yake run passes", func(t *testing.T) {
		fakeYake(t, "exit 0\n")

		ctx, _, _ := newCtx()
		assert.NoError(t, preCommitHandler(ctx))
	})

	t.Run("fails when yake run fails", func(t *testing.T) {
		fakeYake(t, "exit 1\n")

		ctx, _, _ := newCtx()
		assert.ErrorContains(t, preCommitHandler(ctx), "yake run failed")
	})

	t.Run("forwards yake output to streams", func(t *testing.T) {
		fakeYake(t, "echo to-out; echo to-err >&2\n")

		ctx, stdout, stderr := newCtx()
		require.NoError(t, preCommitHandler(ctx))
		assert.Contains(t, stdout.String(), "to-out")
		assert.Contains(t, stderr.String(), "to-err")
	})

	t.Run("passes the run subcommand to yake", func(t *testing.T) {
		fakeYake(t, "printf '%s' \"$1\" > \"$GITHOOK_ARGFILE\"\n")
		argFile := filepath.Join(t.TempDir(), "arg")
		t.Setenv("GITHOOK_ARGFILE", argFile)

		ctx, _, _ := newCtx()
		require.NoError(t, preCommitHandler(ctx))
		data, err := os.ReadFile(argFile)
		require.NoError(t, err)
		assert.Equal(t, "run", string(data))
	})

	t.Run("bypasses checks when skip marker exists", func(t *testing.T) {
		gitDir := t.TempDir()
		t.Setenv("GIT_DIR", gitDir)
		require.NoError(t, os.WriteFile(filepath.Join(gitDir, skipMarker), nil, 0o600))

		// A failing yake would abort the commit; the marker must short-circuit
		// before yake is ever run, and do so silently.
		fakeYake(t, "exit 1\n")

		ctx, stdout, stderr := newCtx()
		assert.NoError(t, preCommitHandler(ctx))
		assert.Empty(t, stdout.String())
		assert.Empty(t, stderr.String())
	})
}

func TestSkipPreCommit(t *testing.T) {
	t.Run("true when marker present", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GIT_DIR", dir)
		require.NoError(t, os.WriteFile(filepath.Join(dir, skipMarker), nil, 0o600))
		assert.True(t, skipPreCommit())
	})

	t.Run("false when marker absent", func(t *testing.T) {
		t.Setenv("GIT_DIR", t.TempDir())
		assert.False(t, skipPreCommit())
	})
}

func TestGitDir(t *testing.T) {
	t.Run("uses GIT_DIR when set", func(t *testing.T) {
		t.Setenv("GIT_DIR", "/custom/git/dir")
		assert.Equal(t, "/custom/git/dir", gitDir())
	})

	t.Run("falls back to .git", func(t *testing.T) {
		t.Setenv("GIT_DIR", "")
		assert.Equal(t, ".git", gitDir())
	})
}
