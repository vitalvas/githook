package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vitalvas/githook/internal/hooks"
)

// initRepo creates a git repository in a temp dir, makes it the working
// directory for the test, and returns its path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, exec.Command("git", "init", "-q", dir).Run())
	return dir
}

func fakeBinary(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "githook")
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755))
	return path
}

func TestInstallAndUninstallRepo(t *testing.T) {
	initRepo(t)
	binary := fakeBinary(t)

	dir, err := Install(binary, false)
	require.NoError(t, err)

	t.Run("creates a symlink per hook pointing at the binary", func(t *testing.T) {
		for _, name := range hooks.Names {
			link := filepath.Join(dir, name)
			target, err := os.Readlink(link)
			require.NoError(t, err, "hook %s should be a symlink", name)
			assert.Equal(t, binary, target)
		}
	})

	t.Run("uninstall removes managed links", func(t *testing.T) {
		_, err := Uninstall(binary, false)
		require.NoError(t, err)
		for _, name := range hooks.Names {
			_, err := os.Lstat(filepath.Join(dir, name))
			assert.True(t, os.IsNotExist(err), "hook %s should be removed", name)
		}
	})
}

func TestInstallOverwritesExisting(t *testing.T) {
	initRepo(t)
	binary := fakeBinary(t)

	dir, err := HooksDir(false)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dir, 0o755))

	// Pre-existing regular file at a hook path must be overwritten.
	existing := filepath.Join(dir, "pre-commit")
	require.NoError(t, os.WriteFile(existing, []byte("old"), 0o755))

	_, err = Install(binary, false)
	require.NoError(t, err)

	target, err := os.Readlink(existing)
	require.NoError(t, err)
	assert.Equal(t, binary, target)
}

func TestUninstallLeavesForeignHooks(t *testing.T) {
	initRepo(t)
	binary := fakeBinary(t)

	dir, err := Install(binary, false)
	require.NoError(t, err)

	// Replace one managed link with a foreign file.
	foreign := filepath.Join(dir, "pre-commit")
	require.NoError(t, os.Remove(foreign))
	require.NoError(t, os.WriteFile(foreign, []byte("keep me"), 0o755))

	_, err = Uninstall(binary, false)
	require.NoError(t, err)

	data, err := os.ReadFile(foreign)
	require.NoError(t, err)
	assert.Equal(t, "keep me", string(data))
}

func TestHooksDirHonoursCoreHooksPath(t *testing.T) {
	dir := initRepo(t)
	custom := filepath.Join(dir, "custom-hooks")
	require.NoError(t, exec.Command("git", "config", "core.hooksPath", custom).Run())

	got, err := HooksDir(false)
	require.NoError(t, err)
	assert.Equal(t, custom, got)
}

func TestHooksDirOutsideRepo(t *testing.T) {
	// A bare temp dir with no .git and an isolated HOME is not a repository, so
	// resolving the repo hooks directory must fail clearly.
	t.Chdir(t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))

	_, err := HooksDir(false)
	assert.ErrorContains(t, err, "git repository")
}

func TestHookNames(t *testing.T) {
	assert.Equal(t, hooks.Names, HookNames())
}

func TestManagedSymlink(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "githook")
	require.NoError(t, os.WriteFile(binary, []byte("x"), 0o755))

	t.Run("true for link to binary", func(t *testing.T) {
		link := filepath.Join(dir, "managed")
		require.NoError(t, os.Symlink(binary, link))
		assert.True(t, managedSymlink(link, binary))
	})

	t.Run("false for link to other target", func(t *testing.T) {
		link := filepath.Join(dir, "other")
		require.NoError(t, os.Symlink("/somewhere/else", link))
		assert.False(t, managedSymlink(link, binary))
	})

	t.Run("false for regular file", func(t *testing.T) {
		file := filepath.Join(dir, "regular")
		require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
		assert.False(t, managedSymlink(file, binary))
	})

	t.Run("false for missing path", func(t *testing.T) {
		assert.False(t, managedSymlink(filepath.Join(dir, "absent"), binary))
	})
}
