package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.vitalvas.com/gokit/xflags"

	"github.com/vitalvas/githook/internal/install"
)

// newInstallCmd builds the `githook install` command.
func newInstallCmd(out io.Writer) (*xflags.Command, error) {
	var global bool

	cmd := xflags.NewCommand("install", "Install hook symlinks for all supported git hooks")
	cmd.SetDescription(`Create a symlink for every supported git hook, each pointing at the githook
binary. Existing hooks are overwritten. Installs into the current repository's
hooks directory by default, or into the shared global core.hooksPath with
--global.`)

	if err := cmd.BoolVar(&global, "global", "", false, "install into the global core.hooksPath instead of the current repository"); err != nil {
		return nil, err
	}

	cmd.SetRun(func(_ context.Context, _ *xflags.Command, args []string) error {
		if err := requireNoArgs("install", args); err != nil {
			return err
		}

		binary, err := os.Executable()
		if err != nil {
			return err
		}

		dir, err := install.Install(binary, global)
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "Installed %d hooks into %s\n", len(install.HookNames()), dir)

		return nil
	})

	return cmd, nil
}
