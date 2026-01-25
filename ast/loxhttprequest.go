package ast

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func formDictToStr(dict *LoxDict, name string) (string, error) {
	formValues := url.Values{}
	dictErrMsg := fmt.Sprintf(
		"Form dictionary in 'http request.%v' must only have strings or lists of strings.",
		name,
	)
	it := dict.Iterator()
	for it.HasNext() {
		pair := it.Next().(*LoxList).elements
		var key string

		switch pairKey := pair[0].(type) {
		case *LoxString:
			key = pairKey.str
		default:
			return "", loxerror.Error(dictErrMsg)
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
					return "", loxerror.Error(dictErrMsg)
				}
			}
		default:
			return "", loxerror.Error(dictErrMsg)
		}
	}
	return formValues.Encode(), nil
}

type LoxHTTPRequest struct {
	request *http.Request
	client  *http.Client
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxHTTPRequest(req *http.Request) *LoxHTTPRequest {
	return &LoxHTTPRequest{
		request: req,
		client:  &http.Client{},
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxHTTPRequestURL(url string) (*LoxHTTPRequest, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return NewLoxHTTPRequest(req), nil
}

func EmptyLoxHTTPRequest() (*LoxHTTPRequest, error) {
	return NewLoxHTTPRequestURL("")
}

func (l *LoxHTTPRequest) clearCookies() {
	if l.client.Jar != nil {
		l.client.Jar = nil
	} else {
		l.request.Header.Del("Cookie")
	}
}

func (l *LoxHTTPRequest) header() http.Header {
	if l.request.Header == nil {
		l.request.Header = make(http.Header)
	}
	return l.request.Header
}

func (l *LoxHTTPRequest) initCookieJar() {
	if l.client.Jar == nil {
		jar, _ := cookiejar.New(nil)
		l.client.Jar = jar
	}
}

func (l *LoxHTTPRequest) send() (*LoxHTTPResponse, error) {
	return LoxHTTPSendRequestClient(l.client, l.request)
}

func (l *LoxHTTPRequest) setBodyToReader(reader io.Reader) {
	if reader == nil {
		l.request.Body = nil
		return
	}
	rc, ok := reader.(io.ReadCloser)
	if !ok {
		rc = io.NopCloser(reader)
	}
	l.request.Body = rc
}

func (l *LoxHTTPRequest) setCookie(cookie *http.Cookie) {
	if l.client.Jar != nil {
		l.client.Jar.SetCookies(l.request.URL, []*http.Cookie{cookie})
	} else {
		l.request.AddCookie(cookie)
	}
}

func (l *LoxHTTPRequest) setMethod(method string) {
	l.request.Method = method
}

func (l *LoxHTTPRequest) setNumRedirects(num int64) {
	l.client.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		if int64(len(via)) >= num {
			return http.ErrUseLastResponse
		}
		return nil
	}
}

func (l *LoxHTTPRequest) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if field, ok := l.methods[methodName]; ok {
		return field, nil
	}
	requestFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native http request fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'http request.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'http request.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "body":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxBuffer:
				b := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					b = append(b, byte(element.(int64)))
				}
				l.setBodyToReader(bytes.NewReader(b))
			case *LoxDict:
				formStr, err := formDictToStr(arg, methodName)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.setBodyToReader(strings.NewReader(formStr))
			case *LoxString:
				l.setBodyToReader(strings.NewReader(arg.str))
			case *LoxURLValues:
				l.setBodyToReader(strings.NewReader(arg.values.Encode()))
			default:
				return argMustBeType("buffer, dictionary, string, or url values object")
			}
			return l, nil
		})
	case "bodyClear":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setBodyToReader(nil)
			return l, nil
		})
	case "contentLength":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if contentLength, ok := args[0].(int64); ok {
				l.request.ContentLength = contentLength
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "contentType":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.header().Set("Content-Type", loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "cookieClear":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.clearCookies()
			return l, nil
		})
	case "cookieJar":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.initCookieJar()
			return l, nil
		})
	case "cookieKV":
		return requestFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'http request.cookieKV' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'http request.cookieKV' must be a string.")
			}
			name := args[0].(*LoxString).str
			value := args[1].(*LoxString).str
			l.setCookie(&http.Cookie{
				Name:  name,
				Value: value,
			})
			return l, nil
		})
	case "form":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxBuffer:
				b := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					b = append(b, byte(element.(int64)))
				}
				l.setBodyToReader(bytes.NewReader(b))
			case *LoxDict:
				formStr, err := formDictToStr(arg, methodName)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.setBodyToReader(strings.NewReader(formStr))
			case *LoxString:
				l.setBodyToReader(strings.NewReader(arg.str))
			case *LoxURLValues:
				l.setBodyToReader(strings.NewReader(arg.values.Encode()))
			default:
				return argMustBeType("buffer, dictionary, string, or url values object")
			}
			l.header().Set("Content-Type", "application/x-www-form-urlencoded")
			return l, nil
		})
	case "formData":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxBuffer:
				b := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					b = append(b, byte(element.(int64)))
				}
				l.setBodyToReader(bytes.NewReader(b))
			case *LoxDict:
				formStr, err := formDictToStr(arg, methodName)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.setBodyToReader(strings.NewReader(formStr))
			case *LoxString:
				l.setBodyToReader(strings.NewReader(arg.str))
			case *LoxURLValues:
				l.setBodyToReader(strings.NewReader(arg.values.Encode()))
			default:
				return argMustBeType("buffer, dictionary, string, or url values object")
			}
			l.header().Set("Content-Type", "multipart/form-data")
			return l, nil
		})
	case "headerAdd":
		return requestFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'http request.headerAdd' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'http request.headerAdd' must be a string.")
			}
			l.header().Add(
				args[0].(*LoxString).str,
				args[1].(*LoxString).str,
			)
			return l, nil
		})
	case "headerClear":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.request.Header = nil
			return l, nil
		})
	case "headerDel":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.header().Del(loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "headerSet":
		return requestFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'http request.headerSet' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'http request.headerSet' must be a string.")
			}
			l.header().Set(
				args[0].(*LoxString).str,
				args[1].(*LoxString).str,
			)
			return l, nil
		})
	case "json":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxBuffer:
				b := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					b = append(b, byte(element.(int64)))
				}
				l.setBodyToReader(bytes.NewReader(b))
			case *LoxDict:
				formStr, err := formDictToStr(arg, methodName)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.setBodyToReader(strings.NewReader(formStr))
			case *LoxString:
				l.setBodyToReader(strings.NewReader(arg.str))
			case *LoxURLValues:
				l.setBodyToReader(strings.NewReader(arg.values.Encode()))
			default:
				return argMustBeType("buffer, dictionary, string, or url values object")
			}
			l.header().Set("Content-Type", "application/json")
			return l, nil
		})
	case "method":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.setMethod(loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "methodCONNECT":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("CONNECT")
			return l, nil
		})
	case "methodDELETE":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("DELETE")
			return l, nil
		})
	case "methodGET":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("GET")
			return l, nil
		})
	case "methodHEAD":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("HEAD")
			return l, nil
		})
	case "methodOPTIONS":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("OPTIONS")
			return l, nil
		})
	case "methodPATCH":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("PATCH")
			return l, nil
		})
	case "methodPOST":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("POST")
			return l, nil
		})
	case "methodPUT":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("PUT")
			return l, nil
		})
	case "methodTRACE":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setMethod("TRACE")
			return l, nil
		})
	case "redirects":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if numRedirects, ok := args[0].(int64); ok {
				if numRedirects < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'http request.redirects' cannot be negative.")
				}
				l.setNumRedirects(numRedirects)
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "send":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			res, err := l.send()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return res, nil
		})
	case "sendRet":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			_, err := l.send()
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return l, nil
		})
	case "sendThreads":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if times, ok := args[0].(int64); ok {
				if times > 0 {
					var wg sync.WaitGroup
					for i := int64(0); i < times; i++ {
						wg.Add(1)
						go func(num int64) {
							defer wg.Done()
							_, err := l.send()
							if err != nil {
								fmt.Fprintf(
									os.Stderr,
									"http request.sendThreads: runtime error "+
										"in thread #%v: %v\n",
									num,
									err.Error(),
								)
							}
						}(i + 1)
					}
					wg.Wait()
				}
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "timeout":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				l.client.Timeout = loxDuration.duration
				return l, nil
			}
			return argMustBeType("duration")
		})
	case "timeoutClear":
		return requestFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.client.Timeout = 0
			return l, nil
		})
	case "url":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxString:
				parsedURL, err := url.Parse(arg.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				l.request.URL = parsedURL
			case *LoxURL:
				l.request.URL = arg.url
			default:
				return argMustBeType("string or url object")
			}
			return l, nil
		})
	case "userAgent":
		return requestFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.header().Set("User-Agent", loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	}
	return nil, loxerror.RuntimeError(name, "HTTP requests have no property called '"+methodName+"'.")
}

func (l *LoxHTTPRequest) String() string {
	return fmt.Sprintf("<http request at %p>", l)
}

func (l *LoxHTTPRequest) Type() string {
	return "http request"
}
