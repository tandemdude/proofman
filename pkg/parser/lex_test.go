package parser

import (
	asrt "github.com/stretchr/testify/assert"
	"testing"
)

type testcase struct {
	input string
	idx   int
	want  string
}

func TestParseIdentifier(t *testing.T) {
	assert := asrt.New(t)

	cases := []testcase{
		{"foo", 0, "foo"},
		{"baz_123       ", 0, "baz_123"},
		{"abc_DEF_456", 0, "abc_DEF_456"},
		{"    hello_world", 4, "hello_world"},
		{"      test123      ", 6, "test123"},
		{" varName", 1, "varName"},
		{"identifier1     identifier2", 0, "identifier1"},
	}

	for _, tt := range cases {
		str, err := parseIdentifier(&tt.input, tt.idx)
		assert.NoError(err, "Unexpected error for input: %s", tt.input)
		assert.Equal(tt.want, str, "Expected: %s, got: %s", tt.want, str)
	}
}

func TestParseStringLiteral(t *testing.T) {
	assert := asrt.New(t)

	cases := []testcase{
		{`"foo"`, 0, `foo`},
		{`"hello, world!"`, 0, `hello, world!`},
		{`"string with a newline\n"`, 0, `string with a newline\n`},
		{`"escaped quote: \"quoted\" text"`, 0, `escaped quote: \"quoted\" text`},
		{`"multi\nline\nstring"`, 0, `multi\nline\nstring`},
		{`"string with an escaped backslash: \\"`, 0, `string with an escaped backslash: \\`},
		{`"empty string:"`, 0, `empty string:`},
		{`"string with \t tab"`, 0, `string with \t tab`},
		{`"string with escape sequences: \n\t\\"`, 0, `string with escape sequences: \n\t\\`},
	}

	for _, tt := range cases {
		str, err := parseStringLiteral(&tt.input, tt.idx)
		assert.NoError(err, "Unexpected error for input: %s", tt.input)
		assert.Equal(tt.want, str, "Expected: %s, got: %s", tt.want, str)
	}

	cases = []testcase{
		{`"unterminated`, 0, ""},
		{`"still unterminated`, 0, ""},
		{`"another unterminated string`, 0, ""},
		{`"missing closing quote`, 0, ""},
		{`"escaped quote: \"still unterminated`, 0, ""},
	}

	for _, tt := range cases {
		str, err := parseStringLiteral(&tt.input, tt.idx)
		assert.Error(err, "Expected an error for input: %s", tt.input)
		assert.Empty(str, "Expected empty string for input: %s", tt.input)
	}
}

var (
	rawLatexTestString = `\<open>
foo bar baz
\<close>`
	latexTestString = rawLatexTestString + "      "
)

func TestParseLatexStringLiteral(t *testing.T) {
	assert := asrt.New(t)

	str, err := parseLatexStringLiteral(&latexTestString, 0)
	assert.NoError(err, "Unexpected error for input: %s", latexTestString)
	assert.Equal(rawLatexTestString, str, "Expected: %s, got: %s", rawLatexTestString, str)
}

func TestParseNumberLiteral(t *testing.T) {
	assert := asrt.New(t)

	cases := []testcase{
		{"123", 0, "123"},
		{"0", 0, "0"},
		{"456.789", 0, "456.789"},
		{"3.14", 0, "3.14"},
		{"100.000", 0, "100.000"},
		{"987654321", 0, "987654321"},
		{"12.34 extra", 0, "12.34"},
		{"12.34\n", 0, "12.34"},
		{"3.14  ", 0, "3.14"},
		{"12.34 56.78", 0, "12.34"},
	}

	for _, tt := range cases {
		str, err := parseNumberLiteral(&tt.input, tt.idx)
		assert.NoError(err, "Unexpected error for input: %s", tt.input)
		assert.Equal(tt.want, str, "Expected: %s, got: %s", tt.want, str)
	}

	cases = []testcase{
		{".5", 0, ".5"},
		{"5.", 0, "5."},
	}

	for _, tt := range cases {
		str, err := parseStringLiteral(&tt.input, tt.idx)
		assert.Error(err, "Expected an error for input: %s", tt.input)
		assert.Empty(str, "Expected empty string for input: %s", tt.input)
	}
}

func TestParseComment(t *testing.T) {
	assert := asrt.New(t)

	cases := []testcase{
		{"Some text (* This is a comment *) and more text", 10, " This is a comment "},
		{"(* Single comment *)", 0, " Single comment "},
		{"(* Comment with nested (* inner comment *) inside *)", 0, " Comment with nested (* inner comment *) inside "},
		{"(* Escaped \\*) and text *)", 0, " Escaped \\*) and text "},
		{"(* Nested (* comment (* another *) *) *)", 0, " Nested (* comment (* another *) *) "},
	}

	for _, tt := range cases {
		comment, err := parseComment(&tt.input, tt.idx)
		assert.NoError(err, "Unexpected error for input: %s", tt.input)
		assert.Equal(tt.want, comment, "Expected: %s, got: %s", tt.want, comment)
	}

	cases = []testcase{
		{"(* Missing closing sequence", 0, ""},
		{"(* Unmatched (* opening", 0, ""},
		{"(* Comment with an escaped closing \\*) here", 0, ""},
	}

	for _, tt := range cases {
		comment, err := parseComment(&tt.input, tt.idx)
		assert.Error(err, "Expected an error for input: %s", tt.input)
		assert.Empty(comment, "Expected empty string for input: %s", tt.input)
	}
}
