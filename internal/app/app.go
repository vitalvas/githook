// Package app is the multi-call entry point: it inspects the name the binary was
// invoked as and either runs the matching git hook (busybox-style) or hands off
// to the management CLI.
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/vitalvas/githook/internal/cli"
	"github.com/vitalvas/githook/internal/hooks"
)

// Run dispatches based on the name the binary was invoked as. A hook name runs
// the hook; anything else runs the management CLI. It returns the process exit
// code.
func Run() int {
	invoked := filepath.Base(os.Args[0])

	if hooks.IsHookName(invoked) {
		return runHook(invoked, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	}

	return cli.Execute()
}

// runHook executes the named git hook and returns the process exit code.
func runHook(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	ctx := &hooks.Context{
		Hook:   name,
		Args:   args,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	if err := hooks.Dispatch(ctx); err != nil {
		fmt.Fprintf(stderr, "githook: %v\n", err)
		return 1
	}

	return 0
}
