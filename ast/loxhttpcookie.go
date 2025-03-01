package ast

import (
	"fmt"
	"net/http"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxHTTPCookie struct {
	cookie     *http.Cookie
	properties map[string]any
}

func NewLoxHTTPCookie(cookie *http.Cookie) *LoxHTTPCookie {
	return &LoxHTTPCookie{
		cookie:     cookie,
		properties: make(map[string]any),
	}
}

func (l *LoxHTTPCookie) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	cookieField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	cookieFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native http cookie fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	switch lexemeName {
	case "domain":
		domainRunes := []rune(l.cookie.Domain)
		if len(domainRunes) > 0 && domainRunes[0] == '.' {
			return cookieField(NewLoxStringQuote(string(domainRunes[1:])))
		}
		return cookieField(NewLoxStringQuote(l.cookie.Domain))
	case "expires":
		return cookieField(NewLoxDate(l.cookie.Expires))
	case "httpOnly":
		return cookieField(l.cookie.HttpOnly)
	case "maxAge":
		return cookieField(int64(l.cookie.MaxAge))
	case "name":
		return cookieField(NewLoxStringQuote(l.cookie.Name))
	case "path":
		return cookieField(NewLoxStringQuote(l.cookie.Path))
	case "printSimpleString":
		return cookieFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			name, value := l.cookie.Name, l.cookie.Value
			if name == "" && value == "" {
				fmt.Println("''")
			} else {
				fmt.Println(name + "=" + value)
			}
			return nil, nil
		})
	case "printString":
		return cookieFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			str := l.cookie.String()
			if str == "" {
				fmt.Println("''")
			} else {
				fmt.Println(str)
			}
			return nil, nil
		})
	case "raw":
		return cookieField(NewLoxStringQuote(l.cookie.Raw))
	case "rawExpires":
		return cookieField(NewLoxStringQuote(l.cookie.RawExpires))
	case "secure":
		return cookieField(l.cookie.Secure)
	case "simpleString":
		name, value := l.cookie.Name, l.cookie.Value
		if name == "" && value == "" {
			return cookieField(EmptyLoxString())
		}
		return cookieField(NewLoxStringQuote(name + "=" + value))
	case "string":
		return cookieFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.cookie.String()), nil
		})
	case "unparsed":
		strList := list.NewListCap[any](int64(len(l.cookie.Unparsed)))
		for _, str := range l.cookie.Unparsed {
			strList.Add(NewLoxStringQuote(str))
		}
		return cookieField(NewLoxList(strList))
	case "valid":
		return cookieFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.cookie.Valid() == nil, nil
		})
	case "validErr":
		return cookieFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxError(l.cookie.Valid()), nil
		})
	case "validThrowErr":
		return cookieFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			err := l.cookie.Valid()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return nil, nil
		})
	case "value":
		return cookieField(NewLoxStringQuote(l.cookie.Value))
	}
	return nil, loxerror.RuntimeError(name, "HTTP cookies have no property called '"+lexemeName+"'.")
}

func (l *LoxHTTPCookie) String() string {
	return fmt.Sprintf("<http cookie at %p>", l)
}

func (l *LoxHTTPCookie) Type() string {
	return "http cookie"
}
