// Package hooks defines the supported git hook types and provides the
// busybox-style dispatch used when the binary is invoked through a symlink
// named after one of those hooks.
package hooks

// Names lists the git client-side and server-side hook names that githook
// manages. The install command creates one symlink per name pointing back to
// the githook binary, and dispatch uses the invoked name to select behaviour.
var Names = []string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"post-receive",
}

// IsHookName reports whether name is one of the managed git hook names.
func IsHookName(name string) bool {
	for _, n := range Names {
		if n == name {
			return true
		}
	}

	return false
}
