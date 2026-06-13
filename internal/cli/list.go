package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/vitalvas/githook/internal/install"
)

// newListCmd builds the `githook list` command, which reports hook status.
func newListCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"status"},
		Short:   "Show the install status of each git hook",
		Long: `Report, for every supported git hook, whether it is managed by githook,
occupied by an unrelated file, or missing. Inspects the current repository by
default, or the global core.hooksPath with --global.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			binary, err := os.Executable()
			if err != nil {
				return err
			}

			dir, statuses, err := install.Status(binary, global)
			if err != nil {
				return err
			}

			cmd.Printf("Hooks directory: %s\n\n", dir)
			for _, s := range statuses {
				if s.Target != "" {
					cmd.Printf("  %-20s %-8s -> %s\n", s.Name, s.State, s.Target)
					continue
				}

				cmd.Printf("  %-20s %s\n", s.Name, s.State)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "inspect the global core.hooksPath instead of the current repository")

	return cmd
}
