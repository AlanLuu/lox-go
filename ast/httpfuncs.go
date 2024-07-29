package ast

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
	httpFunc("post", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 1:
			if loxStr, ok := args[0].(*LoxString); ok {
				url := loxStr.str
				res, err := LoxHTTPPostUrl(url)
				if err != nil {
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				return res, nil
			}
			return argMustBeType(in.callToken, "post", "string")
		case 2:
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'http.post' must be a string.")
			}
			if _, ok := args[1].(*LoxDict); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'http.post' must be a dictionary.")
			}

			url := args[0].(*LoxString).str
			headers := args[1].(*LoxDict)
			req, reqErr := http.NewRequest("POST", url, nil)
			if reqErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
			}

			strDictErrMsg := "Headers dictionary in 'http.post' must only have strings."
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
	httpFunc("postForm", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 2 && argsLen != 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'http.postForm' must be a string.")
		}
		if _, ok := args[1].(*LoxDict); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'http.postForm' must be a dictionary.")
		}
		if argsLen == 3 {
			if _, ok := args[2].(*LoxDict); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Third argument to 'http.postForm' must be a dictionary.")
			}
		}

		urlStr := args[0].(*LoxString).str
		formDict := args[1].(*LoxDict)
		formValues := url.Values{}
		formDictErrMsg := "Form dictionary in 'http.postForm' must only have strings or lists of strings."
		formDictIterator := formDict.Iterator()
		for formDictIterator.HasNext() {
			pair := formDictIterator.Next().(*LoxList).elements
			var key string

			switch pairKey := pair[0].(type) {
			case *LoxString:
				key = pairKey.str
			default:
				return nil, loxerror.RuntimeError(in.callToken, formDictErrMsg)
			}

			switch pairValue := pair[1].(type) {
			case *LoxString:
				formValues.Add(key, pairValue.str)
			case *LoxList:
				for _, element := range pairValue.elements {
					switch element := element.(type) {
					case *LoxString:
						formValues.Add(key, element.str)
					default:
						return nil, loxerror.RuntimeError(in.callToken, formDictErrMsg)
					}
				}
			default:
				return nil, loxerror.RuntimeError(in.callToken, formDictErrMsg)
			}
		}

		var res *LoxHTTPResponse
		var resErr error
		if argsLen == 3 {
			headers := args[2].(*LoxDict)
			req, reqErr := http.NewRequest("POST", urlStr, strings.NewReader(formValues.Encode()))
			if reqErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			strDictErrMsg := "Headers dictionary in 'http.postForm' must only have strings."
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

			res, resErr = LoxHTTPSendRequest(req)
		} else {
			res, resErr = LoxHTTPPostForm(urlStr, formValues)
		}

		if resErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, resErr.Error())
		}
		return res, nil
	})
	httpFunc("postText", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 2 && argsLen != 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'http.postText' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'http.postText' must be a string.")
		}
		if argsLen == 3 {
			if _, ok := args[2].(*LoxDict); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Third argument to 'http.postText' must be a dictionary.")
			}
		}

		urlStr := args[0].(*LoxString).str
		bodyStr := args[1].(*LoxString).str
		var res *LoxHTTPResponse
		var resErr error
		if argsLen == 3 {
			headers := args[2].(*LoxDict)
			req, reqErr := http.NewRequest("POST", urlStr, strings.NewReader(bodyStr))
			if reqErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
			}
			if len(bodyStr) > 0 {
				req.Header.Set("Content-Type", "text/plain")
			}

			strDictErrMsg := "Headers dictionary in 'http.postText' must only have strings."
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

			res, resErr = LoxHTTPSendRequest(req)
		} else {
			res, resErr = LoxHTTPPostText(urlStr, bodyStr)
		}

		if resErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, resErr.Error())
		}
		return res, nil
	})

	i.globals.Define(className, httpClass)
}
