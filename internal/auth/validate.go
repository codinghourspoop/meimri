package auth

import "regexp"

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

func ValidUsername(s string) bool { return usernameRe.MatchString(s) }
func ValidPassword(s string) bool { return len(s) >= 8 }
