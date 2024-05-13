package token

import "fmt"

type TokenType int

const (
	//Bracket operators
	LEFT_PAREN TokenType = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	LEFT_BRACKET
	RIGHT_BRACKET

	//Operators
	AND_SYMBOL
	ARROW
	CARET
	COLON
	COMMA
	DOT
	DOUBLE_GREATER
	DOUBLE_LESS
	DOUBLE_STAR
	MINUS
	OR_SYMBOL
	PERCENT
	PLUS
	SEMICOLON
	SLASH
	STAR
	TILDE

	//Comparison operators
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	//Names, strings, and numbers
	IDENTIFIER
	STRING
	NUMBER

	//Reserved keywords
	AND
	BREAK
	CLASS
	CONTINUE
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	STATIC
	SUPER
	THIS
	TRUE
	VAR
	WHILE

	//EOF token
	EOF
)

var tokenArr = [...]string{
	//Bracket operators
	"LEFT_PAREN",
	"RIGHT_PAREN",
	"LEFT_BRACE",
	"RIGHT_BRACE",
	"LEFT_BRACKET",
	"RIGHT_BRACKET",

	//Operators
	"AND_SYMBOL",
	"ARROW",
	"CARET",
	"COLON",
	"COMMA",
	"DOT",
	"DOUBLE_GREATER",
	"DOUBLE_LESS",
	"DOUBLE_STAR",
	"MINUS",
	"OR_SYMBOL",
	"PERCENT",
	"PLUS",
	"SEMICOLON",
	"SLASH",
	"STAR",
	"TILDE",

	//Comparison operators
	"BANG",
	"BANG_EQUAL",
	"EQUAL",
	"EQUAL_EQUAL",
	"GREATER",
	"GREATER_EQUAL",
	"LESS",
	"LESS_EQUAL",

	//Names, strings, and numbers
	"IDENTIFIER",
	"STRING",
	"NUMBER",

	//Reserved keywords
	"AND",
	"BREAK",
	"CLASS",
	"CONTINUE",
	"ELSE",
	"FALSE",
	"FUN",
	"FOR",
	"IF",
	"NIL",
	"OR",
	"PRINT",
	"RETURN",
	"STATIC",
	"SUPER",
	"THIS",
	"TRUE",
	"VAR",
	"WHILE",

	//EOF token
	"EOF",
}

type Token struct {
	TokenType
	Lexeme  string
	Literal any
	Line    int
	Quote   byte
}

func NewToken(tokenType TokenType, lexeme string, literal any, line int, quote byte) Token {
	return Token{
		TokenType: tokenType,
		Lexeme:    lexeme,
		Literal:   literal,
		Line:      line,
		Quote:     quote,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("Token [TokenType=%v, Lexeme=%v, Literal=%v, Line=%v, Quote=%c]",
		tokenArr[t.TokenType], t.Lexeme, t.Literal, t.Line, t.Quote)
}
