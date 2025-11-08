package scanner

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

var keywords = map[string]token.TokenType{
	"and":      token.AND,
	"assert":   token.ASSERT,
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
	"loop":     token.LOOP,
	"NaN":      token.NAN,
	"nil":      token.NIL,
	"or":       token.OR,
	"print":    token.PRINT,
	"printerr": token.PRINTERR,
	"put":      token.PUT,
	"puterr":   token.PUTERR,
	"repeat":   token.REPEAT,
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

var escapeChars = map[rune]rune{
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

var keywordAsIdentifier = false

type Scanner struct {
	sourceRunes  []rune
	sourceLen    int
	Tokens       list.List[*token.Token]
	startIndex   int
	currentIndex int
	lineNum      int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		sourceRunes:  []rune(source),
		sourceLen:    utf8.RuneCountInString(source),
		Tokens:       list.NewList[*token.Token](),
		startIndex:   0,
		currentIndex: 0,
		lineNum:      1,
	}
}

func (sc *Scanner) advance() rune {
	c := sc.sourceRunes[sc.currentIndex]
	sc.currentIndex++
	return c
}

func (sc *Scanner) addToken(tokenType token.TokenType, literal any, quote byte) {
	text := string(sc.sourceRunes[sc.startIndex:sc.currentIndex])
	sc.Tokens.Add(token.NewToken(tokenType, text, literal, sc.lineNum, quote))
}

func (sc *Scanner) handleNumber() error {
	digitSeparator := '_'
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

	bigNum := false
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
	case 'n':
		bigNum = true
		sc.advance()
	}

	numStr := string(sc.sourceRunes[sc.startIndex:sc.currentIndex])
	invalidLiteral := func(numType string) error {
		return loxerror.GiveError(sc.lineNum, "", "Invalid "+numType+" literal")
	}
	if bigNum {
		tokenStr := numStr[:len(numStr)-1]
		if strings.Contains(tokenStr, "__") {
			return invalidLiteral("number")
		}
		if numHasDot {
			sc.addToken(token.BIG_NUMBER, tokenStr, 255)
		} else {
			sc.addToken(token.BIG_NUMBER, tokenStr, 0)
		}
	} else if numHasDot || scientific {
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

	text := string(sc.sourceRunes[sc.startIndex:sc.currentIndex])
	tokenType, ok := keywords[text]
	if keywordAsIdentifier || !ok {
		tokenType = token.IDENTIFIER
		keywordAsIdentifier = false
	}
	sc.addToken(tokenType, nil, 0)
}

func (sc *Scanner) handleString(quote rune) error {
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
		currentChar := sc.sourceRunes[sc.currentIndex]
		if !foundBackslash && currentChar == '\\' {
			foundBackslash = true
		} else if foundBackslash {
			escapeChar, ok := escapeChars[currentChar]
			if !ok {
				return loxerror.GiveError(sc.lineNum, "",
					"Unknown escape character '"+string(currentChar)+"'.")
			}
			builder.WriteRune(escapeChar)
			foundBackslash = false
		} else {
			builder.WriteRune(sc.sourceRunes[sc.currentIndex])
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

func (sc *Scanner) handleMultiLineComment() error {
	unclosedCommentErr := func() error {
		return loxerror.GiveError(sc.lineNum, "", "Unclosed multi-line comment")
	}
	foundStar := false
	for cond := true; cond; sc.advance() {
		if sc.isAtEnd() {
			return unclosedCommentErr()
		}
		switch sc.peek() {
		case '\n':
			sc.lineNum++
		case '*':
			foundStar = true
			continue
		case '/':
			if foundStar {
				cond = false
			}
			continue
		}
		foundStar = false
	}
	return nil
}

func (sc *Scanner) isAtEnd() bool {
	return sc.currentIndex >= sc.sourceLen
}

func isAlpha(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
}

func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || isDigit(r)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isBinaryDigit(r rune) bool {
	return r == '0' || r == '1'
}

func isHexDigit(r rune) bool {
	return isDigit(r) || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

func isOctalDigit(r rune) bool {
	return r >= '0' && r <= '7'
}

func (sc *Scanner) match(expected rune) bool {
	if sc.isAtEnd() {
		return false
	}
	if sc.sourceRunes[sc.currentIndex] != expected {
		return false
	}
	sc.currentIndex++
	return true
}

func (sc *Scanner) peek() rune {
	if sc.isAtEnd() {
		return 0
	}
	return sc.sourceRunes[sc.currentIndex]
}

func (sc *Scanner) previous() rune {
	if sc.currentIndex == 0 {
		return 0
	}
	return sc.sourceRunes[sc.currentIndex-1]
}

func (sc *Scanner) scanToken() error {
	c := sc.advance()
	addToken := func(tokenType token.TokenType) {
		sc.addToken(tokenType, nil, 0)
	}
	foundDot := false
	foundWhitespaceChar := false
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
		if sc.match('.') {
			if sc.match('.') {
				addToken(token.ELLIPSIS)
			} else {
				sc.currentIndex--
				addToken(token.DOT)
				sc.currentIndex++
				addToken(token.DOT)
				foundDot = true
			}
		} else {
			addToken(token.DOT)
			foundDot = true
		}
	case '?':
		addToken(token.QUESTION)
	case '&':
		if sc.match('&') { //handle "&&"
			addToken(token.AND)
		} else {
			addToken(token.AMPERSAND)
		}
	case '|':
		if sc.match('|') { //handle "||"
			addToken(token.OR)
		} else {
			addToken(token.PIPE)
		}
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
		} else if sc.match('*') { //handle "/*" (multi-line comment)
			multiLineCommentErr := sc.handleMultiLineComment()
			if multiLineCommentErr != nil {
				return multiLineCommentErr
			}
		} else {
			addToken(token.SLASH)
		}
	case '%':
		addToken(token.PERCENT)

	case '\n':
		foundWhitespaceChar = true
		sc.lineNum++

	case ' ', '\r', '\t':
		foundWhitespaceChar = true

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
	if !foundWhitespaceChar {
		keywordAsIdentifier = foundDot
	}
	return nil
}

func (sc *Scanner) ScanTokens() error {
	source := &sc.sourceRunes
	if sc.sourceLen > 1 && (*source)[0] == '#' && (*source)[1] == '!' {
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
	sc.sourceRunes = []rune(sourceLine)
	sc.sourceLen = utf8.RuneCountInString(sourceLine)
}
