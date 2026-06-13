package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommandReportsManaged(t *testing.T) {
	initRepo(t)

	_, err := runCmd(t, "install")
	require.NoError(t, err)

	out, err := runCmd(t, "list")
	require.NoError(t, err)
	assert.Contains(t, out, "Hooks directory:")
	assert.Contains(t, out, "pre-commit")
	assert.Contains(t, out, "managed")
}

func TestStatusAliasWorks(t *testing.T) {
	initRepo(t)

	out, err := runCmd(t, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "missing")
}
