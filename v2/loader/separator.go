//
// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// NOTE: This code taken from https://github.com/cloudspannerecosystem/spanner-cli/blob/5eebf0a802df2a02c47776dc6aa52f59600e0b5e/separator.go
package loader

import (
	"strings"
)

type delimiter int

const (
	delimiterUndefined delimiter = iota
	delimiterHorizontal
	delimiterVertical
)

type inputStatement struct {
	statement string
	delim     delimiter
}

func (d delimiter) String() string {
	switch d {
	case delimiterUndefined:
		return ""
	case delimiterHorizontal:
		return ";"
	case delimiterVertical:
		return `\G`
	}
	return ""
}

func separateInput(input string) []inputStatement {
	return newSeparator(input).separate()
}

type separator struct {
	str []rune // remaining input
	sb  *strings.Builder
}

func newSeparator(s string) *separator {
	return &separator{
		str: []rune(s),
		sb:  &strings.Builder{},
	}
}

func (s *separator) consumeRawString() {
	// consume 'r' or 'R'
	s.sb.WriteRune(s.str[0])
	s.str = s.str[1:]

	delim := s.consumeStringDelimiter()
	s.consumeStringContent(delim, true)
}

func (s *separator) consumeBytesString() {
	// consume 'b' or 'B'
	s.sb.WriteRune(s.str[0])
	s.str = s.str[1:]

	delim := s.consumeStringDelimiter()
	s.consumeStringContent(delim, false)
}

func (s *separator) consumeRawBytesString() {
	// consume 'rb', 'Rb', 'rB', or 'RB'
	s.sb.WriteRune(s.str[0])
	s.sb.WriteRune(s.str[1])
	s.str = s.str[2:]

	delim := s.consumeStringDelimiter()
	s.consumeStringContent(delim, true)
}

func (s *separator) consumeString() {
	delim := s.consumeStringDelimiter()
	s.consumeStringContent(delim, false)
}

func (s *separator) consumeStringContent(delim string, raw bool) {
	var i int
	for i < len(s.str) {
		// check end of string
		switch {
		// check single-quoted delim
		case len(delim) == 1 && string(s.str[i]) == delim:
			s.str = s.str[i+1:]
			s.sb.WriteString(delim)
			return
		// check triple-quoted delim
		case len(delim) == 3 && len(s.str) >= i+3 && string(s.str[i:i+3]) == delim:
			s.str = s.str[i+3:]
			s.sb.WriteString(delim)
			return
		}

		// escape sequence
		if s.str[i] == '\\' {
			if raw {
				// raw string treats escape character as backslash
				s.sb.WriteRune('\\')
				i++
				continue
			}

			// invalid escape sequence
			if i+1 >= len(s.str) {
				s.sb.WriteRune('\\')
				return
			}

			s.sb.WriteRune('\\')
			s.sb.WriteRune(s.str[i+1])
			i += 2
			continue
		}
		s.sb.WriteRune(s.str[i])
		i++
	}
	s.str = s.str[i:]
	return
}

func (s *separator) consumeStringDelimiter() string {
	c := s.str[0]
	// check triple-quoted delim
	if len(s.str) >= 3 && s.str[1] == c && s.str[2] == c {
		delim := strings.Repeat(string(c), 3)
		s.sb.WriteString(delim)
		s.str = s.str[3:]
		return delim
	}
	s.str = s.str[1:]
	s.sb.WriteRune(c)
	return string(c)
}

func (s *separator) skipComments() {
	var i int
	for i < len(s.str) {
		var terminate string
		if s.str[i] == '#' {
			// single line comment "#"
			terminate = "\n"
			i++
		} else if i+1 < len(s.str) && s.str[i] == '-' && s.str[i+1] == '-' {
			// single line comment "--"
			terminate = "\n"
			i += 2
		} else if i+1 < len(s.str) && s.str[i] == '/' && s.str[i+1] == '*' {
			// multi line comments "/* */"
			// NOTE: Nested multiline comments are not supported in Spanner.
			// https://cloud.google.com/spanner/docs/lexical#multiline_comments
			terminate = "*/"
			i += 2
		}

		// no comment found
		if terminate == "" {
			return
		}

		// not terminated, but end of string
		if i >= len(s.str) {
			s.str = s.str[len(s.str):]
			return
		}

		for ; i < len(s.str); i++ {
			if l := len(terminate); l == 1 {
				if string(s.str[i]) == terminate {
					s.str = s.str[i+1:]
					i = 0
					break
				}
			} else if l == 2 {
				if i+1 < len(s.str) && string(s.str[i:i+2]) == terminate {
					s.str = s.str[i+2:]
					i = 0
					break
				}
			}
		}

		// not terminated, but end of string
		if i >= len(s.str) {
			s.str = s.str[len(s.str):]
			return
		}
	}
}

// separate separates input string into multiple Spanner statements.
// This does not validate syntax of statements.
//
// NOTE: Logic for parsing a statement is mostly taken from spansql.
// https://github.com/googleapis/google-cloud-go/blob/master/spanner/spansql/parser.go
func (s *separator) separate() []inputStatement {
	var statements []inputStatement
	for len(s.str) > 0 {
		s.skipComments()
		if len(s.str) == 0 {
			break
		}

		switch s.str[0] {
		// possibly string literal
		case '"', '\'', 'r', 'R', 'b', 'B':
			// valid string prefix: "b", "B", "r", "R", "br", "bR", "Br", "BR"
			// https://cloud.google.com/spanner/docs/lexical#string_and_bytes_literals
			raw, bytes, str := false, false, false
			for i := 0; i < 3 && i < len(s.str); i++ {
				switch {
				case !raw && (s.str[i] == 'r' || s.str[i] == 'R'):
					raw = true
					continue
				case !bytes && (s.str[i] == 'b' || s.str[i] == 'B'):
					bytes = true
					continue
				case s.str[i] == '"' || s.str[i] == '\'':
					str = true
					switch {
					case raw && bytes:
						s.consumeRawBytesString()
					case raw:
						s.consumeRawString()
					case bytes:
						s.consumeBytesString()
					default:
						s.consumeString()
					}
				}
				break
			}
			if !str {
				s.sb.WriteRune(s.str[0])
				s.str = s.str[1:]
			}
		// quoted identifier
		case '`':
			s.sb.WriteRune(s.str[0])
			s.str = s.str[1:]
			s.consumeStringContent("`", false)
		// horizontal delim
		case ';':
			statements = append(statements, inputStatement{
				statement: strings.TrimSpace(s.sb.String()),
				delim:     delimiterHorizontal,
			})
			s.sb.Reset()
			s.str = s.str[1:]
		// possibly vertical delim
		case '\\':
			if len(s.str) >= 2 && s.str[1] == 'G' {
				statements = append(statements, inputStatement{
					statement: strings.TrimSpace(s.sb.String()),
					delim:     delimiterVertical,
				})
				s.sb.Reset()
				s.str = s.str[2:]
				continue
			}
			s.sb.WriteRune(s.str[0])
			s.str = s.str[1:]
		default:
			s.sb.WriteRune(s.str[0])
			s.str = s.str[1:]
		}
	}

	// flush remained
	if s.sb.Len() > 0 {
		if str := strings.TrimSpace(s.sb.String()); len(str) > 0 {
			statements = append(statements, inputStatement{
				statement: str,
				delim:     delimiterUndefined,
			})
			s.sb.Reset()
		}
	}
	return statements
}
