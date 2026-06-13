package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("no filenames yields empty config", func(t *testing.T) {
		cfg, err := Load()
		require.NoError(t, err)
		assert.Empty(t, cfg.Hooks)
	})

	t.Run("missing file yields empty config", func(t *testing.T) {
		cfg, err := Load(filepath.Join(t.TempDir(), "absent.yaml"))
		require.NoError(t, err)
		assert.Empty(t, cfg.Hooks)
	})

	t.Run("parses hooks and rules", func(t *testing.T) {
		path := writeFile(t, `
hooks:
  pre-commit:
    - name: lint
      run: golangci-lint run
    - run: go test ./...
      allow_failure: true
`)
		cfg, err := Load(path)
		require.NoError(t, err)

		rules := cfg.RulesFor("pre-commit")
		require.Len(t, rules, 2)
		assert.Equal(t, "lint", rules[0].Name)
		assert.Equal(t, "golangci-lint run", rules[0].Run)
		assert.False(t, rules[0].AllowFailure)
		assert.True(t, rules[1].AllowFailure)
	})

	t.Run("rejects unknown fields in strict mode", func(t *testing.T) {
		path := writeFile(t, "hooks:\n  pre-commit:\n    - run: x\n      bogus: true\n")
		_, err := Load(path)
		assert.Error(t, err)
	})

	t.Run("rejects malformed yaml", func(t *testing.T) {
		path := writeFile(t, "hooks: [::invalid")
		_, err := Load(path)
		assert.Error(t, err)
	})
}

func TestRulesFor(t *testing.T) {
	t.Run("nil config returns nil", func(t *testing.T) {
		var cfg *Config
		assert.Nil(t, cfg.RulesFor("pre-commit"))
	})

	t.Run("missing hook returns nil", func(t *testing.T) {
		cfg := &Config{Hooks: map[string][]Rule{"pre-commit": {{Run: "x"}}}}
		assert.Nil(t, cfg.RulesFor("commit-msg"))
	})
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".githook.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}
