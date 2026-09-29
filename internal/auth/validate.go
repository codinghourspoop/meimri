package auth

import (
	"regexp"
	"strings"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

// MaxDisplayNameLength is the longest display name the API accepts. Display
// names are free-form single-line text, so only length and whitespace are
// constrained.
const MaxDisplayNameLength = 64

func ValidUsername(s string) bool { return usernameRe.MatchString(s) }
func ValidPassword(s string) bool { return len(s) >= 8 }

// ValidDisplayName accepts any non-empty single-line name that is not padded
// with whitespace. Interior spaces are allowed; control characters are not.
func ValidDisplayName(s string) bool {
	if s == "" || len(s) > MaxDisplayNameLength || strings.TrimSpace(s) != s {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}
