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
	AMPERSAND
	ARROW
	CARET
	COLON
	COMMA
	DOT
	DOUBLE_GREATER
	DOUBLE_LESS
	DOUBLE_STAR
	ELLIPSIS
	MINUS
	PERCENT
	PIPE
	PLUS
	QUESTION
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
	BIG_NUMBER

	//Reserved keywords
	AND
	ASSERT
	BREAK
	CATCH
	CLASS
	CONTINUE
	DO
	ELSE
	ENUM
	FALSE
	FINALLY
	FOR
	FOREACH
	FUN
	IF
	IMPORT
	INFINITY
	LOOP
	NAN
	NIL
	OR
	PRINT
	PRINTERR
	PUT
	PUTERR
	REPEAT
	RETURN
	STATIC
	SUPER
	THIS
	THROW
	TRUE
	TRY
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
	"AMPERSAND",
	"ARROW",
	"CARET",
	"COLON",
	"COMMA",
	"DOT",
	"DOUBLE_GREATER",
	"DOUBLE_LESS",
	"DOUBLE_STAR",
	"ELLIPSIS",
	"MINUS",
	"PERCENT",
	"PIPE",
	"PLUS",
	"QUESTION",
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
	"BIG_NUMBER",

	//Reserved keywords
	"AND",
	"ASSERT",
	"BREAK",
	"CATCH",
	"CLASS",
	"CONTINUE",
	"DO",
	"ELSE",
	"ENUM",
	"FALSE",
	"FINALLY",
	"FOR",
	"FOREACH",
	"FUN",
	"IF",
	"IMPORT",
	"INFINITY",
	"LOOP",
	"NAN",
	"NIL",
	"OR",
	"PRINT",
	"PRINTERR",
	"PUT",
	"PUTERR",
	"REPEAT",
	"RETURN",
	"STATIC",
	"SUPER",
	"THIS",
	"THROW",
	"TRUE",
	"TRY",
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

func NewToken(tokenType TokenType, lexeme string, literal any, line int, quote byte) *Token {
	return &Token{
		TokenType: tokenType,
		Lexeme:    lexeme,
		Literal:   literal,
		Line:      line,
		Quote:     quote,
	}
}

func (t *Token) String() string {
	return fmt.Sprintf("Token [TokenType=%v, Lexeme=%v, Literal=%v, Line=%v, Quote=%c]",
		tokenArr[t.TokenType], t.Lexeme, t.Literal, t.Line, t.Quote)
}
