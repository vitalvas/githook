package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateString(t *testing.T) {
	assert.Equal(t, "managed", StateManaged.String())
	assert.Equal(t, "foreign", StateForeign.String())
	assert.Equal(t, "missing", StateMissing.String())
	assert.Equal(t, "missing", State(99).String())
}

func TestStatus(t *testing.T) {
	initRepo(t)
	binary := fakeBinary(t)

	dir, err := HooksDir(false)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dir, 0o755))

	// One managed, one foreign file, one foreign symlink; the rest missing.
	require.NoError(t, os.Symlink(binary, filepath.Join(dir, "pre-commit")))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "commit-msg"), []byte("x"), 0o755))
	require.NoError(t, os.Symlink("/elsewhere", filepath.Join(dir, "pre-push")))

	gotDir, statuses, err := Status(binary, false)
	require.NoError(t, err)
	assert.Equal(t, dir, gotDir)

	byName := make(map[string]HookStatus, len(statuses))
	for _, s := range statuses {
		byName[s.Name] = s
	}

	assert.Equal(t, StateManaged, byName["pre-commit"].State)
	assert.Equal(t, binary, byName["pre-commit"].Target)
	assert.Equal(t, StateForeign, byName["commit-msg"].State)
	assert.Equal(t, StateForeign, byName["pre-push"].State)
	assert.Equal(t, "/elsewhere", byName["pre-push"].Target)
	assert.Equal(t, StateMissing, byName["post-commit"].State)
}
