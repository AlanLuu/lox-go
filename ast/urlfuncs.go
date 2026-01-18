package ast

import (
	"fmt"
	"net/url"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineURLFuncs() {
	className := "URL"
	urlClass := NewLoxClass(className, nil, false)
	urlFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native URL fn %v at %p>", name, &s)
		}
		urlClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'URL.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	urlFunc("dictQuery", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDict, ok := args[0].(*LoxDict); ok {
			const errMsg1 = "Dictionary argument to 'URL.dictQuery' must only have string keys."
			const errMsg2 = "Dictionary argument to 'URL.dictQuery' must only have string or list values."
			dictQuery := map[string][]string{}
			it := loxDict.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements

				var key string
				switch pairKey := pair[0].(type) {
				case *LoxString:
					key = pairKey.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, errMsg1)
				}

				var value []string
				switch pairValue := pair[1].(type) {
				case *LoxString:
					value = []string{pairValue.str}
				case *LoxList:
					value = make([]string, 0, len(pairValue.elements))
					for index, element := range pairValue.elements {
						switch element := element.(type) {
						case *LoxString:
							value = append(value, element.str)
						default:
							value = nil
							return nil, loxerror.RuntimeError(
								in.callToken,
								fmt.Sprintf(
									"List element corresponding to dictionary key "+
										"'%v' at index %v must be a string.",
									key,
									index,
								),
							)
						}
					}
				default:
					return nil, loxerror.RuntimeError(in.callToken, errMsg2)
				}

				dictQuery[key] = value
			}
			return NewLoxURLValues(url.Values(dictQuery)), nil
		}
		return argMustBeType(in.callToken, "dictQuery", "dictionary")
	})
	urlFunc("joinPath", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"URL.joinPath: expected at least 1 argument but got 0.")
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'URL.joinPath' must be a string.")
		}
		var strArgs []string
		for index, arg := range args[1:] {
			switch arg := arg.(type) {
			case *LoxString:
				if index == 0 {
					strArgs = make([]string, 0, argsLen)
				}
				strArgs = append(strArgs, arg.str)
			default:
				strArgs = nil
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"Argument #%v in 'URL.joinPath' must be a string.",
						index+2,
					),
				)
			}
		}
		base := args[0].(*LoxString).str
		newUrl, err := url.JoinPath(base, strArgs...)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxStringQuote(newUrl), nil
	})
	urlFunc("parse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			loxUrl, err := NewLoxURLStr(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return loxUrl, nil
		}
		return argMustBeType(in.callToken, "parse", "string")
	})
	urlFunc("parseURI", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			loxUrl, err := NewLoxURLURIStr(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return loxUrl, nil
		}
		return argMustBeType(in.callToken, "parseURI", "string")
	})
	urlFunc("pathEscape", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(url.PathEscape(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "pathEscape", "string")
	})
	urlFunc("pathUnescape", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			s, err := url.PathUnescape(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(s), nil
		}
		return argMustBeType(in.callToken, "pathUnescape", "string")
	})
	urlFunc("query", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			query, err := url.ParseQuery(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxURLValues(query), nil
		}
		return argMustBeType(in.callToken, "query", "string")
	})
	urlFunc("queryDict", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			query, err := url.ParseQuery(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			dict := EmptyLoxDict()
			for key, values := range query {
				inner := list.NewListCap[any](int64(len(values)))
				for _, value := range values {
					inner.Add(NewLoxStringQuote(value))
				}
				dict.setKeyValue(NewLoxStringQuote(key), NewLoxList(inner))
			}
			return dict, nil
		}
		return argMustBeType(in.callToken, "queryDict", "string")
	})
	urlFunc("queryEmpty", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return EmptyLoxURLValues(), nil
	})
	urlFunc("queryEscape", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(url.QueryEscape(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "queryEscape", "string")
	})
	urlFunc("queryUnescape", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			s, err := url.QueryUnescape(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(s), nil
		}
		return argMustBeType(in.callToken, "queryUnescape", "string")
	})

	i.globals.Define(className, urlClass)
}
