package parser

import (
	"fmt"
	"github.com/tandemdude/proofman/pkg/parser/structure"
	tks "github.com/tandemdude/proofman/pkg/parser/tokens"
	"slices"
	"strings"
)

const (
	ChapterDefinition = "chapter_definition"
	Chapter           = "chapter"
	Session           = "session"
	Description       = "description"
	Directories       = "directories"
	Options           = "options"
	Sessions          = "sessions"
	Theories          = "theories"
	Global            = "global"
	DocumentTheories  = "document_theories"
	DocumentFiles     = "document_files"
	ExportFiles       = "export_files"
	ExportClasspath   = "export_classpath"
	In                = "in"
)

var allkeywords = []string{
	ChapterDefinition,
	Chapter,
	Session,
	Description,
	Directories,
	Options,
	Sessions,
	Theories,
	Global,
	DocumentTheories,
	DocumentFiles,
	ExportFiles,
	ExportClasspath,
	In,
}

type RootParser struct {
	Parser
	currentChapter string
}

/****************
 * PARSE METHODS
 ****************/

// maybeDescription attempts to parse a description from the parse position.
//
// A description construct matches the following expression: "description" (Identifier | StringLiteral)
//
// Returns nil when the parse position does not represent a description. If a token is returned, the
// parse position will have been advanced past both the 'description' keyword and the value.
func (p *RootParser) maybeDescription() (*tks.Token, error) {
	_, err := p.eatKeyword(Description)
	if err != nil {
		// This does not represent a description
		return nil, nil
	}

	return p.eat(tks.Identifier, tks.StringLiteral)
}

// maybeGroups attempts to parse a groups construct from the parse position.
//
// A groups construct matches the following expression: "(" (Identifier | StringLiteral)* ")"
//
// Returns nil when the parse position does not represent a groups construct. If the returned slice is not nil, the
// parse position will have been advanced past the closing right paren token.
func (p *RootParser) maybeGroups() ([]*tks.Token, error) {
	_, err := p.eat(tks.LeftParen)
	if err != nil {
		// This can't be the start of a groups section
		return nil, nil
	}

	items := make([]*tks.Token, 0)

	item := p.current()
	for item.Type != tks.RightParen {
		if item.Type != tks.Identifier && item.Type != tks.StringLiteral {
			return nil, fmt.Errorf(
				"Parsing failed at L%d\nExpected: Identifier, StringLiteral\nGot: %s",
				item.LineNo, tks.TokenTypeName[item.Type],
			)
		}

		items = append(items, item)
		p.currentIndex++
		item = p.current()

		if item == nil {
			return nil, fmt.Errorf("Parsing failed at EOF\nExpected: ')'\nGot: EOF")
		}
	}

	// This is guaranteed to be a right paren - we checked in the loop above
	_, _ = p.eat(tks.RightParen)
	return items, nil
}

// chapterDefinition parses a chapter structure from the parse position, including values only permitted
// by a 'chapter_def' construct (groups and description). This assumes that the 'chapter_definition' keyword token
// has already been eaten - and so expects the token at the parse position to be the name of the chapter. Further
// tokens will be interpreted as the group or description constructs through calling maybeGroups and maybeDescription.
func (p *RootParser) chapterDefinition() (*structure.Chapter, error) {
	name, err := p.eat(tks.Identifier, tks.StringLiteral)
	if err != nil {
		return nil, err
	}

	chapter := &structure.Chapter{
		Name:        name.Value,
		Groups:      make([]string, 0),
		Description: "",
		Sessions:    make([]*structure.Session, 0),
	}

	groups, err := p.maybeGroups()
	if err != nil {
		return nil, err
	}
	if groups != nil {
		for _, group := range groups {
			chapter.Groups = append(chapter.Groups, group.Value)
		}
	}

	description, err := p.maybeDescription()
	if err != nil {
		return nil, err
	}
	if description != nil {
		chapter.Description = description.Value
	}

	return chapter, nil
}

// chapter parses a chapter structure from the parse position. This assumes that the 'chapter' keyword token
// has already been eaten - and so expects the token at the parse position to be the name of the chapter.
func (p *RootParser) chapter() (*structure.Chapter, error) {
	name, err := p.eat(tks.Identifier, tks.StringLiteral)
	if err != nil {
		return nil, err
	}

	chapter := &structure.Chapter{
		Name:        name.Value,
		Groups:      make([]string, 0),
		Description: "",
		Sessions:    make([]*structure.Session, 0),
	}

	return chapter, nil
}

