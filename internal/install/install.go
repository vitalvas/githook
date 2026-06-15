// Package install manages the githook hook symlinks: creating them for every
// supported hook type, removing them, and reporting their status. Hooks can be
// installed into the current repository's .git/hooks directory or into a shared
// directory configured as git's global core.hooksPath.
package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vitalvas/githook/internal/hooks"
)

// HookNames returns the list of git hook names that githook manages.
func HookNames() []string {
	return hooks.Names
}

// HooksDir resolves the directory that hook symlinks are installed into. When
// global is true it returns the shared global directory; otherwise it returns
// the current repository's hooks directory.
func HooksDir(global bool) (string, error) {
	if global {
		return globalHooksDir()
	}

	return repoHooksDir()
}

// repoHooksDir returns the absolute path to the current repository's hooks
// directory. A core.hooksPath set in the repository's local config is honoured
// (resolved the way git resolves it: relative values are relative to the work
// tree root). A core.hooksPath inherited from the global or system config is
// deliberately ignored so that a non-global install always targets this
// repository rather than a shared directory.
func repoHooksDir() (string, error) {
	if path, err := gitConfigLocal("core.hooksPath"); err == nil && path != "" {
		return resolveRepoPath(path)
	}

	// Resolve the hooks directory from the git directory itself rather than via
	// "git rev-parse --git-path hooks": that form honours core.hooksPath from any
	// scope, including the global config, which would redirect a non-global
	// install away from this repository.
	gitDir, err := gitRevParse("--absolute-git-dir")
	if err != nil {
		return "", err
	}

	return filepath.Join(gitDir, "hooks"), nil
}

// resolveRepoPath turns a path reported by git into an absolute path. Absolute
// paths are returned unchanged; relative paths are resolved against the work
// tree root, matching how git interprets a relative core.hooksPath.
func resolveRepoPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	root, err := gitRevParse("--show-toplevel")
	if err != nil {
		return "", err
	}

	return filepath.Join(root, path), nil
}

// Install creates a symlink for every supported hook in the target directory,
// each pointing at the githook binary. Existing entries are overwritten. When
// global is true it also points git's global core.hooksPath at the directory.
// It returns the directory hooks were installed into.
func Install(binary string, global bool) (string, error) {
	binary, err := filepath.Abs(binary)
	if err != nil {
		return "", fmt.Errorf("resolving binary path: %w", err)
	}

	dir, err := HooksDir(global)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating hooks directory: %w", err)
	}

	for _, name := range hooks.Names {
		link := filepath.Join(dir, name)
		if err := replaceSymlink(binary, link); err != nil {
			return "", err
		}
	}

	if global {
		if err := setGlobalHooksPath(dir); err != nil {
			return "", err
		}
	}

	return dir, nil
}

// Uninstall removes the githook-managed symlinks from the target directory,
// leaving unrelated hook files untouched. When global is true it also clears
// git's global core.hooksPath if it points at the managed directory. It returns
// the directory hooks were removed from.
func Uninstall(binary string, global bool) (string, error) {
	binary, err := filepath.Abs(binary)
	if err != nil {
		return "", fmt.Errorf("resolving binary path: %w", err)
	}

	dir, err := HooksDir(global)
	if err != nil {
		return "", err
	}

	for _, name := range hooks.Names {
		link := filepath.Join(dir, name)
		if !managedSymlink(link, binary) {
			continue
		}

		if err := os.Remove(link); err != nil {
			return "", fmt.Errorf("removing hook %s: %w", name, err)
		}
	}

	if global {
		if err := unsetGlobalHooksPath(dir); err != nil {
			return "", err
		}
	}

	return dir, nil
}

// replaceSymlink creates link as a symlink to target, removing any existing
// file, directory entry, or symlink at link first.
func replaceSymlink(target, link string) error {
	if _, err := os.Lstat(link); err == nil {
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("removing existing hook %s: %w", filepath.Base(link), err)
		}
	}

	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("linking hook %s: %w", filepath.Base(link), err)
	}

	return nil
}

// managedSymlink reports whether link is a symlink pointing at binary, i.e. one
// created by githook.
func managedSymlink(link, binary string) bool {
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false
	}

	target, err := os.Readlink(link)
	if err != nil {
		return false
	}

	return target == binary
}

// gitRevParse runs `git rev-parse` with the given arguments and returns the
// trimmed output.
func gitRevParse(args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"rev-parse"}, args...)...).Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// gitConfigLocal reads a git config value from the repository's local config
// only, ignoring values inherited from the global or system config. It returns
// an empty string when the key is unset locally.
func gitConfigLocal(key string) (string, error) {
	out, err := exec.Command("git", "config", "--local", "--get", key).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
