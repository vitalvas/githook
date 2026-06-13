// Package hooks defines the supported git hook types and provides the
// busybox-style dispatch used when the binary is invoked through a symlink
// named after one of those hooks.
package hooks

// Names lists the git client-side and server-side hook names that githook
// manages. The install command creates one symlink per name pointing back to
// the githook binary, and dispatch uses the invoked name to select behaviour.
//
// Only hooks whose contract is "run and check the exit status" are managed.
// Hooks that git invokes with a special protocol or that override git's own
// behaviour are intentionally excluded, because a no-op symlink to this binary
// would break them: push-to-checkout (replaces the built-in checkout),
// proc-receive (pkt-line protocol), and fsmonitor-watchman (emits the changed
// file list on stdout).
var Names = []string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"pre-merge-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"sendemail-validate",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
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
