package report

import (
	"html"
	"strings"
)

// sanitizeTerminal removes control characters that could inject ANSI/OSC
// escape sequences or cursor-control bytes into a terminal. C0 controls
// (0x00–0x1F), DEL (0x7F), and C1 controls (0x80–0x9F) are dropped, except a
// tab (0x09) which becomes a single space. All other printable/Unicode runes
// are preserved.
func sanitizeTerminal(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' {
			b.WriteByte(' ')
			continue
		}
		if r <= 0x1F || r == 0x7F || (r >= 0x80 && r <= 0x9F) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// sanitizeMarkdown neutralizes markdown/HTML injection. It first strips
// control characters (collapsing newlines/CR so a value cannot break out of a
// table row), then HTML-escapes, then escapes pipe characters.
func sanitizeMarkdown(s string) string {
	s = sanitizeTerminal(s)
	s = html.EscapeString(s)
	return strings.ReplaceAll(s, "|", "\\|")
}
