package hooks

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// yakeBinary is the development tool run by the pre-commit hook to execute the
// project's tests and policy checks.
const yakeBinary = "yake"

// skipMarker is the file, relative to the git directory, whose presence bypasses
// the pre-commit checks.
const skipMarker = "skip-pre-commit"

// preCommitHandler runs `yake run` before a commit when the yake tool is
// available on PATH. When yake is not installed the hook succeeds quietly so
// the commit is not blocked in environments without it. When yake is present,
// its exit status is propagated: a failing check aborts the commit. The checks
// are bypassed entirely when the skip marker file exists in the git directory.
func preCommitHandler(ctx *Context) error {
	if skipPreCommit() {
		return nil
	}

	path, err := exec.LookPath(yakeBinary)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Fprintln(ctx.Stderr, "githook: yake not found, skipping pre-commit checks")
			return nil
		}

		return fmt.Errorf("pre-commit: locating yake: %w", err)
	}

	cmd := exec.Command(path, "run")
	cmd.Stdout = ctx.Stdout
	cmd.Stderr = ctx.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pre-commit: yake run failed: %w", err)
	}

	return nil
}

// skipPreCommit reports whether the bypass marker file exists in the git
// directory.
func skipPreCommit() bool {
	_, err := os.Stat(filepath.Join(gitDir(), skipMarker))
	return err == nil
}

// gitDir resolves the repository's git directory. Git exports GIT_DIR in a
// hook's environment; when it is absent the conventional ".git" directory
// relative to the current working directory is used.
func gitDir() string {
	if dir := os.Getenv("GIT_DIR"); dir != "" {
		return dir
	}

	return ".git"
}
