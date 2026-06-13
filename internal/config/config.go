// Package config loads the optional .githook.yaml rules file that extends the
// built-in hook handlers with user-defined commands.
package config

import (
	"fmt"

	"github.com/vitalvas/gokit/xconfig"
)

// FileNames lists the config file names searched in the repository root, in
// precedence order. The first existing file is used.
var FileNames = []string{
	".githook.yaml",
	".githook.yml",
}

// Config is the root of the .githook.yaml file. Hooks maps a git hook name to
// the rules executed for that hook after the built-in handler succeeds.
type Config struct {
	Hooks map[string][]Rule `yaml:"hooks" json:"hooks"`
}

// Rule is a single user-defined command run for a hook. A non-zero exit code
// aborts the hook unless AllowFailure is set.
type Rule struct {
	// Name is an optional human-readable label shown in output.
	Name string `yaml:"name" json:"name"`
	// Run is the command line executed through the system shell.
	Run string `yaml:"run" json:"run"`
	// AllowFailure lets the hook continue even when this rule exits non-zero.
	AllowFailure bool `yaml:"allow_failure" json:"allow_failure"`
}

// Load reads the first existing config file from filenames. A missing file is
// not an error and yields an empty Config so callers can run built-in handlers
// only. Passing no filenames also yields an empty Config.
func Load(filenames ...string) (*Config, error) {
	cfg := &Config{}

	if len(filenames) == 0 {
		return cfg, nil
	}

	if err := xconfig.Load(cfg, xconfig.WithFiles(filenames...), xconfig.WithStrict(true)); err != nil {
		return nil, fmt.Errorf("loading githook config: %w", err)
	}

	return cfg, nil
}

// RulesFor returns the configured rules for the named hook, or nil when none
// are defined.
func (c *Config) RulesFor(hook string) []Rule {
	if c == nil || c.Hooks == nil {
		return nil
	}

	return c.Hooks[hook]
}
