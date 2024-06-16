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
	"catch":    token.CATCH,
	"class":    token.CLASS,
	"continue": token.CONTINUE,
	"do":       token.DO,
	"else":     token.ELSE,
	"enum":     token.ENUM,
	"false":    token.FALSE,
	"finally":  token.FINALLY,
	"for":      token.FOR,
	"foreach":  token.FOREACH,
	"fun":      token.FUN,
	"if":       token.IF,
	"import":   token.IMPORT,
	"Infinity": token.INFINITY,
	"NaN":      token.NAN,
	"nil":      token.NIL,
	"or":       token.OR,
	"print":    token.PRINT,
	"put":      token.PUT,
	"return":   token.RETURN,
	"static":   token.STATIC,
	"super":    token.SUPER,
	"this":     token.THIS,
	"throw":    token.THROW,
	"true":     token.TRUE,
	"try":      token.TRY,
	"var":      token.VAR,
	"while":    token.WHILE,
}

var escapeChars = map[byte]byte{
	'\'': '\'',
	'"':  '"',
	'\\': '\\',
	'a':  '\a',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'b':  '\b',
	'f':  '\f',
	'v':  '\v',
}

type Scanner struct {
	sourceLine   string
	Tokens       list.List[*token.Token]
	startIndex   int
	currentIndex int
	lineNum      int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		sourceLine:   source,
		Tokens:       list.NewList[*token.Token](),
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
	var digitSeparator byte = '_'
	isBinaryNum := false
	isHexNum := false
	isOctalNum := false
	if sc.previous() == '0' {
		switch sc.peek() {
		case 'b', 'B':
			isBinaryNum = true
			sc.advance()
			for isBinaryDigit(sc.peek()) || sc.peek() == digitSeparator {
				sc.advance()
			}
		case 'x', 'X':
			isHexNum = true
			sc.advance()
			for isHexDigit(sc.peek()) || sc.peek() == digitSeparator {
				sc.advance()
			}
		case 'o', 'O':
			isOctalNum = true
			sc.advance()
			for isOctalDigit(sc.peek()) || sc.peek() == digitSeparator {
				sc.advance()
			}
		default:
			for isOctalDigit(sc.peek()) {
				if !isOctalNum {
					isOctalNum = true
				}
				sc.advance()
			}
		}
	} else {
		for isDigit(sc.peek()) || sc.peek() == digitSeparator {
			sc.advance()
		}
	}

	numHasDot := false
	if sc.peek() == '.' {
		unexpectedDotIn := func(numType string) error {
			return loxerror.GiveError(sc.lineNum, "", "Unexpected '.' in "+numType)
		}
		switch {
		case isBinaryNum:
			return unexpectedDotIn("binary literal")
		case isHexNum:
			return unexpectedDotIn("hex literal")
		case isOctalNum:
			return unexpectedDotIn("octal literal")
		}
		numHasDot = true
		sc.advance()
		if isDigit(sc.peek()) {
			for isDigit(sc.peek()) || sc.peek() == digitSeparator {
				sc.advance()
			}
		} else {
			return unexpectedDotIn("number")
		}
	}

	scientific := false
	switch sc.peek() {
	case 'e', 'E':
		if isBinaryNum || isHexNum || isOctalNum {
			break
		}
		sc.advance()
		foundSign := false
		switch sc.peek() {
		case '-', '+':
			foundSign = true
			sc.advance()
		}
		if isDigit(sc.peek()) {
			scientific = true
			for isDigit(sc.peek()) || sc.peek() == digitSeparator {
				sc.advance()
			}
		} else {
			sc.currentIndex--
			if foundSign {
				sc.currentIndex--
			}
		}
	}

	numStr := sc.sourceLine[sc.startIndex:sc.currentIndex]
	invalidLiteral := func(numType string) error {
		return loxerror.GiveError(sc.lineNum, "", "Invalid "+numType+" literal")
	}
	if numHasDot || scientific {
		num, numErr := strconv.ParseFloat(numStr, 64)
		if numErr != nil {
			return invalidLiteral("number")
		}
		sc.addToken(token.NUMBER, num, 0)
	} else {
		num, numErr := strconv.ParseInt(numStr, 0, 64)
		if numErr != nil {
			switch {
			case isBinaryNum:
				return invalidLiteral("binary")
			case isHexNum:
				return invalidLiteral("hex")
			case isOctalNum:
				return invalidLiteral("octal")
			}
			return invalidLiteral("number")
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
	var builder strings.Builder
	var tokenQuote byte = '\''
	foundBackslash := false
	for foundBackslash || (sc.peek() != quote && !sc.isAtEnd()) {
		if sc.peek() == '\n' {
			sc.lineNum++
		}
		if tokenQuote != '"' && sc.peek() == '\'' {
			tokenQuote = '"'
		} else if tokenQuote == '"' && sc.peek() == '"' {
			tokenQuote = '\''
		}
		currentChar := sc.sourceLine[sc.currentIndex]
		if !foundBackslash && currentChar == '\\' {
			foundBackslash = true
		} else if foundBackslash {
			escapeChar, ok := escapeChars[currentChar]
			if !ok {
				return loxerror.GiveError(sc.lineNum, "",
					"Unknown escape character '"+string(currentChar)+"'.")
			}
			builder.WriteByte(escapeChar)
			foundBackslash = false
		} else {
			builder.WriteByte(sc.sourceLine[sc.currentIndex])
		}
		sc.advance()
	}

	if sc.isAtEnd() {
		return unclosedStringErr()
	}
	sc.advance()

	sc.addToken(token.STRING, builder.String(), tokenQuote)
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

func (sc *Scanner) previous() byte {
	if sc.currentIndex == 0 {
		return 0
	}
	return sc.sourceLine[sc.currentIndex-1]
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
	case '?':
		addToken(token.QUESTION)
	case '&':
		addToken(token.AMPERSAND)
	case '|':
		addToken(token.PIPE)
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
