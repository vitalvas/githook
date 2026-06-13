package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vitalvas/githook/internal/hooks"
)

// isolateGlobalGit redirects git's global config and the user home directory to
// throwaway temp locations so global-install tests never touch the real
// environment. It returns the fake home directory.
func isolateGlobalGit(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig"))
	return home
}

func globalConfig(t *testing.T, key string) string {
	t.Helper()
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func TestGlobalInstallAndUninstall(t *testing.T) {
	home := isolateGlobalGit(t)
	binary := fakeBinary(t)

	dir, err := Install(binary, true)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "githook", "hooks"), dir)

	t.Run("links every hook to the binary", func(t *testing.T) {
		for _, name := range hooks.Names {
			target, err := os.Readlink(filepath.Join(dir, name))
			require.NoError(t, err)
			assert.Equal(t, binary, target)
		}
	})

	t.Run("sets global core.hooksPath", func(t *testing.T) {
		assert.Equal(t, dir, globalConfig(t, "core.hooksPath"))
	})

	t.Run("status reports the global directory", func(t *testing.T) {
		gotDir, statuses, err := Status(binary, true)
		require.NoError(t, err)
		assert.Equal(t, dir, gotDir)
		assert.Len(t, statuses, len(hooks.Names))
	})

	t.Run("uninstall clears links and global config", func(t *testing.T) {
		_, err := Uninstall(binary, true)
		require.NoError(t, err)

		for _, name := range hooks.Names {
			_, err := os.Lstat(filepath.Join(dir, name))
			assert.True(t, os.IsNotExist(err))
		}
		assert.Empty(t, globalConfig(t, "core.hooksPath"))
	})
}

func TestUnsetGlobalHooksPathKeepsForeignValue(t *testing.T) {
	isolateGlobalGit(t)

	// A core.hooksPath pointing somewhere else must be left untouched.
	other := t.TempDir()
	require.NoError(t, exec.Command("git", "config", "--global", "core.hooksPath", other).Run())

	require.NoError(t, unsetGlobalHooksPath("/some/githook/dir"))
	assert.Equal(t, other, globalConfig(t, "core.hooksPath"))
}
