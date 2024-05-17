package scanner

import (
	"strconv"

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
	"static":   token.STATIC,
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
	isBinaryNum := false
	isHexNum := false
	isOctalNum := false
	switch sc.peek() {
	case 'b', 'B':
		isBinaryNum = true
		sc.advance()
		for isBinaryDigit(sc.peek()) {
			sc.advance()
		}
	case 'x', 'X':
		isHexNum = true
		sc.advance()
		for isHexDigit(sc.peek()) {
			sc.advance()
		}
	case 'o', 'O':
		isOctalNum = true
		sc.advance()
		for isOctalDigit(sc.peek()) {
			sc.advance()
		}
	default:
		for isDigit(sc.peek()) {
			sc.advance()
		}
	}

	numHasDot := false
	if sc.peek() == '.' {
		unexpectedDotIn := func(numType string) error {
			return loxerror.GiveError(sc.lineNum, "", "Unexpected '.' in "+numType)
		}
		if isBinaryNum {
			return unexpectedDotIn("binary literal")
		}
		if isHexNum {
			return unexpectedDotIn("hex literal")
		}
		if isOctalNum {
			return unexpectedDotIn("octal literal")
		}
		numHasDot = true
		sc.advance()
		if isDigit(sc.peek()) {
			for isDigit(sc.peek()) {
				sc.advance()
			}
		} else {
			return unexpectedDotIn("number")
		}
	}

	numStr := sc.sourceLine[sc.startIndex:sc.currentIndex]
	if numHasDot {
		num, _ := strconv.ParseFloat(numStr, 64)
		sc.addToken(token.NUMBER, num, 0)
	} else {
		var num int64
		if isBinaryNum || isHexNum || isOctalNum {
			num, _ = strconv.ParseInt(numStr, 0, 64)
		} else {
			num, _ = strconv.ParseInt(numStr, 10, 64)
		}
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

func isBinaryDigit(b byte) bool {
	return b == '0' || b == '1'
}

func isHexDigit(b byte) bool {
	return isDigit(b) || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
}

func isOctalDigit(b byte) bool {
	return b >= '0' && b <= '7'
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
	source := &sc.sourceLine
	if len(*source) > 1 && (*source)[0] == '#' && (*source)[1] == '!' {
		//Ignore line with "#!" (Unix shebang) at beginning of first line
		for sc.peek() != '\n' && !sc.isAtEnd() {
			sc.currentIndex++
		}
	}
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
