package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/vitalvas/githook/internal/install"
)

// newUninstallCmd builds the `githook uninstall` command.
func newUninstallCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove githook-managed hook symlinks",
		Long: `Remove the hook symlinks created by githook, leaving any unrelated hook files
in place. Operates on the current repository by default, or on the global
core.hooksPath with --global.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			binary, err := os.Executable()
			if err != nil {
				return err
			}

			dir, err := install.Uninstall(binary, global)
			if err != nil {
				return err
			}

			cmd.Printf("Removed githook hooks from %s\n", dir)

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "uninstall from the global core.hooksPath instead of the current repository")

	return cmd
}
