package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initRepo creates a git repo and makes it the working directory for the test.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, exec.Command("git", "init", "-q", dir).Run())
	return dir
}

// runCmd builds the root parser, runs it with the given args, and returns its
// combined output.
func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	root, err := newRootParser(&out)
	require.NoError(t, err)
	root.SetArgs(args)
	execErr := root.Execute(context.Background())
	return out.String(), execErr
}

func TestUnknownCommandErrors(t *testing.T) {
	_, err := runCmd(t, "frobnicate")
	assert.Error(t, err)
}

func TestExecuteReturnsZeroForHelp(t *testing.T) {
	out, err := runCmd(t, "--help")
	assert.NoError(t, err)
	assert.Contains(t, out, "githook")
}

func TestGlobalFlagBinds(t *testing.T) {
	for _, cmd := range []string{"install", "uninstall", "list", "status"} {
		t.Run(cmd, func(t *testing.T) {
			root, err := newRootParser(&bytes.Buffer{})
			require.NoError(t, err)
			require.NoError(t, root.Parse(cmd, "--global"))
			assert.Equal(t, cmd, root.SelectedCommand().Name())
			assert.True(t, root.SelectedCommand().IsSet("global"))
		})
	}
}

func TestStrayArgRejected(t *testing.T) {
	initRepo(t)
	for _, cmd := range []string{"install", "uninstall", "list", "status"} {
		t.Run(cmd, func(t *testing.T) {
			_, err := runCmd(t, cmd, "extra")
			assert.ErrorContains(t, err, "takes no arguments")
		})
	}
}

func TestExecuteReadsOSArgs(t *testing.T) {
	t.Chdir(t.TempDir())
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"githook", "--help"}

	assert.Equal(t, 0, Execute())
}

func TestExecuteReturnsOneOnError(t *testing.T) {
	t.Chdir(t.TempDir())
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"githook", "frobnicate"}

	assert.Equal(t, 1, Execute())
}