func (p *RootParser) maybeDir() (*tks.Token, error) {
	_, err := p.eatKeyword(In)
	if err != nil {
		return nil, nil
	}

	dirName, err := p.eat(tks.Identifier, tks.StringLiteral)
	if err != nil {
		return nil, err
	}

	return dirName, nil
}

func (p *RootParser) maybeOptionsMap() (map[string]*tks.Token, error) {
	_, err := p.eat(tks.LeftSquareParen)
	if err != nil {
		return nil, nil
	}

	optionsMap := make(map[string]*tks.Token)
	current := p.current()
	for current.Type != tks.RightSquareParen {
		optName, err := p.eat(tks.Identifier, tks.StringLiteral)
		if err != nil {
			return nil, err
		}

		var optValue *tks.Token = nil
		if p.current().Type == tks.Equal {
			_, _ = p.eat(tks.Equal)
			optValue, err = p.eat(tks.Identifier, tks.StringLiteral, tks.NumberLiteral)
			if err != nil {
				return nil, err
			}
		}

		optionsMap[optName.Value] = optValue

		current = p.current()
		if current == nil {
			return nil, fmt.Errorf("Parsing failed at EOF\nExpected: ',', ']'\nGot: EOF")
		}

		// next token HAS to be a comma or a right square bracket
		// any other token type is a grammar error
		if current.Type == tks.RightSquareParen {
			continue
		}

		_, err = p.eat(tks.Comma)
		if err != nil {
			return nil, err
		}

		current = p.current()
	}

	_, _ = p.eat(tks.RightSquareParen)

	return optionsMap, nil
}

func (p *RootParser) maybeOptions() (map[string]*tks.Token, error) {
	optionsKw, err := p.eatKeyword(Options)
	if err != nil {
		return nil, nil
	}

	optionsMap, err := p.maybeOptionsMap()
	if err != nil {
		return nil, err
	}
	if optionsMap == nil {
		return nil, fmt.Errorf("Parsing failed at L%d\nExpected: option name", optionsKw.LineNo)
	}

	return optionsMap, nil
}

func (p *RootParser) maybeTheories() (*structure.Theories, error) {
	_, err := p.eatKeyword(Theories)
	if err != nil {
		return nil, nil
	}

	theories := &structure.Theories{
		Options:  make(map[string]*tks.Token),
		Entries:  make([]string, 0),
		IsGlobal: make(map[string]bool),
	}

	// check for the options section
	optionsMap, err := p.maybeOptionsMap()
	if err != nil {
		return nil, err
	}
	if optionsMap != nil {
		theories.Options = optionsMap
	}

	// parse theory entries
	next := p.current()
	for next != nil && (next.Type != tks.Identifier || !slices.Contains(allkeywords, next.Value)) {
		entry, err := p.eat(tks.Identifier, tks.StringLiteral)
		if err != nil {
			return nil, err
		}

		// check if there is a global qualifier
		global := false
		if current := p.current(); current != nil && current.Type == tks.LeftParen {
			_, _ = p.eat(tks.LeftParen)
			_, err := p.eatKeyword(Global)
			if err != nil {
				return nil, err
			}
			global = true

			_, err = p.eat(tks.RightParen)
			if err != nil {
				return nil, err
			}
		}

		theories.Entries = append(theories.Entries, entry.Value)
		theories.IsGlobal[entry.Value] = global

		next = p.current()
	}

	return theories, nil
}

func (p *RootParser) maybeDocumentFiles() (*structure.DocumentFiles, error) {
	_, err := p.eatKeyword(DocumentFiles)
	if err != nil {
		return nil, nil
	}

	documentFiles := &structure.DocumentFiles{}

	// check for a directory qualifier
	if next := p.current(); next != nil && next.Type == tks.LeftParen {
		_, _ = p.eat(tks.LeftParen)

		dir, err := p.maybeDir()
		if err != nil {
			return nil, err
		}
		if dir == nil {
			return nil, fmt.Errorf("FIXME")
		}
		documentFiles.Dir = dir.Value

		_, err = p.eat(tks.RightParen)
		if err != nil {
			return nil, err
		}
	}

	// get the actual document file entries
	entries, err := p.stringArray(allkeywords)
	if err != nil {
		return nil, err
	}
	documentFiles.Entries = entries

	return documentFiles, nil
}

