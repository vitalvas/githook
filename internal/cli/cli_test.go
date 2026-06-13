package cli

import (
	"bytes"
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

// runCmd builds the root command, runs it with the given args, and returns its
// combined output.
func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestUnknownCommandErrors(t *testing.T) {
	_, err := runCmd(t, "frobnicate")
	assert.Error(t, err)
}

func TestExecuteReturnsZeroForHelp(t *testing.T) {
	// Execute() reads os.Args via cobra; exercising the success path here keeps
	// the wrapper covered without spawning a process.
	root := newRootCmd()
	root.SetArgs([]string{"--help"})
	root.SetOut(&bytes.Buffer{})
	assert.NoError(t, root.Execute())
}
