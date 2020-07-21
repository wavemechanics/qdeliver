package token

import (
	"strings"

	"github.com/wavemechanics/etype"
)

const (
	ErrEscape = etype.Sentinel("unterminated escape")
	ErrSquote = etype.Sentinel("unterminated single quote")
	ErrDquote = etype.Sentinel("unterminated double quote")
)

type splitter struct {
	src    string   // string we are splitting
	next   int      // next char in string
	tok    string   // token we are building up
	tokens []string // tokens accumulated so far
}

// SplitFile splits a string into lines delimited by \r, \r\n, or \n
//
func SplitFile(contents string) []string {
	contents = strings.ReplaceAll(contents, "\r\n", "\n")
	contents = strings.ReplaceAll(contents, "\r", "\n")
	return strings.Split(contents, "\n")
}

func SplitLine(line string) ([]string, error) {
	s := &splitter{
		src: line,
	}
	if err := s.split(); err != nil {
		return nil, err
	}
	return s.tokens, nil
}

func (s *splitter) split() error {
	for s.next < len(s.src) {
		var err error
		switch s.src[s.next] {
		case ' ', '\t':
			if s.tok != "" {
				s.tokens = append(s.tokens, s.tok)
				s.tok = ""
			}
			s.next++
		case '#':
			return nil
		case '\\':
			err = s.escape()
		case '"':
			err = s.dquote()
		case '\'':
			err = s.squote()
		default:
			s.chunk()
		}
		if err != nil {
			return err
		}
	}
	if s.tok != "" {
		s.tokens = append(s.tokens, s.tok)
		s.tok = ""
	}
	return nil
}

func (s *splitter) chunk() {
	start := s.next // remember first position in chunk
	var c byte
	for s.next < len(s.src) {
		c = s.src[s.next]
		if c == ' ' || c == '\t' || c == '\\' || c == '"' || c == '\'' {
			break
		}
		s.next++
	}
	s.save(start)
}

func (s *splitter) escape() error {
	s.next++ // skip the '\'
	if s.next >= len(s.src) {
		return ErrEscape
	}
	s.next++
	s.save(s.next - 1)
	return nil
}

func (s *splitter) squote() error {
	s.next++ // skip the initial "'"
	start := s.next
	for s.next < len(s.src) && s.src[s.next] != '\'' {
		s.next++
	}
	if s.next >= len(s.src) {
		return ErrSquote
	}
	s.save(start)
	s.next++ // skip final "'"
	return nil
}

func (s *splitter) dquote() error {
	s.next++ // skip the initial '"'
	start := s.next
	var c byte
	var err error
loop:
	for s.next < len(s.src) {
		c = s.src[s.next]
		switch c {
		case '"':
			break loop
		case '\\':
			if s.next > start {
				s.save(start)
			}
			if err = s.escape(); err != nil {
				break loop
			}
			start = s.next
		default:
			s.next++
		}
	}
	if err != nil {
		return err
	}
	if c != '"' {
		return ErrDquote
	}
	s.save(start)
	s.next++ // skip final "'"
	return nil
}

func (s *splitter) save(start int) {
	s.tok += s.src[start:s.next]
}
