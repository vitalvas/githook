// Package cli builds the githook command-line interface used when the binary is
// invoked directly (not through a hook symlink).
package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.vitalvas.com/gokit/xflags"
)

// Execute builds the root parser and runs it against os.Args, returning the
// process exit code.
func Execute() int {
	root, err := newRootParser(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "githook: %v\n", err)
		return 1
	}

	if err := root.Execute(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "githook: %v\n", err)
		return 1
	}

	return 0
}

// newRootParser assembles the root parser and its subcommands. Handler and
// built-in help/version output are written to out.
func newRootParser(out io.Writer) (*xflags.Parser, error) {
	root := xflags.New("githook")
	root.SetDescription("Manage git hooks through a single multi-call binary")
	root.SetOutput(out)

	installCmd, err := newInstallCmd(out)
	if err != nil {
		return nil, err
	}

	uninstallCmd, err := newUninstallCmd(out)
	if err != nil {
		return nil, err
	}

	listCmds, err := newListCmds(out)
	if err != nil {
		return nil, err
	}

	commands := append([]*xflags.Command{installCmd, uninstallCmd}, listCmds...)
	if err := root.AddCommand(commands...); err != nil {
		return nil, err
	}

	return root, nil
}

// requireNoArgs reports an error when a command that takes no positional
// arguments is given some.
func requireNoArgs(name string, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("%s takes no arguments", name)
	}

	return nil
}
