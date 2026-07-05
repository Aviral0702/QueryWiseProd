package report

import "testing"

func TestSanitizeTerminal(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello world", "hello world"},
		{"esc color", "\x1b[31mred\x1b[0m", "[31mred[0m"},
		{"osc title", "\x1b]0;pwned\x07", "]0;pwned"},
		{"carriage return", "real\rfake", "realfake"},
		{"newline", "line1\nline2", "line1line2"},
		{"bell", "ding\a", "ding"},
		{"backspace", "ab\bc", "abc"},
		{"del", "ab\x7fc", "abc"},
		{"c1 control", "abc", "abc"},
		{"tab to space", "a\tb", "a b"},
		{"unicode kept", "café — 日本語", "café — 日本語"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeTerminal(tc.in)
			if got != tc.want {
				t.Errorf("sanitizeTerminal(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSanitizeMarkdown(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "add an index", "add an index"},
		{"pipe escaped", "a | b", "a \\| b"},
		{
			"xss payload",
			"safe\n<img src=x onerror=alert(1)>|end",
			"safe&lt;img src=x onerror=alert(1)&gt;\\|end",
		},
		{"ampersand", "a & b", "a &amp; b"},
		{"quotes", `say "hi" 'yo'`, "say &#34;hi&#34; &#39;yo&#39;"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeMarkdown(tc.in)
			if got != tc.want {
				t.Errorf("sanitizeMarkdown(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
