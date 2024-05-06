package scanner

import (
	"strconv"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

var keywords = map[string]token.TokenType{
	"and":      token.AND,
	"break":    token.BREAK,
	"class":    token.CLASS,
	"continue": token.CONTINUE,
	"else":     token.ELSE,
	"false":    token.FALSE,
	"for":      token.FOR,
	"fun":      token.FUN,
	"if":       token.IF,
	"nil":      token.NIL,
	"or":       token.OR,
	"print":    token.PRINT,
	"return":   token.RETURN,
	"super":    token.SUPER,
	"this":     token.THIS,
	"true":     token.TRUE,
	"var":      token.VAR,
	"while":    token.WHILE,
}

type Scanner struct {
	sourceLine   string
	Tokens       list.List[token.Token]
	startIndex   int
	currentIndex int
	lineNum      int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		sourceLine:   source,
		Tokens:       list.NewList[token.Token](),
		startIndex:   0,
		currentIndex: 0,
		lineNum:      1,
	}
}

func (sc *Scanner) advance() byte {
	c := sc.sourceLine[sc.currentIndex]
	sc.currentIndex++
	return c
}

func (sc *Scanner) addToken(tokenType token.TokenType, literal any, quote byte) {
	text := sc.sourceLine[sc.startIndex:sc.currentIndex]
	sc.Tokens.Add(token.NewToken(tokenType, text, literal, sc.lineNum, quote))
}

func (sc *Scanner) handleNumber() error {
	for isDigit(sc.peek()) {
		sc.advance()
	}

	if sc.peek() == '.' {
		sc.advance()
		if isDigit(sc.peek()) {
			for isDigit(sc.peek()) {
				sc.advance()
			}
		} else {
			return loxerror.GiveError(sc.lineNum, "", "Unexpected '.' in number")
		}
	}

	numStr := sc.sourceLine[sc.startIndex:sc.currentIndex]
	if strings.Contains(numStr, ".") {
		num, _ := strconv.ParseFloat(numStr, 64)
		sc.addToken(token.NUMBER, num, 0)
	} else {
		num, _ := strconv.ParseInt(numStr, 10, 64)
		sc.addToken(token.NUMBER, num, 0)
	}

	return nil
}

func (sc *Scanner) handleIdentifier() {
	for isAlphaNumeric(sc.peek()) {
		sc.advance()
	}

	text := sc.sourceLine[sc.startIndex:sc.currentIndex]
	tokenType, ok := keywords[text]
	if !ok {
		tokenType = token.IDENTIFIER
	}
	sc.addToken(tokenType, nil, 0)
}

func (sc *Scanner) handleString(quote byte) error {
	unclosedStringErr := func() error {
		return loxerror.GiveError(sc.lineNum, "", "Unclosed string")
	}
	var tokenQuote byte = '\''
	for sc.peek() != quote && !sc.isAtEnd() {
		if sc.peek() == '\n' {
			return unclosedStringErr()
		}
		if tokenQuote != '"' && quote == '"' && sc.peek() == '\'' {
			tokenQuote = '"'
		}
		sc.advance()
	}

	if sc.isAtEnd() {
		return unclosedStringErr()
	}
	sc.advance()

	theString := sc.sourceLine[sc.startIndex+1 : sc.currentIndex-1]
	sc.addToken(token.STRING, theString, tokenQuote)
	return nil
}

func (sc *Scanner) isAtEnd() bool {
	return sc.currentIndex >= len(sc.sourceLine)
}

func isAlpha(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '_'
}

