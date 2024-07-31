package ast

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
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
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'http.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	jsonDictToStr := func(in *Interpreter, dict *LoxDict) (string, error) {
		jsonClassErrStr := "Could not find JSON class for stringifying JSON dictionary."
		jsonClassAny, jsonClassErr := in.globals.GetFromStr("JSON")
		if jsonClassErr != nil {
			return "", loxerror.RuntimeError(in.callToken, jsonClassErrStr)
		}
		if _, ok := jsonClassAny.(*LoxClass); !ok {
			return "", loxerror.RuntimeError(in.callToken, jsonClassErrStr)
		}

		jsonClass := jsonClassAny.(*LoxClass)
		jsonStringifyErrStr := "Could not find 'stringify' method in JSON class for stringifying JSON dictionary."
		jsonStringifyFuncAny, foundJsonFunc := jsonClass.classProperties["stringify"]
		if !foundJsonFunc {
			return "", loxerror.RuntimeError(in.callToken, jsonStringifyErrStr)
		}
		if _, ok := jsonStringifyFuncAny.(LoxCallable); !ok {
			return "", loxerror.RuntimeError(in.callToken, jsonStringifyErrStr)
		}

		jsonStringifyFunc := jsonStringifyFuncAny.(LoxCallable)
		argList := list.NewList[any]()
		argList.Add(dict)
		result, resultErr := jsonStringifyFunc.call(in, argList)
		if resultErr != nil {
			errMsg := resultErr.Error()
			index := strings.LastIndex(errMsg, "\n")
			if index > 0 {
				errMsg = errMsg[:index]
			}
			return "", loxerror.RuntimeError(in.callToken,
				"Error occurred when stringifying JSON dictionary:\n"+errMsg)
		}
		return result.(*LoxString).str, nil
	}
	populateHeaders := func(in *Interpreter, headers *LoxDict, req *http.Request, name string) error {
		errMsg := "Headers dictionary in 'http." + name + "' must only have strings."
		it := headers.Iterator()
		for it.HasNext() {
			pair := it.Next().(*LoxList).elements
			var key, value string
			switch pairKey := pair[0].(type) {
			case *LoxString:
				key = pairKey.str
			default:
				return loxerror.RuntimeError(in.callToken, errMsg)
			}
			switch pairValue := pair[1].(type) {
			case *LoxString:
				value = pairValue.str
			default:
				return loxerror.RuntimeError(in.callToken, errMsg)
			}
			req.Header.Set(key, value)
		}
		return nil
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

			headersErr := populateHeaders(in, headers, req, "get")
			if headersErr != nil {
				return nil, headersErr
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

			headersErr := populateHeaders(in, headers, req, "head")
			if headersErr != nil {
				return nil, headersErr
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

			headersErr := populateHeaders(in, headers, req, "post")
			if headersErr != nil {
				return nil, headersErr
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

			headersErr := populateHeaders(in, headers, req, "postForm")
			if headersErr != nil {
				return nil, headersErr
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
	httpFunc("postJSON", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen != 2 && argsLen != 3 {
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 2 or 3 arguments but got %v.", argsLen))
		}
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'http.postJSON' must be a string.")
		}
		switch args[1].(type) {
		case *LoxString:
		case *LoxDict:
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'http.postJSON' must be a string or dictionary.")
		}
		if argsLen == 3 {
			if _, ok := args[2].(*LoxDict); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Third argument to 'http.postJSON' must be a dictionary.")
			}
		}

		urlStr := args[0].(*LoxString).str
		var jsonStr string
		switch secondArg := args[1].(type) {
		case *LoxString:
			jsonStr = secondArg.str
		case *LoxDict:
			var jsonStrErr error
			jsonStr, jsonStrErr = jsonDictToStr(in, secondArg)
			if jsonStrErr != nil {
				return nil, jsonStrErr
			}
		}

		var res *LoxHTTPResponse
		var resErr error
		if argsLen == 3 {
			headers := args[2].(*LoxDict)
			req, reqErr := http.NewRequest("POST", urlStr, strings.NewReader(jsonStr))
			if reqErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
			}
			if len(jsonStr) > 0 {
				req.Header.Set("Content-Type", "application/json")
			}

			headersErr := populateHeaders(in, headers, req, "postJSON")
			if headersErr != nil {
				return nil, headersErr
			}

			res, resErr = LoxHTTPSendRequest(req)
		} else {
			res, resErr = LoxHTTPPostJSONText(urlStr, jsonStr)
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

			headersErr := populateHeaders(in, headers, req, "postText")
			if headersErr != nil {
				return nil, headersErr
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
	httpFunc("request", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		switch argsLen {
		case 3, 4:
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'http.request' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'http.request' must be a string.")
			}

			method := strings.ToUpper(args[0].(*LoxString).str)
			url := args[1].(*LoxString).str
			var req *http.Request
			switch method {
			case "GET", "HEAD":
				if args[2] != nil {
					return nil, loxerror.RuntimeError(in.callToken,
						fmt.Sprintf("Third argument to 'http.request' must be nil for %v requests.", method))
				}
				if argsLen == 4 {
					if _, ok := args[3].(*LoxDict); !ok {
						return nil, loxerror.RuntimeError(in.callToken,
							"Fourth argument to 'http.request' must be a dictionary.")
					}
				}
				var reqErr error
				req, reqErr = http.NewRequest(method, url, nil)
				if reqErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
				}
			default:
				switch args[2].(type) {
				case *LoxBuffer:
				case *LoxDict:
				case *LoxString:
				default:
					return nil, loxerror.RuntimeError(in.callToken,
						"Third argument to 'http.request' must be a buffer, dictionary, or string.")
				}
				if argsLen == 4 {
					if _, ok := args[3].(*LoxDict); !ok {
						return nil, loxerror.RuntimeError(in.callToken,
							"Fourth argument to 'http.request' must be a dictionary.")
					}
				}

				switch thirdArg := args[2].(type) {
				case *LoxBuffer:
					byteArr := list.NewList[byte]()
					for _, element := range thirdArg.elements {
						byteArr.Add(byte(element.(int64)))
					}
					var reqErr error
					if len(byteArr) > 0 {
						req, reqErr = http.NewRequest(method, url, bytes.NewBuffer(byteArr))
						if reqErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
						}
						req.Header.Set("Content-Type", "application/octet-stream")
					} else {
						req, reqErr = http.NewRequest(method, url, nil)
						if reqErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
						}
					}
				case *LoxDict:
					body, bodyErr := jsonDictToStr(in, thirdArg)
					if bodyErr != nil {
						return nil, bodyErr
					}
					var reqErr error
					req, reqErr = http.NewRequest(method, url, strings.NewReader(body))
					if reqErr != nil {
						return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
					}
					req.Header.Set("Content-Type", "application/json")
				case *LoxString:
					body := thirdArg.str
					var reqErr error
					if len(body) > 0 {
						req, reqErr = http.NewRequest(method, url, strings.NewReader(body))
						if reqErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
						}
						req.Header.Set("Content-Type", "text/plain")
					} else {
						req, reqErr = http.NewRequest(method, url, nil)
						if reqErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, reqErr.Error())
						}
					}
				}
			}

			if argsLen == 4 {
				headers := args[3].(*LoxDict)
				headersErr := populateHeaders(in, headers, req, "request")
				if headersErr != nil {
					return nil, headersErr
				}
			}

			res, resErr := LoxHTTPSendRequest(req)
			if resErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, resErr.Error())
			}
			return res, nil
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 3 or 4 arguments but got %v.", argsLen))
		}
	})
	httpFunc("serve", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		var dir string
		var port int64

		argsLen := len(args)
		switch argsLen {
		case 1:
			if portNum, ok := args[0].(int64); ok {
				cwd, cwdErr := os.Getwd()
				if cwdErr != nil {
					dir = "."
				} else {
					dir = cwd
				}
				port = portNum
			} else {
				return argMustBeTypeAn(in.callToken, "serve", "integer")
			}
		case 2:
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"First argument to 'http.serve' must be a string.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'http.serve' must be an integer.")
			}
			dir = args[0].(*LoxString).str
			port = args[1].(int64)
			stat, statErr := os.Stat(dir)
			if statErr != nil {
				return nil, loxerror.RuntimeError(in.callToken, statErr.Error())
			}
			if !stat.IsDir() {
				return nil, loxerror.RuntimeError(in.callToken,
					"'"+dir+"' is not a directory.")
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
		}

		serveMux := NewLoxServeMux()
		fs := http.FileServer(http.Dir(dir))
		serveMux.Handle("/", fs)

		sigChan := make(chan os.Signal, 1)
		closeChan := make(chan struct{}, 1)
		serveChan := make(chan struct{}, 1)
		signal.Notify(sigChan, os.Interrupt)
		defer signal.Stop(sigChan)

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: serveMux,
		}
		fmt.Printf("Serving path '%v' at http://localhost:%d\n", dir, port)
		var serveErr error
		go func() {
			serveErr = srv.ListenAndServe()
			if errors.Is(serveErr, http.ErrServerClosed) {
				closeChan <- struct{}{}
			} else {
				serveChan <- struct{}{}
			}
		}()
		select {
		case <-sigChan:
			srv.Close()
			<-closeChan
		case <-serveChan:
		}

		serveMux.RemoveHandler("/")
		if serveErr != nil {
			return nil, loxerror.RuntimeError(in.callToken, serveErr.Error())
		}
		return nil, nil
	})

	i.globals.Define(className, httpClass)
}
