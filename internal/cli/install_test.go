package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCommand(t *testing.T) {
	initRepo(t)

	out, err := runCmd(t, "install")
	require.NoError(t, err)
	assert.Contains(t, out, "Installed 14 hooks")
}
