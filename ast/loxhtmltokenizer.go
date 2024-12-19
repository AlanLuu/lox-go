package ast

import (
	"fmt"
	"io"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/net/html"
)

type LoxHTMLTokenizerIterator struct {
	loxTokenizer *LoxHTMLTokenizer
	token        html.Token
}

func (l *LoxHTMLTokenizerIterator) HasNext() bool {
	if l.loxTokenizer.tokenizer.Err() != nil {
		return false
	}
	token := l.loxTokenizer.token()
	if token.Type == html.ErrorToken {
		return false
	}
	l.token = token
	return true
}

func (l *LoxHTMLTokenizerIterator) Next() any {
	l.loxTokenizer.next()
	return NewLoxHTMLToken(l.token)
}

type LoxHTMLTokenizer struct {
	tokenizer    *html.Tokenizer
	tokenType    html.TokenType
	currentToken html.Token
	calledNext   bool
	methods      map[string]*struct{ ProtoLoxCallable }
}

func NewLoxHTMLTokenizer(reader io.Reader) *LoxHTMLTokenizer {
	tokenizer := html.NewTokenizer(reader)
	tokenType := tokenizer.Next()
	currentToken := tokenizer.Token()
	return &LoxHTMLTokenizer{
		tokenizer:    tokenizer,
		tokenType:    tokenType,
		currentToken: currentToken,
		calledNext:   false,
		methods:      make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxHTMLTokenizerStr(str string) *LoxHTMLTokenizer {
	return NewLoxHTMLTokenizer(strings.NewReader(str))
}

func (l *LoxHTMLTokenizer) next() html.TokenType {
	tokenType := l.tokenizer.Next()
	l.tokenType = tokenType
	l.calledNext = true
	return tokenType
}

func (l *LoxHTMLTokenizer) nextNoNewLines() html.TokenType {
	onlyNewLines := func() bool {
		data := l.token().Data
		if len(data) == 0 {
			return false
		}
		carriageReturn := false
		for _, c := range data {
			switch c {
			case '\r':
				if carriageReturn {
					return false
				}
				carriageReturn = true
			case '\n':
				carriageReturn = false
			default:
				return false
			}
		}
		return true
	}
	tokenType := l.next()
	for tokenType == html.TextToken && onlyNewLines() {
		tokenType = l.next()
	}
	return tokenType
}

func (l *LoxHTMLTokenizer) token() html.Token {
	if !l.calledNext {
		return l.currentToken
	}
	token := l.tokenizer.Token()
	l.currentToken = token
	l.calledNext = false
	return token
}

func (l *LoxHTMLTokenizer) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	htmlTokenizerFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native HTML tokenizer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	switch methodName {
	case "err":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.tokenizer.Err()
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "EOF") {
					return nil, loxerror.RuntimeError(name, "HTML tokenizer: EOF")
				}
				return nil, loxerror.RuntimeError(name, errStr)
			}
			return nil, nil
		})
	case "iterNoNewLines":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			iterator := struct {
				ProtoIterator
				token html.Token
			}{}
			iterator.hasNextMethod = func() bool {
				if l.tokenizer.Err() != nil {
					return false
				}
				token := l.token()
				if token.Type == html.ErrorToken {
					return false
				}
				iterator.token = token
				return true
			}
			iterator.nextMethod = func() any {
				l.nextNoNewLines()
				return NewLoxHTMLToken(iterator.token)
			}
			return NewLoxIterator(iterator), nil
		})
	case "next":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.next()), nil
		})
	case "nextNoNewLines":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.nextNoNewLines()), nil
		})
	case "raw":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			rawBytes := l.tokenizer.Raw()
			buffer := EmptyLoxBufferCap(int64(len(rawBytes)))
			for _, b := range rawBytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "rawStr":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(string(l.tokenizer.Raw())), nil
		})
	case "token":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxHTMLToken(l.token()), nil
		})
	case "tokenType":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.tokenType), nil
		})
	case "tokenTypeStr":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			typeStr := l.tokenType.String()
			if strings.HasPrefix(typeStr, "Invalid") {
				typeStr = "Unknown"
			}
			return NewLoxString(typeStr, '\''), nil
		})
	case "toList":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			tokens := list.NewList[any]()
			for l.next() != html.ErrorToken {
				tokens.Add(NewLoxHTMLToken(l.token()))
			}
			return NewLoxList(tokens), nil
		})
	case "toListNoNewLines":
		return htmlTokenizerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			tokens := list.NewList[any]()
			for l.nextNoNewLines() != html.ErrorToken {
				tokens.Add(NewLoxHTMLToken(l.token()))
			}
			return NewLoxList(tokens), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "HTML tokenizers have no property called '"+methodName+"'.")
}

func (l *LoxHTMLTokenizer) Iterator() interfaces.Iterator {
	return &LoxHTMLTokenizerIterator{
		loxTokenizer: l,
	}
}

func (l *LoxHTMLTokenizer) String() string {
	return fmt.Sprintf("<HTML tokenizer at %p>", l)
}

func (l *LoxHTMLTokenizer) Type() string {
	return "html tokenizer"
}
