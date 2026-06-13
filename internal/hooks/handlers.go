package hooks

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Context carries the runtime inputs a hook handler may need. It mirrors how
// git invokes a hook: positional arguments and, for some hooks, data on stdin.
type Context struct {
	// Hook is the name of the hook being run (e.g. "commit-msg").
	Hook string
	// Args are the positional arguments git passed to the hook.
	Args []string
	// Stdin is the hook's standard input. Only a few hooks receive data here
	// (pre-push, pre-receive, post-receive, update via args).
	Stdin io.Reader
	// Stdout and Stderr are where the handler writes user-facing output.
	Stdout io.Writer
	Stderr io.Writer
}

// Handler is the built-in behaviour for a single git hook. Returning a non-nil
// error aborts the hook (git treats a non-zero exit as a failed hook).
type Handler func(ctx *Context) error

// builtins maps hook names to their compiled-in handler. Hooks without an entry
// use noopHandler, which always succeeds.
var builtins = map[string]Handler{
	"commit-msg": commitMsgHandler,
}

// HandlerFor returns the built-in handler for the named hook, falling back to a
// no-op handler when none is registered.
func HandlerFor(hook string) Handler {
	if h, ok := builtins[hook]; ok {
		return h
	}

	return noopHandler
}

// noopHandler succeeds without doing anything. It is the default for hooks that
// have no built-in validation.
func noopHandler(_ *Context) error {
	return nil
}

// commitMsgHandler validates the commit message against the conventional-commit
// policy. Git passes the path to the message file as the first argument; comment
// lines git adds to the file are stripped before validation.
func commitMsgHandler(ctx *Context) error {
	if len(ctx.Args) < 1 {
		return fmt.Errorf("commit-msg: expected message file path argument")
	}

	data, err := os.ReadFile(ctx.Args[0])
	if err != nil {
		return fmt.Errorf("commit-msg: opening message file: %w", err)
	}

	if err := validateCommitMessage(stripComments(string(data))); err != nil {
		return fmt.Errorf("commit-msg: %w", err)
	}

	return nil
}

// stripComments removes git comment lines (those beginning with "#") from the
// raw message file content, leaving the user-authored text for validation.
func stripComments(message string) string {
	lines := strings.Split(message, "\n")

	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		kept = append(kept, line)
	}

	return strings.Join(kept, "\n")
}
