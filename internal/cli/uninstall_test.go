package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUninstallCommand(t *testing.T) {
	initRepo(t)

	_, err := runCmd(t, "install")
	require.NoError(t, err)

	out, err := runCmd(t, "uninstall")
	require.NoError(t, err)
	assert.Contains(t, out, "Removed githook hooks")

	listOut, err := runCmd(t, "list")
	require.NoError(t, err)
	assert.NotContains(t, listOut, "managed")
}
