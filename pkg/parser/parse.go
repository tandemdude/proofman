package parser

import (
	"fmt"
	tks "github.com/tandemdude/proofman/pkg/parser/tokens"
	"slices"
	"strings"
)

type Parser struct {
	tokens       []*tks.Token
	currentIndex int
}

/**********
 * HELPERS
 **********/

func (p *Parser) current() *tks.Token {
	if p.currentIndex >= len(p.tokens) {
		return nil
	}

	return p.tokens[p.currentIndex]
}

func (p *Parser) peek(offset int) *tks.Token {
	if p.currentIndex+offset >= len(p.tokens) {
		return nil
	}

	return p.tokens[p.currentIndex+offset]
}

func (p *Parser) peekNext() *tks.Token {
	return p.peek(1)
}

func (p *Parser) eat(tokenTypes ...tks.TokenType) (*tks.Token, error) {
	reprs := make([]string, 0)
	for _, tokenType := range tokenTypes {
		reprs = append(reprs, tks.TokenTypeName[tokenType])
	}

	next := p.current()
	if next == nil {
		return nil, fmt.Errorf("Parsing failed at EOF\nExpected: %s\nGot: EOF", strings.Join(reprs, ", "))
	}

	for _, tokenType := range tokenTypes {
		if next.Type == tokenType {
			p.currentIndex++
			return next, nil
		}
	}

	return nil, fmt.Errorf("Parsing failed at L%d\nExpected: %s\nGot: %s",
		next.LineNo, strings.Join(reprs, ", "), tks.TokenTypeName[next.Type],
	)
}

func (p *Parser) eatKeyword(keywords ...string) (*tks.Token, error) {
	identifier, err := p.eat(tks.Identifier)
	if err != nil {
		return nil, err
	}

	for _, keyword := range keywords {
		if identifier.Value == keyword {
			return identifier, nil
		}
	}

	p.currentIndex--
	return nil, fmt.Errorf("Parsing failed at L%d\nExpected: %s\nGot: %s",
		identifier.LineNo, strings.Join(keywords, ", "), identifier.Value,
	)
}

func (p *Parser) stringArray(terminators []string) ([]string, error) {
	found := make([]string, 0)

	next := p.current()
	for next != nil && (next.Type != tks.Identifier || !slices.Contains(terminators, next.Value)) {
		session, err := p.eat(tks.Identifier, tks.StringLiteral)
		if err != nil {
			return nil, err
		}
		found = append(found, session.Value)

		next = p.current()
	}

	return found, nil
}

func (p *Parser) maybeQualifiedStringArray(kw string, terminators []string) ([]string, error) {
	_, err := p.eatKeyword(kw)
	if err != nil {
		return nil, nil
	}

	return p.stringArray(terminators)
}
