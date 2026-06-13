package hooks

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerFor(t *testing.T) {
	t.Run("returns built-in for registered hook", func(t *testing.T) {
		assert.NotNil(t, HandlerFor("commit-msg"))
	})

	t.Run("returns noop for unregistered hook", func(t *testing.T) {
		err := HandlerFor("post-commit")(&Context{Hook: "post-commit"})
		assert.NoError(t, err)
	})
}

func TestCommitMsgHandler(t *testing.T) {
	writeMsg := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}

	newCtx := func(args ...string) *Context {
		return &Context{Hook: "commit-msg", Args: args, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}
	}

	t.Run("accepts non-empty message", func(t *testing.T) {
		path := writeMsg(t, "fix: a real message\n")
		assert.NoError(t, commitMsgHandler(newCtx(path)))
	})

	t.Run("accepts message with leading comments", func(t *testing.T) {
		path := writeMsg(t, "# comment\n\nfeat: actual subject\n")
		assert.NoError(t, commitMsgHandler(newCtx(path)))
	})

	t.Run("rejects empty message", func(t *testing.T) {
		path := writeMsg(t, "\n  \n")
		assert.Error(t, commitMsgHandler(newCtx(path)))
	})

	t.Run("rejects comment-only message", func(t *testing.T) {
		path := writeMsg(t, "# only a comment\n#another\n")
		assert.Error(t, commitMsgHandler(newCtx(path)))
	})

	t.Run("rejects invalid conventional commit", func(t *testing.T) {
		path := writeMsg(t, "Add a thing\n")
		err := commitMsgHandler(newCtx(path))
		assert.ErrorContains(t, err, "commit-msg:")
	})

	t.Run("rejects breaking change marker", func(t *testing.T) {
		path := writeMsg(t, "feat!: drop support\n")
		assert.ErrorContains(t, commitMsgHandler(newCtx(path)), "breaking change")
	})

	t.Run("errors without argument", func(t *testing.T) {
		assert.Error(t, commitMsgHandler(newCtx()))
	})

	t.Run("errors on missing file", func(t *testing.T) {
		assert.Error(t, commitMsgHandler(newCtx(filepath.Join(t.TempDir(), "nope"))))
	})
}

func TestStripComments(t *testing.T) {
	t.Run("removes hash comment lines", func(t *testing.T) {
		got := stripComments("feat: x\n# a comment\n# another\n")
		assert.Equal(t, "feat: x\n", got)
	})

	t.Run("keeps non-comment lines", func(t *testing.T) {
		got := stripComments("feat: x\nbody\n")
		assert.Equal(t, "feat: x\nbody\n", got)
	})
}
