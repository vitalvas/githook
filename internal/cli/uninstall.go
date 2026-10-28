package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.vitalvas.com/gokit/xflags"

	"github.com/vitalvas/githook/internal/install"
)

// newUninstallCmd builds the `githook uninstall` command.
func newUninstallCmd(out io.Writer) (*xflags.Command, error) {
	var global bool

	cmd := xflags.NewCommand("uninstall", "Remove githook-managed hook symlinks")
	cmd.SetDescription(`Remove the hook symlinks created by githook, leaving any unrelated hook files
in place. Operates on the current repository by default, or on the global
core.hooksPath with --global.`)

	if err := cmd.BoolVar(&global, "global", "", false, "uninstall from the global core.hooksPath instead of the current repository"); err != nil {
		return nil, err
	}

	cmd.SetRun(func(_ context.Context, _ *xflags.Command, args []string) error {
		if err := requireNoArgs("uninstall", args); err != nil {
			return err
		}

		binary, err := os.Executable()
		if err != nil {
			return err
		}

		dir, err := install.Uninstall(binary, global)
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "Removed githook hooks from %s\n", dir)

		return nil
	})

	return cmd, nil
}
