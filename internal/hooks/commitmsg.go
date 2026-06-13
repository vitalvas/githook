package hooks

import (
	"fmt"
	"regexp"
	"strings"
)

// maxSubjectLength is the maximum allowed length of the commit header line
// (type, optional scope, ": ", and subject combined).
const maxSubjectLength = 72

// allowedTypes are the conventional-commit types githook accepts.
var allowedTypes = []string{"feat", "fix", "perf", "deps", "revert", "docs", "chore"}

// headerPattern parses a conventional-commit header into its type, optional
// scope, and subject. The breaking-change "!" marker is intentionally excluded
// so that a header carrying it fails to match and is rejected.
//
//	group 1: type
//	group 2: scope (without surrounding parentheses), empty when absent
//	group 3: subject
var headerPattern = regexp.MustCompile(`^([a-z]+)(?:\(([^)]+)\))?: (.+)$`)

// validateCommitMessage enforces the commit message policy: a single-line
// conventional-commit header with an allowed type, no breaking-change marker,
// no body, a non-empty subject without a trailing period, and a header no
// longer than maxSubjectLength characters. The message is the raw file content
// as written by the user (git comment lines are stripped before validation).
func validateCommitMessage(message string) error {
	header, err := commitHeader(message)
	if err != nil {
		return err
	}

	if len(header) > maxSubjectLength {
		return fmt.Errorf("subject must be at most %d characters, got %d", maxSubjectLength, len(header))
	}

	if hasBreakingChangeMarker(header) {
		return fmt.Errorf("breaking change marker %q is not allowed", "!")
	}

	groups := headerPattern.FindStringSubmatch(header)
	if groups == nil {
		return fmt.Errorf("message must follow %q or %q", "<type>: <subject>", "<type>(<scope>): <subject>")
	}

	commitType, subject := groups[1], groups[3]

	if !isAllowedType(commitType) {
		return fmt.Errorf("type %q is not allowed; use one of: %s", commitType, strings.Join(allowedTypes, ", "))
	}

	if strings.HasSuffix(subject, ".") {
		return fmt.Errorf("subject must not end with a period")
	}

	if strings.TrimSpace(subject) != subject {
		return fmt.Errorf("subject must not have leading or trailing whitespace")
	}

	return nil
}

// commitHeader extracts the single content line from message, rejecting empty
// messages and any message that spans more than one content line (a body or
// other multi-line content is not allowed). Git comment lines have already been
// stripped by the caller.
func commitHeader(message string) (string, error) {
	lines := strings.Split(strings.TrimRight(message, "\n"), "\n")

	content := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		content = append(content, line)
	}

	if len(content) == 0 {
		return "", fmt.Errorf("commit message is empty")
	}

	if len(content) > 1 {
		return "", fmt.Errorf("commit message must contain only a subject line; body and multi-line messages are not allowed")
	}

	return content[0], nil
}

// hasBreakingChangeMarker reports whether the header carries the conventional-
// commit breaking-change marker "!", which sits in the prefix immediately
// before the colon (e.g. "feat!:" or "feat(api)!:"). The check is limited to
// the prefix so that a "!" appearing in the subject text is not rejected.
func hasBreakingChangeMarker(header string) bool {
	prefix, _, found := strings.Cut(header, ": ")
	if !found {
		return false
	}

	return strings.Contains(prefix, "!")
}

// isAllowedType reports whether t is one of the accepted conventional-commit
// types.
func isAllowedType(t string) bool {
	for _, allowed := range allowedTypes {
		if t == allowed {
			return true
		}
	}

	return false
}
