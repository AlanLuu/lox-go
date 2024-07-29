package ast

import (
	"fmt"
	"net/http"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineHTTPFuncs() {
	className := "http"
	httpClass := NewLoxClass(className, nil, false)
	httpFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native http fn %v at %p>", name, &s)
		}
		httpClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'http.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	httpFunc("get", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 1:
			if loxStr, ok := args[0].(*LoxString); ok {
				url := loxStr.str
				res, err := LoxHTTPGetUrl(url)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return res, nil
			}
			return argMustBeType(in.callToken, "get", "string")
		case 2:
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'http.get' must be a string.")
			}
			if _, ok := args[1].(*LoxDict); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'http.get' must be a dictionary.")
			}

			url := args[0].(*LoxString).str
			headers := args[1].(*LoxDict)
			req, reqErr := http.NewRequest("GET", url, nil)
			if reqErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
			}

			strDictErrMsg := "Headers dictionary in 'http.get' must only have strings."
			it := headers.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				var key, value string
				switch pairKey := pair[0].(type) {
				case *LoxString:
					key = pairKey.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
				}
				switch pairValue := pair[1].(type) {
				case *LoxString:
					value = pairValue.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
				}
				req.Header.Set(key, value)
			}

			res, resErr := LoxHTTPSendRequest(req)
			if resErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, resErr.Error())
			}
			return res, nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
	})
	httpFunc("head", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 1:
			if loxStr, ok := args[0].(*LoxString); ok {
				url := loxStr.str
				res, err := LoxHTTPHeadUrl(url)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return res, nil
			}
			return argMustBeType(in.callToken, "head", "string")
		case 2:
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'http.head' must be a string.")
			}
			if _, ok := args[1].(*LoxDict); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'http.head' must be a dictionary.")
			}

			url := args[0].(*LoxString).str
			headers := args[1].(*LoxDict)
			req, reqErr := http.NewRequest("HEAD", url, nil)
			if reqErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
			}

			strDictErrMsg := "Headers dictionary in 'http.head' must only have strings."
			it := headers.Iterator()
			for it.HasNext() {
				pair := it.Next().(*LoxList).elements
				var key, value string
				switch pairKey := pair[0].(type) {
				case *LoxString:
					key = pairKey.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
				}
				switch pairValue := pair[1].(type) {
				case *LoxString:
					value = pairValue.str
				default:
					return nil, loxerror.RuntimeError(in.callToken, strDictErrMsg)
				}
				req.Header.Set(key, value)
			}

			res, resErr := LoxHTTPSendRequest(req)
			if resErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, resErr.Error())
			}
			return res, nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}
	})

	i.globals.Define(className, httpClass)
}
