// Package cli builds the githook command-line interface used when the binary is
// invoked directly (not through a hook symlink).
package cli

import (
	"github.com/spf13/cobra"
)

// Execute builds the root command and runs it, returning the process exit code.
func Execute() int {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		return 1
	}

	return 0
}

// newRootCmd assembles the root command and its subcommands.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "githook",
		Short:         "Manage git hooks through a single multi-call binary",
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newListCmd())

	return root
}
