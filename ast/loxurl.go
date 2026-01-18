package ast

import (
	"fmt"
	"net/url"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxURL struct {
	url        *url.URL
	properties map[string]any
}

func NewLoxURL(url *url.URL) *LoxURL {
	return &LoxURL{
		url:        url,
		properties: make(map[string]any),
	}
}

func NewLoxURLStr(str string) (*LoxURL, error) {
	url, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	return NewLoxURL(url), nil
}

func NewLoxURLURIStr(str string) (*LoxURL, error) {
	url, err := url.ParseRequestURI(str)
	if err != nil {
		return nil, err
	}
	return NewLoxURL(url), nil
}

func (l *LoxURL) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxURL:
		if l == obj {
			return true
		}
		return l.url.String() == obj.url.String()
	default:
		return false
	}
}

type loxurl_getStrField struct {
	loxStr  *LoxString
	strFunc func() string
}

func (l *LoxURL) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		switch field := field.(type) {
		case *loxurl_getStrField:
			if s := field.strFunc(); s != field.loxStr.str {
				field.loxStr = NewLoxStringQuote(s)
			}
			return field.loxStr, nil
		default:
			return field, nil
		}
	}
	urlStrField := func(strFunc func() string) (any, error) {
		loxStr := NewLoxStringQuote(strFunc())
		if value, ok := l.properties[lexemeName]; ok {
			switch value := value.(type) {
			case *loxurl_getStrField:
				value.loxStr = loxStr
				value.strFunc = strFunc
			}
		} else {
			l.properties[lexemeName] = &loxurl_getStrField{loxStr, strFunc}
		}
		return loxStr, nil
	}
	urlFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native URL object fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'URL object.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "clearUserPass":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.url.User = nil
			return nil, nil
		})
	case "escapedFragment":
		return urlStrField(l.url.EscapedFragment)
	case "escapedPath":
		return urlStrField(l.url.EscapedPath)
	case "forceQuery":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if forceQuery, ok := args[0].(bool); ok {
				l.url.ForceQuery = forceQuery
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "fragment":
		return urlStrField(func() string { return l.url.Fragment })
	case "host":
		return urlStrField(func() string { return l.url.Host })
	case "hostname":
		return urlStrField(l.url.Hostname)
	case "isAbs":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.url.IsAbs(), nil
		})
	case "joinPath":
		return urlFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var strArgs []string
			for i, arg := range args {
				switch arg := arg.(type) {
				case *LoxString:
					if i == 0 {
						strArgs = make([]string, 0, len(args))
					}
					strArgs = append(strArgs, arg.str)
				default:
					strArgs = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument #%v in 'URL object.joinPath' must be a string.",
							i+1,
						),
					)
				}
			}
			if newUrl := l.url.JoinPath(strArgs...); newUrl != nil {
				return NewLoxURL(newUrl), nil
			}
			return nil, nil
		})
	case "omitHost":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if omitHost, ok := args[0].(bool); ok {
				l.url.OmitHost = omitHost
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "opaque":
		return urlStrField(func() string { return l.url.Opaque })
	case "parse":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				newUrl, err := l.url.Parse(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				if newUrl != nil {
					return NewLoxURL(newUrl), nil
				}
				return nil, nil
			}
			return argMustBeType("string")
		})
	case "path":
		return urlStrField(func() string { return l.url.Path })
	case "port":
		return urlStrField(l.url.Port)
	case "query":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxURLValues(l.url.Query()), nil
		})
	case "queryDict":
		return urlFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			for key, values := range l.url.Query() {
				inner := list.NewListCap[any](int64(len(values)))
				for _, value := range values {
					inner.Add(NewLoxStringQuote(value))
				}
				dict.setKeyValue(NewLoxStringQuote(key), NewLoxList(inner))
			}
			return dict, nil
		})
	case "rawFragment":
		return urlStrField(func() string { return l.url.RawFragment })
	case "rawPath":
		return urlStrField(func() string { return l.url.RawPath })
	case "rawQuery":
		return urlStrField(func() string { return l.url.RawQuery })
	case "redacted":
		return urlStrField(l.url.Redacted)
	case "requestURI":
		return urlStrField(l.url.RequestURI)
	case "scheme":
		return urlStrField(func() string { return l.url.Scheme })
	case "setUser":
		return urlFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.url.User = url.User(loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "setUserPass":
		return urlFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'URL object.setUserPass' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'URL object.setUserPass' must be a string.")
			}
			l.url.User = url.UserPassword(
				args[0].(*LoxString).str,
				args[1].(*LoxString).str,
			)
			return l, nil
		})
	case "str":
		return urlStrField(l.url.String)
	case "userInfoPass":
		return urlStrField(func() string {
			if l.url.User == nil {
				return ""
			}
			pass, _ := l.url.User.Password()
			return pass
		})
	case "userInfoStr":
		return urlStrField(func() string {
			if l.url.User == nil {
				return ""
			}
			return l.url.User.String()
		})
	case "userInfoUser":
		return urlStrField(func() string {
			if l.url.User == nil {
				return ""
			}
			return l.url.User.Username()
		})
	}
	return nil, loxerror.RuntimeError(name, "URL objects have no property called '"+lexemeName+"'.")
}

func (l *LoxURL) String() string {
	urlStr := l.url.String()
	if urlStr == "" {
		return fmt.Sprintf("<empty URL object at %p>", l)
	}
	return fmt.Sprintf("<URL: \"%v\" at %p>", urlStr, l)
}

func (l *LoxURL) Type() string {
	return "url"
}
