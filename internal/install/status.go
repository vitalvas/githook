package install

import (
	"os"
	"path/filepath"

	"github.com/vitalvas/githook/internal/hooks"
)

// State describes the installed status of a single hook.
type State int

const (
	// StateMissing means no file exists at the hook path.
	StateMissing State = iota
	// StateManaged means the hook is a symlink created by githook.
	StateManaged
	// StateForeign means a file or symlink exists that githook did not create.
	StateForeign
)

// String returns a human-readable label for the state.
func (s State) String() string {
	switch s {
	case StateManaged:
		return "managed"
	case StateForeign:
		return "foreign"
	default:
		return "missing"
	}
}

// HookStatus pairs a hook name with its current state and link target.
type HookStatus struct {
	Name   string
	State  State
	Target string
}

// Status reports the state of every supported hook in the target directory.
// binary is the githook binary path used to distinguish managed symlinks from
// foreign hook files.
func Status(binary string, global bool) (string, []HookStatus, error) {
	binary, err := filepath.Abs(binary)
	if err != nil {
		return "", nil, err
	}

	dir, err := HooksDir(global)
	if err != nil {
		return "", nil, err
	}

	statuses := make([]HookStatus, 0, len(hooks.Names))
	for _, name := range hooks.Names {
		statuses = append(statuses, hookStatus(dir, name, binary))
	}

	return dir, statuses, nil
}

// hookStatus inspects a single hook path and classifies it.
func hookStatus(dir, name, binary string) HookStatus {
	link := filepath.Join(dir, name)

	info, err := os.Lstat(link)
	if err != nil {
		return HookStatus{Name: name, State: StateMissing}
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return HookStatus{Name: name, State: StateForeign}
	}

	target, err := os.Readlink(link)
	if err != nil {
		return HookStatus{Name: name, State: StateForeign, Target: target}
	}

	if target == binary {
		return HookStatus{Name: name, State: StateManaged, Target: target}
	}

	return HookStatus{Name: name, State: StateForeign, Target: target}
}
