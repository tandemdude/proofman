package tokens

type TokenType int

const (
	Identifier TokenType = iota
	StringLiteral
	NumberLiteral
	Comment

	Equal
	Plus
	LeftParen
	RightParen
	LeftSquareParen
	RightSquareParen
	Comma
	Hash
	Asterisk
)

var TokenTypeName = map[TokenType]string{
	Identifier:       "Identifier",
	StringLiteral:    "StringLiteral",
	NumberLiteral:    "NumberLiteral",
	Comment:          "Comment",
	Equal:            "'='",
	Plus:             "'+'",
	LeftParen:        "'('",
	RightParen:       "')'",
	LeftSquareParen:  "'('",
	RightSquareParen: "')'",
	Comma:            "','",
	Hash:             "'#'",
	Asterisk:         "'*'",
}

type Token struct {
	Type   TokenType
	Value  string
	LineNo int
}
