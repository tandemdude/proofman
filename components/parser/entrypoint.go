package parser

import (
	"github.com/tandemdude/proofman/components/parser/structure"
	"io"
	"regexp"
)

func ParseRootFile(reader io.Reader) (*structure.RootStructure, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	lexer := NewLexer(string(content))
	tokens, err := lexer.Split()
	if err != nil {
		return nil, err
	}

	return NewRootParser(tokens).Parse()
}

var theoryBlock = regexp.MustCompile(`(?msU)^theory(.+)begin`)

func ParseTheoryFile(reader io.Reader) ([]*structure.TheoryStructure, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	stringContent := string(content)
	parsedQueries := theoryBlock.FindAllString(stringContent, -1)

	output := make([]*structure.TheoryStructure, 0)
	for _, query := range parsedQueries {
		lexer := NewLexer(query)
		tokens, err := lexer.Split()
		if err != nil {
			return nil, err
		}

		parsed, err := NewTheoryParser(tokens).Parse()
		if err != nil {
			return nil, err
		}

		output = append(output, parsed)
	}

	return output, nil
}
