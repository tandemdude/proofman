package parser

import (
	"github.com/tandemdude/proofman/pkg/parser/structure"
	tks "github.com/tandemdude/proofman/pkg/parser/tokens"
)

const (
	Theory  = "theory"
	Imports = "imports"
	Begin   = "begin"
)

type TheoryParser struct {
	Parser
}

func NewTheoryParser(tokens []*tks.Token) *TheoryParser {
	newTokens := make([]*tks.Token, 0)
	for _, tk := range tokens {
		if tk.Type == tks.Comment {
			continue
		}

		newTokens = append(newTokens, tk)
	}

	return &TheoryParser{
		Parser: Parser{newTokens, 0},
	}
}

func (p *TheoryParser) Parse() (*structure.TheoryStructure, error) {
	// TODO - there are some theories in the isabelle source this won't parse
	//        maybe fix in the future
	parsedStructure := &structure.TheoryStructure{
		Name:    "",
		Imports: make([]string, 0),
	}

	_, err := p.eatKeyword(Theory)
	if err != nil {
		return nil, err
	}

	name, err := p.eat(tks.Identifier, tks.StringLiteral)
	if err != nil {
		return nil, err
	}
	parsedStructure.Name = name.Value

	imports, err := p.maybeQualifiedStringArray(Imports, []string{Begin})
	if err != nil {
		return nil, err
	}
	if len(imports) > 0 {
		parsedStructure.Imports = imports
	}

	_, err = p.eatKeyword(Begin)
	if err != nil {
		return nil, err
	}

	return parsedStructure, nil
}
