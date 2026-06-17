package hooks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// allowBranchesMarker is the file, relative to the git directory, whose presence
// permits the creation of new branches. When it is absent, the pre-push and
// update hooks reject any operation that would create a new branch.
const allowBranchesMarker = "allow-branches"

// branchRefPrefix is the ref namespace under which branches live. Only refs in
// this namespace are subject to the new-branch policy; tags and other refs are
// not affected.
const branchRefPrefix = "refs/heads/"

// branchCreationsAllowed reports whether new branches may be created, which is
// the case exactly when the allow-branches marker exists in the git directory.
func branchCreationsAllowed() bool {
	_, err := os.Stat(filepath.Join(gitDir(), allowBranchesMarker))
	return err == nil
}

// isZeroOID reports whether an object id is the all-zeroes value git uses for a
// ref that does not yet exist (creation) or is being removed (deletion). The
// length differs between SHA-1 (40) and SHA-256 (64) repositories, so the check
// is on the content rather than a fixed length.
func isZeroOID(oid string) bool {
	if oid == "" {
		return false
	}

	return strings.Trim(oid, "0") == ""
}

// newBranchName returns the short branch name and reports true when ref/oid
// together describe the creation of a new branch: a refs/heads/ ref whose
// pre-update object id is all zeroes (the branch does not yet exist).
func newBranchName(ref, oid string) (string, bool) {
	name, ok := strings.CutPrefix(ref, branchRefPrefix)
	if !ok || !isZeroOID(oid) {
		return "", false
	}

	return name, true
}

// blockMessage is the error text returned when a new branch is rejected. The
// bypass marker is intentionally not mentioned: it is a hidden escape hatch, not
// advertised guidance. The %q branch name is filled in by the caller.
const blockMessage = "creating branch %q is blocked"

// prePushHandler blocks pushes that would create a new branch on the remote
// unless the allow-branches marker is present. Git supplies one line per ref on
// stdin: "<local-ref> <local-oid> <remote-ref> <remote-oid>". A new branch is a
// refs/heads/ remote ref whose remote-oid is all zeroes (the ref does not yet
// exist). Deletions, which also use a zero oid but in the local-oid field, are
// not affected.
func prePushHandler(ctx *Context) error {
	if branchCreationsAllowed() || ctx.Stdin == nil {
		return nil
	}

	scanner := bufio.NewScanner(ctx.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}

		if name, isNew := newBranchName(fields[2], fields[3]); isNew {
			return fmt.Errorf(blockMessage, name)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("pre-push: reading ref list: %w", err)
	}

	return nil
}

// updateHandler blocks the creation of new branches on a repository receiving a
// push unless the allow-branches marker is present. Git invokes the update hook
// once per ref with three arguments: "<ref> <old-oid> <new-oid>". A new branch
// is a refs/heads/ ref whose old-oid is all zeroes. A deletion has a zero
// new-oid and is not affected.
func updateHandler(ctx *Context) error {
	if len(ctx.Args) < 3 {
		return fmt.Errorf("update: expected <ref> <old-oid> <new-oid> arguments")
	}

	name, isNew := newBranchName(ctx.Args[0], ctx.Args[1])
	if !isNew || branchCreationsAllowed() {
		return nil
	}

	return fmt.Errorf(blockMessage, name)
}