func (p *RootParser) maybeExportFiles() (*structure.ExportFiles, error) {
	_, err := p.eatKeyword(ExportFiles)
	if err != nil {
		return nil, nil
	}

	exportFiles := &structure.ExportFiles{}

	// check for a directory qualifier
	if next := p.current(); next != nil && next.Type == tks.LeftParen {
		_, _ = p.eat(tks.LeftParen)

		dir, err := p.maybeDir()
		if err != nil {
			return nil, err
		}
		if dir == nil {
			return nil, fmt.Errorf("FIXME")
		}
		exportFiles.Dir = dir.Value

		_, err = p.eat(tks.RightParen)
		if err != nil {
			return nil, err
		}
	}

	// check for nat qualifier
	if next := p.current(); next != nil && next.Type == tks.LeftSquareParen {
		_, _ = p.eat(tks.LeftSquareParen)

		nat, err := p.eat(tks.NumberLiteral)
		if err != nil {
			return nil, err
		}

		if strings.Contains(nat.Value, ".") {
			return nil, fmt.Errorf("only natural numbers permitted")
		}

		exportFiles.Nat = nat.Value

		_, err = p.eat(tks.RightSquareParen)
		if err != nil {
			return nil, err
		}
	}

	entries, err := p.stringArray(allkeywords)
	if err != nil {
		return nil, err
	}
	exportFiles.Entries = entries

	return exportFiles, nil
}

// session syntax
// "session" name groups? (in dir)? = (parent +) ...
func (p *RootParser) session() (*structure.Session, error) {
	name, err := p.eat(tks.Identifier, tks.StringLiteral)
	if err != nil {
		return nil, err
	}

	session := &structure.Session{
		Name:             name.Value,
		Groups:           make([]string, 0),
		Dir:              "",
		SystemName:       "",
		Description:      "",
		Options:          nil,
		Sessions:         nil,
		Directories:      nil,
		Theories:         nil,
		DocumentTheories: nil,
		DocumentFiles:    nil,
		ExportFiles:      nil,
		ExportClasspath:  nil,
	}

	// parse groups section
	groups, err := p.maybeGroups()
	if err != nil {
		return nil, err
	}
	if groups != nil {
		for _, group := range groups {
			session.Groups = append(session.Groups, group.Value)
		}
	}

	// Parse dir section
	dir, err := p.maybeDir()
	if err != nil {
		return nil, err
	}
	if dir != nil {
		session.Dir = dir.Value
	}

	// An equal sign is required for a valid session definition, everything past the equal sign is optional
	_, err = p.eat(tks.Equal)
	if err != nil {
		return nil, err
	}

	// Check for "system_name +" construct
	if maybePlus := p.peekNext(); maybePlus != nil && maybePlus.Type == tks.Plus {
		systemName, err := p.eat(tks.Identifier, tks.StringLiteral)
		if err != nil {
			return nil, err
		}
		_, _ = p.eat(tks.Plus)

		session.SystemName = systemName.Value
	}

	// description is optional
	description, err := p.maybeDescription()
	if err != nil {
		return nil, err
	}
	if description != nil {
		session.Description = description.Value
	}

	// If the session has options, they must be the first block within the definition - it is not valid
	// to have options after a "theories" block for example
	options, err := p.maybeOptions()
	if err != nil {
		return nil, err
	}
	if options != nil {
		session.Options = options
	}

	// Parse optional "sessions" section
	sessions, err := p.maybeQualifiedStringArray(Sessions, allkeywords)
	if err != nil {
		return nil, err
	}
	if sessions != nil {
		session.Sessions = sessions
	}

	// Parse optional "directories" section
	directories, err := p.maybeQualifiedStringArray(Directories, allkeywords)
	if err != nil {
		return nil, err
	}
	if directories != nil {
		session.Directories = directories
	}

	// Parse repeated "theories" sections
	allTheories := make([]*structure.Theories, 0)
	theories, err := p.maybeTheories()
	for theories != nil && err == nil {
		allTheories = append(allTheories, theories)
		// check for an additional theories section
		theories, err = p.maybeTheories()
	}
	if err != nil {
		return nil, err
	}
	session.Theories = allTheories

	// Parse optional "document_theories" section
	documentTheories, err := p.maybeQualifiedStringArray(DocumentTheories, allkeywords)
	if err != nil {
		return nil, err
	}
	if documentTheories != nil {
		session.DocumentTheories = documentTheories
	}

	// Parse repeated "document_files" section
	allDocumentFiles := make([]*structure.DocumentFiles, 0)
	documentFiles, err := p.maybeDocumentFiles()
	for documentFiles != nil && err == nil {
		allDocumentFiles = append(allDocumentFiles, documentFiles)
		// check for an additional document files section
		documentFiles, err = p.maybeDocumentFiles()
	}
	if err != nil {
		return nil, err
	}
	session.DocumentFiles = allDocumentFiles

	// Parse repeated "export_files" section
	allExportFiles := make([]*structure.ExportFiles, 0)
	exportFiles, err := p.maybeExportFiles()
	for exportFiles != nil && err == nil {
		allExportFiles = append(allExportFiles, exportFiles)
		exportFiles, err = p.maybeExportFiles()
	}
	if err != nil {
		return nil, err
	}
	session.ExportFiles = allExportFiles

	// Parse optional "export_classpath" section
	exportClasspath, err := p.maybeQualifiedStringArray(ExportClasspath, allkeywords)
	if err != nil {
		return nil, err
	}
	if exportClasspath != nil {
		session.ExportClasspath = exportClasspath
	}

	return session, nil
}

