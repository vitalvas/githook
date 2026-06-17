package hooks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pushLine builds a single pre-push stdin line from its four fields.
func pushLine(localRef, localOID, remoteRef, remoteOID string) string {
	return fmt.Sprintf("%s %s %s %s\n", localRef, localOID, remoteRef, remoteOID)
}

const (
	zeroSHA1   = "0000000000000000000000000000000000000000"
	zeroSHA256 = "0000000000000000000000000000000000000000000000000000000000000000"
	someOID    = "1111111111111111111111111111111111111111"
	otherOID   = "2222222222222222222222222222222222222222"
)

// allowBranches creates the allow-branches marker in an isolated git directory
// and points GIT_DIR at it. Without calling this, branch creation is blocked.
func allowBranches(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GIT_DIR", dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, allowBranchesMarker), nil, 0o600))
}

// blockBranches points GIT_DIR at an empty directory so the marker is absent.
func blockBranches(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_DIR", t.TempDir())
}

func TestIsZeroOID(t *testing.T) {
	assert.True(t, isZeroOID(zeroSHA1))
	assert.True(t, isZeroOID(zeroSHA256))
	assert.False(t, isZeroOID(someOID))
	assert.False(t, isZeroOID(""))
	assert.False(t, isZeroOID("0000000000000000000000000000000000000001"))
}

func TestNewBranchName(t *testing.T) {
	t.Run("new branch on heads ref with zero oid", func(t *testing.T) {
		name, isNew := newBranchName("refs/heads/feature/x", zeroSHA1)
		assert.True(t, isNew)
		assert.Equal(t, "feature/x", name)
	})

	t.Run("existing branch is not new", func(t *testing.T) {
		_, isNew := newBranchName("refs/heads/main", someOID)
		assert.False(t, isNew)
	})

	t.Run("non-branch ref is not new", func(t *testing.T) {
		_, isNew := newBranchName("refs/tags/v1", zeroSHA1)
		assert.False(t, isNew)
	})
}

func TestPrePushHandler(t *testing.T) {
	newCtx := func(stdin string) (*Context, *bytes.Buffer) {
		var stderr bytes.Buffer
		return &Context{Hook: "pre-push", Stdin: strings.NewReader(stdin), Stderr: &stderr}, &stderr
	}

	t.Run("blocks pushing a new branch", func(t *testing.T) {
		blockBranches(t)
		ctx, _ := newCtx(pushLine("refs/heads/feature", someOID, "refs/heads/feature", zeroSHA1))
		err := prePushHandler(ctx)
		assert.ErrorContains(t, err, "feature")
		assert.ErrorContains(t, err, "blocked")
		// The bypass marker must not be advertised in the user-facing message.
		assert.NotContains(t, err.Error(), allowBranchesMarker)
	})

	t.Run("allows updating an existing branch", func(t *testing.T) {
		blockBranches(t)
		ctx, _ := newCtx(pushLine("refs/heads/main", otherOID, "refs/heads/main", someOID))
		assert.NoError(t, prePushHandler(ctx))
	})

	t.Run("allows deleting a branch", func(t *testing.T) {
		blockBranches(t)
		// A deletion: the local side is the zero oid, the remote ref exists.
		ctx, _ := newCtx(pushLine("(delete)", zeroSHA1, "refs/heads/old", someOID))
		assert.NoError(t, prePushHandler(ctx))
	})

	t.Run("ignores non-branch refs", func(t *testing.T) {
		blockBranches(t)
		ctx, _ := newCtx(pushLine("refs/tags/v1", someOID, "refs/tags/v1", zeroSHA1))
		assert.NoError(t, prePushHandler(ctx))
	})

	t.Run("blocks a new branch among several refs", func(t *testing.T) {
		blockBranches(t)
		lines := fmt.Sprintf("%s%s",
			pushLine("refs/heads/main", otherOID, "refs/heads/main", someOID),
			pushLine("refs/heads/new", someOID, "refs/heads/new", zeroSHA256))
		ctx, _ := newCtx(lines)
		assert.ErrorContains(t, prePushHandler(ctx), "new")
	})

	t.Run("allows new branch when marker present", func(t *testing.T) {
		allowBranches(t)
		ctx, _ := newCtx(pushLine("refs/heads/feature", someOID, "refs/heads/feature", zeroSHA1))
		assert.NoError(t, prePushHandler(ctx))
	})

	t.Run("ignores malformed lines", func(t *testing.T) {
		blockBranches(t)
		ctx, _ := newCtx("garbage line\n\n")
		assert.NoError(t, prePushHandler(ctx))
	})

	t.Run("nil stdin is a no-op", func(t *testing.T) {
		blockBranches(t)
		assert.NoError(t, prePushHandler(&Context{Hook: "pre-push", Stderr: &bytes.Buffer{}}))
	})
}

func TestUpdateHandler(t *testing.T) {
	newCtx := func(args ...string) *Context {
		return &Context{Hook: "update", Args: args, Stderr: &bytes.Buffer{}}
	}

	t.Run("blocks creating a new branch", func(t *testing.T) {
		blockBranches(t)
		err := updateHandler(newCtx("refs/heads/feature", zeroSHA1, someOID))
		assert.ErrorContains(t, err, "feature")
		assert.ErrorContains(t, err, "blocked")
		assert.NotContains(t, err.Error(), allowBranchesMarker)
	})

	t.Run("allows updating an existing branch", func(t *testing.T) {
		blockBranches(t)
		assert.NoError(t, updateHandler(newCtx("refs/heads/main", otherOID, someOID)))
	})

	t.Run("allows deleting a branch", func(t *testing.T) {
		blockBranches(t)
		assert.NoError(t, updateHandler(newCtx("refs/heads/old", someOID, zeroSHA1)))
	})

	t.Run("ignores non-branch refs", func(t *testing.T) {
		blockBranches(t)
		assert.NoError(t, updateHandler(newCtx("refs/tags/v1", zeroSHA1, someOID)))
	})

	t.Run("allows new branch when marker present", func(t *testing.T) {
		allowBranches(t)
		assert.NoError(t, updateHandler(newCtx("refs/heads/feature", zeroSHA1, someOID)))
	})

	t.Run("errors on missing arguments", func(t *testing.T) {
		blockBranches(t)
		assert.ErrorContains(t, updateHandler(newCtx("refs/heads/x")), "expected")
	})
}
