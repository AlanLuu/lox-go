package ast

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/net/html"
)

type LoxHTTPResWriter struct {
	w http.ResponseWriter

	//True if the Content-Type header has been set
	c_type_isSet bool

	//True if user called write() method on LoxHTTPResWriter object
	calledWrite bool

	properties map[string]any
}

func NewLoxHTTPResWriter(w http.ResponseWriter) *LoxHTTPResWriter {
	return &LoxHTTPResWriter{
		w:            w,
		c_type_isSet: false,
		calledWrite:  false,
		properties:   make(map[string]any),
	}
}

func (l *LoxHTTPResWriter) cookieSetNameValue(name, value string) {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}
	http.SetCookie(l.w, cookie)
}

func (l *LoxHTTPResWriter) headerAdd(key, value string) {
	l.setIsContentType(key, true)
	l.w.Header().Add(key, value)
}

func (l *LoxHTTPResWriter) headerAddList(key string, values []string) {
	l.setIsContentType(key, true)
	header := l.w.Header()
	if arr, ok := header[key]; ok {
		if len(values) > 0 {
			header[key] = append(arr, values...)
		}
	} else {
		header[key] = values
	}
}

func (l *LoxHTTPResWriter) headerClear() {
	clear(l.w.Header())
	l.c_type_isSet = false
}

func (l *LoxHTTPResWriter) headerDel(key string) {
	l.setIsContentType(key, false)
	l.w.Header().Del(key)
}

func (l *LoxHTTPResWriter) headerSet(key, value string) {
	l.setIsContentType(key, true)
	l.w.Header().Set(key, value)
}

func (l *LoxHTTPResWriter) headerSetList(key string, values []string) {
	l.setIsContentType(key, true)
	l.w.Header()[key] = values
}

func (l *LoxHTTPResWriter) setIsContentType(key string, to bool) {
	const header = "Content-Type"
	if to {
		if !l.c_type_isSet {
			l.c_type_isSet = key == header
		}
	} else {
		if l.c_type_isSet {
			l.c_type_isSet = key != header
		}
	}
}

func (l *LoxHTTPResWriter) setContentTypeIfNotSet(value string) {
	if !l.c_type_isSet {
		l.w.Header().Set("Content-Type", value)
	}
}

