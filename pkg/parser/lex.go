package parser

import (
	"errors"
	"fmt"
	tk "github.com/tandemdude/proofman/pkg/parser/tokens"
	"strconv"
	"strings"
	"unicode"
)

var runeTokenMap = map[rune]tk.TokenType{
	'=': tk.Equal,
	'+': tk.Plus,
	'(': tk.LeftParen,
	')': tk.RightParen,
	'[': tk.LeftSquareParen,
	']': tk.RightSquareParen,
	',': tk.Comma,
	'#': tk.Hash,
	'*': tk.Asterisk,
}

type Lexer struct {
	source string
}

func NewLexer(source string) *Lexer {
	return &Lexer{source: source}
}

func parseIdentifier(s *string, idx int) (string, error) {
	start, dotFound := idx, false
	for idx < len(*s) {
		r := rune((*s)[idx])
		// An identifier *may* contain a single dot, but it may not be the first or last element
		if idx == start && r == '.' {
			return "", errors.New("identifiers may not start with '.'")
		}

		if r == '.' {
			if dotFound {
				return "", errors.New("identifiers may only contain a single '.'")
			}
			dotFound = true

			idx++
			continue
		}

		// Valid characters are letters, digits, or underscores
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			break
		}

		idx++
	}

	if (*s)[idx-1] == '.' {
		return "", errors.New("identifiers may not end with '.'")
	}

	return (*s)[start:idx], nil
}

func parseStringLiteral(s *string, idx int) (string, error) {
	start := idx + 1

	idx++ // Skip the opening quotation mark

	escaped := false // Track whether the previous character was an escape character
	for idx < len(*s) {
		r := (*s)[idx]

		if escaped {
			// If we are in an escape sequence, reset the escaped flag
			escaped = false
		} else if r == '\\' {
			// If we encounter a backslash, set the escaped flag
			escaped = true
		} else if r == '"' {
			// Check for termination of the string literal
			return (*s)[start:idx], nil
		}

		idx++
	}

	return "", errors.New("unterminated string literal")
}

func parseBracedStringLiteral(s *string, idx int) (string, error) {
	start := idx + 2

	for idx < len(*s) {
		if (*s)[idx:idx+2] == "*}" {
			return (*s)[start:idx], nil
		}

		idx++
	}

	return "", errors.New("unterminated braced string literal")
}

func parseLatexStringLiteral(s *string, idx int) (string, error) {
	start := idx

	opens := 0

	for idx < len(*s) {
		r := (*s)[idx]

		if r == '\\' {
			if (*s)[idx:idx+8] == `\<close>` {
				opens--

				if opens == 0 {
					return (*s)[start : idx+8], nil
				}
			} else if (*s)[idx:idx+7] == `\<open>` {
				opens++
			}
		}

		idx++
	}

	return "", errors.New("unterminated latex string literal")
}

func parseNumberLiteral(s *string, idx int) (string, error) {
	start, decimalPointEncountered := idx, false

	for idx < len(*s) {
		r := rune((*s)[idx])

		if !unicode.IsDigit(r) && !(r == '.' && !decimalPointEncountered) {
			break
		}

		if r == '.' {
			decimalPointEncountered = true
		}

		idx++
	}

	found := (*s)[start:idx]
	if found[len(found)-1] == '.' {
		return "", errors.New("number literal cannot end with '.'")
	}

	return (*s)[start:idx], nil
}

func parseComment(s *string, idx int) (string, error) {
	start := idx + 2

	nestingLevel := 1 // Start with one open comment
	idx += 2          // Move past the starting sequence '(*'

	for idx < len(*s) {
		if idx+2 > len(*s) {
			break
		}

		if (*s)[idx:idx+2] == "*)" {
			nestingLevel-- // Found a closing comment
			if nestingLevel == 0 {
				return (*s)[start:idx], nil // Return the full comment
			}
			idx += 2 // Move past the closing sequence '*)'
			continue
		}

		if (*s)[idx:idx+2] == "(*" {
			nestingLevel++ // Found a new opening comment
			idx += 2       // Move past the opening sequence '(*'
			continue
		}

		// Check for escape sequences (e.g., *\)
		if idx < len(*s) && (*s)[idx] == '\\' {
			idx++ // Skip the escape character
		}

		idx++ // Move to the next character
	}

	return "", errors.New("unterminated comment")
}

func (l *Lexer) Split() ([]*tk.Token, error) {
	tokens := make([]*tk.Token, 0)

	currentIndex, lineNo := 0, 0
	for currentIndex < len(l.source) {
		currentRune := rune(l.source[currentIndex])

		// Ignore whitespace
		if unicode.IsSpace(currentRune) {
			if currentRune == '\n' {
				lineNo++
			}

			currentIndex++
			continue
		}

		// Parse composite tokens
		switch {
		case unicode.IsLetter(currentRune):
			str, err := parseIdentifier(&l.source, currentIndex)
			if err != nil {
				return tokens, err
			}

			tokens = append(tokens, &tk.Token{tk.Identifier, str, lineNo})
			currentIndex += len(str)
		case currentRune == '"':
			str, err := parseStringLiteral(&l.source, currentIndex)
			if err != nil {
				return tokens, err
			}

			tokens = append(tokens, &tk.Token{tk.StringLiteral, strings.TrimSpace(str), lineNo})
			currentIndex += len(str) + 2

			lineNo += strings.Count(str, "\n")
		case currentRune == '{' && currentIndex+1 < len(l.source) && l.source[currentIndex+1] == '*':
			str, err := parseBracedStringLiteral(&l.source, currentIndex)
			if err != nil {
				return tokens, err
			}

			tokens = append(tokens, &tk.Token{tk.StringLiteral, strings.TrimSpace(str), lineNo})
			currentIndex += len(str) + 4

			lineNo += strings.Count(str, "\n")
		case currentRune == '\\':
			str, err := parseLatexStringLiteral(&l.source, currentIndex)
			if err != nil {
				return tokens, err
			}

			// TODO - consider adding a token info flag mentioning this is latex syntax
			// if the string starts with `\<comment>` then this is a comment instead of a string literal
			if strings.HasPrefix(str, `\<comment>`) {
				tokens = append(tokens, &tk.Token{tk.Comment, str, lineNo})
			} else {
				tokens = append(tokens, &tk.Token{tk.StringLiteral, str, lineNo})
			}
			currentIndex += len(str)

			lineNo += strings.Count(str, "\n")
		case unicode.IsDigit(currentRune):
			str, err := parseNumberLiteral(&l.source, currentIndex)
			if err != nil {
				return tokens, err
			}

			tokens = append(tokens, &tk.Token{tk.NumberLiteral, str, lineNo})
			currentIndex += len(str)
		case currentRune == '(' && currentIndex+1 < len(l.source) && l.source[currentIndex+1] == '*':
			str, err := parseComment(&l.source, currentIndex)
			if err != nil {
				return tokens, err
			}

			tokens = append(tokens, &tk.Token{tk.Comment, strings.TrimSpace(str), lineNo})
			currentIndex += len(str) + 4

			lineNo += strings.Count(str, "\n")
		default:
			tokenType, ok := runeTokenMap[currentRune]
			if !ok {
				return tokens, fmt.Errorf("unknown token type for rune %s", strconv.QuoteRune(currentRune))
			}

			tokens = append(tokens, &tk.Token{tokenType, strconv.QuoteRune(currentRune), lineNo})
			currentIndex++
		}
	}

	return tokens, nil
}