/*************
 * ENTRYPOINT
 *************/

func NewRootParser(tokens []*tks.Token) *RootParser {
	// Strip out any comments because semantically they are unimportant, and add extra
	// annoyance when implementing the grammar Parser.
	newTokens := make([]*tks.Token, 0)
	for _, tk := range tokens {
		if tk.Type == tks.Comment {
			continue
		}

		newTokens = append(newTokens, tk)
	}

	return &RootParser{
		Parser:         Parser{newTokens, 0},
		currentChapter: "Unsorted",
	}
}

func (p *RootParser) Parse() (*structure.RootStructure, error) {
	parsedStructure := structure.RootStructure{
		Chapters:     make(map[string]*structure.Chapter),
		ChapterOrder: make([]string, 0),
	}

	// The default chapter is unsorted, so we need to add it - but there may be no
	// sessions registered to that chapter
	parsedStructure.Chapters["Unsorted"] = &structure.Chapter{
		Name:        "Unsorted",
		Groups:      make([]string, 0),
		Description: "",
		Sessions:    make([]*structure.Session, 0),
	}
	parsedStructure.ChapterOrder = append(parsedStructure.ChapterOrder, "Unsorted")

	for p.currentIndex < len(p.tokens) {
		currentToken, err := p.eatKeyword(ChapterDefinition, Chapter, Session)
		if err != nil {
			return nil, err
		}

		switch currentToken.Value {
		case ChapterDefinition:
			chapter, err := p.chapterDefinition()
			if err != nil {
				return nil, err
			}

			if existing, ok := parsedStructure.Chapters[chapter.Name]; ok {
				existing.Groups = chapter.Groups
				existing.Description = chapter.Description
			} else {
				parsedStructure.Chapters[chapter.Name] = chapter
			}

		case Chapter:
			chapter, err := p.chapter()
			if err != nil {
				return nil, err
			}

			if _, ok := parsedStructure.Chapters[chapter.Name]; !ok {
				parsedStructure.Chapters[chapter.Name] = chapter
			}

			if !slices.Contains(parsedStructure.ChapterOrder, chapter.Name) {
				parsedStructure.ChapterOrder = append(parsedStructure.ChapterOrder, chapter.Name)
			}

			p.currentChapter = chapter.Name

		case Session:
			session, err := p.session()
			if err != nil {
				return nil, err
			}

			parsedStructure.Chapters[p.currentChapter].Sessions = append(parsedStructure.Chapters[p.currentChapter].Sessions, session)

		default:
			panic("unexpected state encountered")
		}
	}

	return &parsedStructure, nil
}