func isAlphaNumeric(b byte) bool {
	return isAlpha(b) || isDigit(b)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func (sc *Scanner) match(expected byte) bool {
	if sc.isAtEnd() {
		return false
	}
	if sc.sourceLine[sc.currentIndex] != expected {
		return false
	}
	sc.currentIndex++
	return true
}

func (sc *Scanner) peek() byte {
	if sc.isAtEnd() {
		return 0
	}
	return sc.sourceLine[sc.currentIndex]
}

func (sc *Scanner) ResetForNextLine() {
	sc.Tokens.Clear()
	sc.currentIndex = 0
	sc.startIndex = 0
}

func (sc *Scanner) scanToken() error {
	c := sc.advance()
	addToken := func(tokenType token.TokenType) {
		sc.addToken(tokenType, nil, 0)
	}
	switch c {
	case '(':
		addToken(token.LEFT_PAREN)
	case ')':
		addToken(token.RIGHT_PAREN)
	case '{':
		addToken(token.LEFT_BRACE)
	case '}':
		addToken(token.RIGHT_BRACE)
	case '[':
		addToken(token.LEFT_BRACKET)
	case ']':
		addToken(token.RIGHT_BRACKET)
	case ':':
		addToken(token.COLON)
	case ',':
		addToken(token.COMMA)
	case '.':
		addToken(token.DOT)
	case '&':
		addToken(token.AND_SYMBOL)
	case '|':
		addToken(token.OR_SYMBOL)
	case '^':
		addToken(token.CARET)
	case '-':
		addToken(token.MINUS)
	case '+':
		addToken(token.PLUS)
	case ';':
		addToken(token.SEMICOLON)
	case '*':
		if sc.match('*') {
			addToken(token.DOUBLE_STAR)
		} else {
			addToken(token.STAR)
		}
	case '!':
		if sc.match('=') { //handle "!="
			addToken(token.BANG_EQUAL)
		} else {
			addToken(token.BANG)
		}
	case '~':
		addToken(token.TILDE)
	case '=':
		if sc.match('=') { //handle "=="
			addToken(token.EQUAL_EQUAL)
		} else if sc.match('>') { //handle "=>"
			addToken(token.ARROW)
		} else {
			addToken(token.EQUAL)
		}
	case '<':
		if sc.match('=') { //handle "<="
			addToken(token.LESS_EQUAL)
		} else if sc.match('<') { //handle "<<"
			addToken(token.DOUBLE_LESS)
		} else {
			addToken(token.LESS)
		}
	case '>':
		if sc.match('=') { //handle ">="
			addToken(token.GREATER_EQUAL)
		} else if sc.match('>') { //handle ">>"
			addToken(token.DOUBLE_GREATER)
		} else {
			addToken(token.GREATER)
		}
	case '/':
		if sc.match('/') { //handle "//" (comment)
			for sc.peek() != '\n' && !sc.isAtEnd() {
				sc.currentIndex++
			}
		} else {
			addToken(token.SLASH)
		}
	case '%':
		addToken(token.PERCENT)

	case '\n':
		sc.lineNum++

	case ' ':
	case '\r':
	case '\t':

	case '"', '\'':
		handleStringErr := sc.handleString(c)
		if handleStringErr != nil {
			return handleStringErr
		}
	default:
		switch {
		case isDigit(c):
			handleNumberErr := sc.handleNumber()
			if handleNumberErr != nil {
				return handleNumberErr
			}
		case isAlpha(c):
			sc.handleIdentifier()
		default:
			unexpectedChar := "Unexpected character '" + string(c) + "'."
			return loxerror.GiveError(sc.lineNum, "", unexpectedChar)
		}
	}
	return nil
}

func (sc *Scanner) ScanTokens() error {
	for !sc.isAtEnd() {
		sc.startIndex = sc.currentIndex
		scanTokenErr := sc.scanToken()
		if scanTokenErr != nil {
			return scanTokenErr
		}
	}
	var eofLineNum int
	if sc.Tokens.IsEmpty() {
		eofLineNum = sc.lineNum
	} else {
		eofLineNum = sc.Tokens.Peek().Line
	}
	sc.Tokens.Add(token.NewToken(token.EOF, "", nil, eofLineNum, 0))
	return nil
}

func (sc *Scanner) SetSourceLine(sourceLine string) {
	sc.sourceLine = sourceLine
}
