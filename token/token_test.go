package token_test

import (
	"testing"

	"github.com/wavemechanics/deliver/token"
)

func TestSplitFile(t *testing.T) {
	var tests = []struct {
		contents string
		lines    []string
	}{
		{"", []string{""}},
		{"no delim", []string{"no delim"}},
		{"\n", []string{"", ""}},
		{"\r", []string{"", ""}},
		{"\r\n", []string{"", ""}},
		{"line\n", []string{"line", ""}},
		{"line\r", []string{"line", ""}},
		{"line\r\n", []string{"line", ""}},
		{"line1\nline2", []string{"line1", "line2"}},
		{"line1\rline2", []string{"line1", "line2"}},
		{"line1\r\nline2", []string{"line1", "line2"}},
	}

	for _, test := range tests {
		lines := token.SplitFile(test.contents)
		if !tokensMatch(test.lines, lines) {
			t.Errorf("%q: %+v, want %+v", test.contents, lines, test.lines)
		}
	}
}
func TestSplitLine(t *testing.T) {
	var tests = []struct {
		line   string
		err    error
		tokens []string
	}{
		{"", nil, nil},

		// comment
		{`#`, nil, nil},

		// chunk tests
		{`a`, nil, []string{"a"}},
		{`a b`, nil, []string{"a", "b"}},

		// escape tests
		{`\`, token.ErrEscape, nil},
		{`\a`, nil, []string{"a"}},

		// concat chunks and escapes
		{`a\b`, nil, []string{"ab"}},
		{`\ab`, nil, []string{"ab"}},
		{`\a\b`, nil, []string{"ab"}},

		// squotes
		{`'`, token.ErrSquote, nil},
		{`'a'`, nil, []string{"a"}},
		{`'a' 'b'`, nil, []string{"a", "b"}},
		{`'a''b'`, nil, []string{"ab"}},
		{`'a\b'`, nil, []string{"a\\b"}},

		// concat chunks and squotes
		{`a'b'`, nil, []string{"ab"}},
		{`'a'b`, nil, []string{"ab"}},

		// concat escapes and squotes
		{`\a'b'`, nil, []string{"ab"}},
		{`'a'\b`, nil, []string{"ab"}},

		// concat chunks, escapes and squotes
		{`a\b'c'`, nil, []string{"abc"}},
		{`'a'\bc`, nil, []string{"abc"}},

		// dquotes
		{`"`, token.ErrDquote, nil},
		//{"\"\"", nil, []string{""}}, // empty strings need to be fixed
		{`"a"`, nil, []string{"a"}},
		{`"a""b"`, nil, []string{"ab"}},
		{`"a" "b"`, nil, []string{"a", "b"}},

		// dquotes with embedded escapes
		{`"\""`, nil, []string{`"`}},
		{`"\`, token.ErrEscape, nil},
		{`"before\-after"`, nil, []string{"before-after"}},
		{`"before\"middle\"after"`, nil, []string{`before"middle"after`}},

		// dquotes and chunks
		{`"a"b`, nil, []string{"ab"}},
		{`a"b"`, nil, []string{"ab"}},

		// dquotes and squotes
		{`"a"'b'`, nil, []string{"ab"}},
		{`'a'"b"`, nil, []string{"ab"}},
		{`'a'"\b"`, nil, []string{"ab"}},

		// dquotes and escapes
		{`"a"\b`, nil, []string{"ab"}},

		// sanity check
		{
			`command '"'"quoted string with actual quotes"'"' "\"another way\"" "\t" '\t'`,
			nil,
			[]string{
				"command",
				`"quoted string with actual quotes"`,
				`"another way"`,
				`t`,
				`\t`,
			},
		},
	}

	for _, test := range tests {

		tokens, err := token.SplitLine(test.line)
		if err != test.err {
			t.Errorf("Split: %q: %v, want %v", test.line, err, test.err)
			continue
		}
		if err != nil {
			continue
		}
		if !tokensMatch(tokens, test.tokens) {
			t.Errorf("Split: %q: %+v, want %+v", test.line, tokens, test.tokens)
		}
	}
}

func tokensMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, s := range a {
		if b[i] != s {
			return false
		}
	}
	return true
}
