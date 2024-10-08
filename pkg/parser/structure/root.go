package structure

import "github.com/tandemdude/proofman/pkg/parser/tokens"

type Theories struct {
	Options  map[string]*tokens.Token
	Entries  []string
	IsGlobal map[string]bool
}

type DocumentFiles struct {
	Dir     string
	Entries []string
}

type ExportFiles struct {
	Dir     string
	Nat     string
	Entries []string
}

type Session struct {
	Name   string
	Groups []string
	Dir    string

	SystemName  string
	Description string
	Options     map[string]*tokens.Token

	Sessions         []string
	Directories      []string
	Theories         []*Theories
	DocumentTheories []string
	DocumentFiles    []*DocumentFiles
	ExportFiles      []*ExportFiles
	ExportClasspath  []string
}

type Chapter struct {
	Name        string
	Groups      []string
	Description string

	Sessions []*Session
}

type RootStructure struct {
	Chapters     map[string]*Chapter
	ChapterOrder []string
}
