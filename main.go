// Command githook is a multi-call binary, busybox-style: when invoked through a
// symlink named after a git hook it runs that hook, and when invoked as
// "githook" it exposes the management CLI (install, uninstall, list).
package main

import (
	"os"

	"github.com/vitalvas/githook/internal/app"
)

func main() {
	os.Exit(app.Run())
}
