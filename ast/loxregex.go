package ast

import (
	"fmt"
	"regexp"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxRegex struct {
	regex      *regexp.Regexp
	properties map[string]any
}

func NewLoxRegex(regex *regexp.Regexp) *LoxRegex {
	return &LoxRegex{
		regex:      regex,
		properties: make(map[string]any),
	}
}

func NewLoxRegexStr(pattern string) (*LoxRegex, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return NewLoxRegex(regex), nil
}

func (l *LoxRegex) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if method, ok := l.properties[lexemeName]; ok {
		return method, nil
	}
	regexField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	regexFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native regex fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'regex.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "findAll":
		return regexFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				matches := l.regex.FindAllString(loxStr.str, -1)
				matchesList := list.NewListCap[any](int64(len(matches)))
				for _, match := range matches {
					matchesList.Add(NewLoxStringQuote(match))
				}
				return NewLoxList(matchesList), nil
			}
			return argMustBeType("string")
		})
	case "findAllGroups":
		return regexFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				matches := l.regex.FindAllStringSubmatch(loxStr.str, -1)
				matchesList := list.NewListCap[any](int64(len(matches)))
				for _, match := range matches {
					innerList := list.NewListCap[any](int64(len(match)))
					for _, str := range match {
						innerList.Add(NewLoxStringQuote(str))
					}
					matchesList.Add(NewLoxList(innerList))
				}
				return NewLoxList(matchesList), nil
			}
			return argMustBeType("string")
		})
	case "numSubexp":
		return regexField(int64(l.regex.NumSubexp()))
	case "pattern":
		return regexField(NewLoxStringQuote(l.regex.String()))
	case "replace":
		return regexFunc(2, func(in *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'regex.replace' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'regex.replace' must be a string.")
			}
			str := args[0].(*LoxString).str
			repl := args[1].(*LoxString).str
			return NewLoxStringQuote(l.regex.ReplaceAllLiteralString(str, repl)), nil
		})
	case "replacen":
		return regexFunc(2, func(in *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'regex.replacen' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'regex.replacen' must be a string.")
			}
			str := args[0].(*LoxString).str
			repl := args[1].(*LoxString).str
			var replaceCount int64 = 0
			replaced := l.regex.ReplaceAllStringFunc(str, func(string) string {
				replaceCount++
				return repl
			})
			elements := list.NewListCap[any](2)
			elements.Add(NewLoxStringQuote(replaced))
			elements.Add(replaceCount)
			return NewLoxList(elements), nil
		})
	case "split":
		return regexFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				split := l.regex.Split(loxStr.str, -1)
				splitList := list.NewListCap[any](int64(len(split)))
				for _, str := range split {
					splitList.Add(NewLoxStringQuote(str))
				}
				return NewLoxList(splitList), nil
			}
			return argMustBeType("string")
		})
	case "test":
		return regexFunc(1, func(in *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return l.regex.MatchString(loxStr.str), nil
			}
			return argMustBeType("string")
		})
	}
	return nil, loxerror.RuntimeError(name, "Regexes have no property called '"+lexemeName+"'.")
}

func (l *LoxRegex) String() string {
	return fmt.Sprintf("<regex pattern='%v' at %p>", l.regex.String(), l)
}

func (l *LoxRegex) Type() string {
	return "regex"
}
