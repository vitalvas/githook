package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vitalvas/githook/internal/config"
)

// shell is the interpreter used to run config-defined rule commands.
const shell = "/bin/sh"

// Dispatch runs the named hook: it executes the built-in handler first and, if
// that succeeds, runs any rules defined for the hook in the repository config.
// A failing built-in handler or a failing rule (without allow_failure) returns
// an error, which the caller surfaces as a non-zero exit.
func Dispatch(ctx *Context) error {
	if !IsHookName(ctx.Hook) {
		return fmt.Errorf("unknown hook: %s", ctx.Hook)
	}

	if err := HandlerFor(ctx.Hook)(ctx); err != nil {
		return err
	}

	cfg, err := loadRepoConfig()
	if err != nil {
		return err
	}

	for _, rule := range cfg.RulesFor(ctx.Hook) {
		if err := runRule(ctx, rule); err != nil {
			return err
		}
	}

	return nil
}

// loadRepoConfig finds and loads the githook config from the current working
// directory's repository root. A missing config yields an empty config.
func loadRepoConfig() (*config.Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolving working directory: %w", err)
	}

	for _, name := range config.FileNames {
		path := filepath.Join(cwd, name)
		if _, err := os.Stat(path); err == nil {
			return config.Load(path)
		}
	}

	return config.Load()
}

// runRule executes a single config rule through the shell, streaming its output
// to the hook context. A non-zero exit aborts the hook unless the rule sets
// allow_failure.
func runRule(ctx *Context, rule config.Rule) error {
	if rule.Run == "" {
		return nil
	}

	// With "sh -c <script>", trailing arguments are assigned to the positional
	// parameters starting at $0. Pass the hook name as $0 so the hook's own
	// arguments map to $1, $2, ... inside the rule command.
	args := append([]string{"-c", rule.Run, ctx.Hook}, ctx.Args...)

	cmd := exec.Command(shell, args...)
	cmd.Stdin = ctx.Stdin
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr

	if err := cmd.Run(); err != nil {
		if rule.AllowFailure {
			fmt.Fprintf(ctx.Stderr, "githook: rule %q failed (ignored): %v\n", ruleLabel(rule), err)
			return nil
		}

		return fmt.Errorf("rule %q failed: %w", ruleLabel(rule), err)
	}

	return nil
}

// ruleLabel returns a display name for a rule, preferring its explicit name and
// falling back to the command itself.
func ruleLabel(rule config.Rule) string {
	if rule.Name != "" {
		return rule.Name
	}

	return rule.Run
}