func (l *LoxHTTPResWriter) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	makeFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) *struct{ ProtoLoxCallable } {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native http response writer fn %v at %p>", lexemeName, s)
		}
		return s
	}
	writerFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		f := makeFunc(arity, method)
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = f
		}
		return f, nil
	}
	writerField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'http response writer.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'http response writer.%v' must be an %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch lexemeName {
	case "cookie":
		return writerFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'http response writer.cookie' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'http response writer.cookie' must be a string.")
			}
			l.cookieSetNameValue(
				args[0].(*LoxString).str,
				args[1].(*LoxString).str,
			)
			return l, nil
		})
	case "header":
		headerArgMustBeType := func(a, b string) (any, error) {
			errStr := fmt.Sprintf("Argument to 'http response writer.header.%v' must be a %v.", a, b)
			return nil, loxerror.RuntimeError(name, errStr)
		}
		var o *LoxObjectType
		o = NewLoxObjectType(map[string]any{
			"add": makeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'http response writer.header.add' must be a string.")
				}
				if _, ok := args[1].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'http response writer.header.add' must be a string.")
				}
				l.headerAdd(
					args[0].(*LoxString).str,
					args[1].(*LoxString).str,
				)
				return o, nil
			}),
			"addArgs": makeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
				argsLen := len(args)
				if argsLen < 2 {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"http response writer.header.addArgs: "+
								"expected at least 2 arguments but got %v.",
							argsLen,
						),
					)
				}
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'http response writer.header.addArgs' must be a string.")
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
							name,
							fmt.Sprintf(
								"Argument #%v in 'http response writer.header.addArgs' must be a string.",
								index+2,
							),
						)
					}
				}
				key := args[0].(*LoxString).str
				l.headerAddList(key, strArgs)
				return o, nil
			}),
			"addList": makeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'http response writer.header.addList' must be a string.")
				}
				if _, ok := args[1].(*LoxList); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'http response writer.header.addList' must be a list.")
				}
				elements := args[1].(*LoxList).elements
				var strArgs []string
				for index, element := range elements {
					switch element := element.(type) {
					case *LoxString:
						if index == 0 {
							strArgs = make([]string, 0, len(elements))
						}
						strArgs = append(strArgs, element.str)
					default:
						strArgs = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"http response writer.header.addList: "+
									"list element at index %v must be a string.",
								index,
							),
						)
					}
				}
				key := args[0].(*LoxString).str
				l.headerAddList(key, strArgs)
				return o, nil
			}),
			"clear": makeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
				l.headerClear()
				return o, nil
			}),
			"contains": makeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
				if loxStr, ok := args[0].(*LoxString); ok {
					_, ok := l.w.Header()[loxStr.str]
					return ok, nil
				}
				return headerArgMustBeType("contains", "string")
			}),
			"del": makeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
				if loxStr, ok := args[0].(*LoxString); ok {
					l.headerDel(loxStr.str)
					return o, nil
				}
				return headerArgMustBeType("del", "string")
			}),
			"get": makeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
				if loxStr, ok := args[0].(*LoxString); ok {
					return NewLoxStringQuote(l.w.Header().Get(loxStr.str)), nil
				}
				return headerArgMustBeType("get", "string")
			}),
			"getList": makeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
				if loxStr, ok := args[0].(*LoxString); ok {
					values, ok := l.w.Header()[loxStr.str]
					if !ok || values == nil {
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"http response writer.header.getList: "+
									"unknown header key '%v'.",
								loxStr.str,
							),
						)
					}
					if len(values) == 0 {
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"http response writer.header.getList: "+
									"no values associated with header key '%v'.",
								loxStr.str,
							),
						)
					}
					valuesList := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						valuesList.Add(NewLoxStringQuote(value))
					}
					return NewLoxList(valuesList), nil
				}
				return headerArgMustBeType("getList", "string")
			}),
			"has": makeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
				if loxStr, ok := args[0].(*LoxString); ok {
					_, ok := l.w.Header()[loxStr.str]
					return ok, nil
				}
				return headerArgMustBeType("has", "string")
			}),
			"set": makeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'http response writer.header.set' must be a string.")
				}
				if _, ok := args[1].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'http response writer.header.set' must be a string.")
				}
				l.headerSet(
					args[0].(*LoxString).str,
					args[1].(*LoxString).str,
				)
				return o, nil
			}),
			"setArgs": makeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
				argsLen := len(args)
				if argsLen < 2 {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"http response writer.header.setArgs: "+
								"expected at least 2 arguments but got %v.",
							argsLen,
						),
					)
				}
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'http response writer.header.setArgs' must be a string.")
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
							name,
							fmt.Sprintf(
								"Argument #%v in 'http response writer.header.setArgs' must be a string.",
								index+2,
							),
						)
					}
				}
				key := args[0].(*LoxString).str
				l.headerSetList(key, strArgs)
				return o, nil
			}),
			"setKeyToEmptyList": makeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
				if loxStr, ok := args[0].(*LoxString); ok {
					l.headerSetList(loxStr.str, []string{})
					return l, nil
				}
				return headerArgMustBeType("setKeyToEmptyList", "string")
			}),
			"setList": makeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'http response writer.header.setList' must be a string.")
				}
				if _, ok := args[1].(*LoxList); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'http response writer.header.setList' must be a list.")
				}
				key := args[0].(*LoxString).str
				elements := args[1].(*LoxList).elements
				if len(elements) == 0 {
					l.headerSetList(key, []string{})
					return o, nil
				}
				var strArgs []string
				for index, element := range elements {
					switch element := element.(type) {
					case *LoxString:
						if index == 0 {
							strArgs = make([]string, 0, len(elements))
						}
						strArgs = append(strArgs, element.str)
					default:
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"http response writer.header.setList: "+
									"list element at index %v must be a string.",
								index,
							),
						)
					}
				}
				l.headerSetList(key, strArgs)
				return o, nil
			}),
			"toDict": makeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
				dict := EmptyLoxDict()
				for key, values := range l.w.Header() {
					if len(values) == 0 {
						continue
					}
					dict.setKeyValue(NewLoxStringQuote(key), NewLoxStringQuote(values[0]))
				}
				return dict, nil
			}),
			"toDictList": makeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
				dict := EmptyLoxDict()
				for key, values := range l.w.Header() {
					inner := list.NewListCap[any](int64(len(values)))
					for _, value := range values {
						inner.Add(NewLoxStringQuote(value))
					}
					dict.setKeyValue(NewLoxStringQuote(key), NewLoxList(inner))
				}
				return dict, nil
			}),
		}).setName(l.Type())
		return writerField(o)
	case "status":
		return writerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if statusCode, ok := args[0].(int64); ok {
				l.w.WriteHeader(int(statusCode))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "write", "send":
		return writerFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case *LoxBuffer:
				data := make([]byte, 0, len(arg.elements))
				for _, element := range arg.elements {
					data = append(data, byte(element.(int64)))
				}
				l.w.Write(data)
			case *LoxDict:
				jsonStr, jsonStrErr, test := jsonDictToStr(i, arg)
				if jsonStrErr != nil {
					if test {
						return nil, loxerror.RuntimeError(name, jsonStrErr.Error())
					}
					return nil, jsonStrErr
				}
				l.setContentTypeIfNotSet("application/json")
				l.w.Write([]byte(jsonStr))
			case *LoxFile:
				if !arg.isRead() {
					return nil, loxerror.RuntimeError(
						name,
						"File argument to 'http response writer.write' "+
							"must be in read mode.",
					)
				}
				io.Copy(l.w, arg.file)
			case *LoxHTMLNode:
				l.setContentTypeIfNotSet("text/html; charset=utf-8")
				l.w.Write([]byte("<!DOCTYPE html>"))
				html.Render(l.w, arg.current)
			case *LoxString:
				l.setContentTypeIfNotSet("text/plain")
				l.w.Write([]byte(arg.str))
			default:
				return argMustBeType(
					"buffer, dictionary, file, HTML node, or string",
				)
			}
			l.calledWrite = true
			return l, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "HTTP response writers have no property called '"+lexemeName+"'.")
}

func (l *LoxHTTPResWriter) String() string {
	return fmt.Sprintf("<http response writer at %p>", l)
}

func (l *LoxHTTPResWriter) Type() string {
	return "http response writer"
}
