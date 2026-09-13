package common

import "strings"

// likeEscapes are the characters that MySQL's LIKE operator treats as
// metacharacters when the default escape character (backslash) is in effect.
const likeEscapes = `\%_`

// EscapeLikePattern neutralises LIKE metacharacters in a value that is about to
// be bound to a LIKE pattern. The value itself is always passed as a query
// parameter, so this is not an injection fix: it stops user input from *shaping*
// the pattern. A search for `%` (or `_`, or a trailing `\`) previously matched
// every row and defeated the caller's own filter — an unindexed full scan that
// also leaked the whole table through a filter the caller believed was applied.
//
// It escapes `\`, `%` and `_` with a backslash, which is MySQL's default LIKE
// escape character, so no `ESCAPE` clause is needed in the SQL. The value is
// returned unchanged when it contains nothing to escape.
func EscapeLikePattern(s string) string {
	if !strings.ContainsAny(s, likeEscapes) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + 8)
	for _, r := range s {
		switch r {
		case '\\', '%', '_':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
