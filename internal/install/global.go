package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// globalHooksDir is the shared directory used for a global install. It is
// configured as git's global core.hooksPath so every repository uses it.
func globalHooksDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}

	return filepath.Join(home, ".config", "githook", "hooks"), nil
}

// setGlobalHooksPath configures git's global core.hooksPath.
func setGlobalHooksPath(dir string) error {
	if err := exec.Command("git", "config", "--global", "core.hooksPath", dir).Run(); err != nil {
		return fmt.Errorf("setting global core.hooksPath: %w", err)
	}

	return nil
}

// unsetGlobalHooksPath clears git's global core.hooksPath when it currently
// points at dir.
func unsetGlobalHooksPath(dir string) error {
	current, err := gitConfigGlobal("core.hooksPath")
	if err != nil || current != dir {
		return nil
	}

	if err := exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run(); err != nil {
		return fmt.Errorf("clearing global core.hooksPath: %w", err)
	}

	return nil
}

// gitConfigGlobal reads a global git config value, returning an empty string
// when it is unset.
func gitConfigGlobal(key string) (string, error) {
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(string(out)), nil
}
