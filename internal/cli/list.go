package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.vitalvas.com/gokit/xflags"

	"go.vitalvas.com/githook/internal/install"
)

// newListCmds builds the `githook list` command and its `status` alias. xflags
// has no alias API, so the alias is a second command sharing the same handler.
func newListCmds(out io.Writer) ([]*xflags.Command, error) {
	list, err := newListCmd(out, "list")
	if err != nil {
		return nil, err
	}

	status, err := newListCmd(out, "status")
	if err != nil {
		return nil, err
	}

	return []*xflags.Command{list, status}, nil
}

func newListCmd(out io.Writer, name string) (*xflags.Command, error) {
	var global bool

	cmd := xflags.NewCommand(name, "Show the install status of each git hook")
	cmd.SetDescription(`Report, for every supported git hook, whether it is managed by githook,
occupied by an unrelated file, or missing. Inspects the current repository by
default, or the global core.hooksPath with --global.`)

	if err := cmd.BoolVar(&global, "global", "", false, "inspect the global core.hooksPath instead of the current repository"); err != nil {
		return nil, err
	}

	cmd.SetRun(func(_ context.Context, _ *xflags.Command, args []string) error {
		if err := requireNoArgs(name, args); err != nil {
			return err
		}

		binary, err := os.Executable()
		if err != nil {
			return err
		}

		dir, statuses, err := install.Status(binary, global)
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "Hooks directory: %s\n\n", dir)
		for _, s := range statuses {
			if s.Target != "" {
				fmt.Fprintf(out, "  %-20s %-8s -> %s\n", s.Name, s.State, s.Target)
				continue
			}

			fmt.Fprintf(out, "  %-20s %s\n", s.Name, s.State)
		}

		return nil
	})

	return cmd, nil
}
