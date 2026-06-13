package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/vitalvas/githook/internal/install"
)

// newInstallCmd builds the `githook install` command.
func newInstallCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install hook symlinks for all supported git hooks",
		Long: `Create a symlink for every supported git hook, each pointing at the githook
binary. Existing hooks are overwritten. Installs into the current repository's
hooks directory by default, or into the shared global core.hooksPath with
--global.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			binary, err := os.Executable()
			if err != nil {
				return err
			}

			dir, err := install.Install(binary, global)
			if err != nil {
				return err
			}

			cmd.Printf("Installed %d hooks into %s\n", len(install.HookNames()), dir)

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "install into the global core.hooksPath instead of the current repository")

	return cmd
}
