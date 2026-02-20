package ast

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/net/html"
)

type LoxHTTPHandler struct {
	LoxCallable
	in              *Interpreter
	callToken       *token.Token
	name            string
	logging         bool
	ignoredEndPaths map[string]struct{}
	methods         map[string]*struct{ ProtoLoxCallable }
}

func NewLoxHTTPHandler(in *Interpreter, callable LoxCallable) *LoxHTTPHandler {
	name := ""
	switch callable := callable.(type) {
	case *LoxClass:
		name = callable.name
	case *LoxFunction:
		name = callable.name
	case *LoxHTTPHandler:
		name = callable.name
	}
	return &LoxHTTPHandler{
		LoxCallable:     callable,
		in:              in,
		callToken:       in.callToken,
		name:            name,
		logging:         false,
		ignoredEndPaths: nil,
		methods:         make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxHTTPHandler) handleErr(w http.ResponseWriter, err error) {
	l.printErr(err)
	l.internalErr(w)
}

func (l *LoxHTTPHandler) handleErrStr(w http.ResponseWriter, s string) {
	l.printErrStr(s)
	l.internalErr(w)
}

func (l *LoxHTTPHandler) internalErr(w http.ResponseWriter) {
	http.Error(
		w,
		"internal server error",
		http.StatusInternalServerError,
	)
}

func (l *LoxHTTPHandler) printErr(err error) {
	if l.name == "" {
		fmt.Fprintf(
			os.Stderr,
			"Error in http handler: %v\n",
			err.Error(),
		)
	} else {
		fmt.Fprintf(
			os.Stderr,
			"Error in http handler \"%v\": %v\n",
			l.name,
			err.Error(),
		)
	}
}

func (l *LoxHTTPHandler) printErrStr(s string) {
	if l.name == "" {
		fmt.Fprintf(
			os.Stderr,
			"Error in http handler: %v\n",
			s,
		)
	} else {
		fmt.Fprintf(
			os.Stderr,
			"Error in http handler \"%v\": %v\n",
			l.name,
			s,
		)
	}
}

func (l *LoxHTTPHandler) ignoredEndPathsAdd(path string) {
	l.ignoredEndPathsInit()
	l.ignoredEndPaths[path] = struct{}{}
}

func (l *LoxHTTPHandler) ignoredEndPathsInit() {
	if l.ignoredEndPaths == nil {
		l.ignoredEndPaths = map[string]struct{}{}
	}
}

func (l *LoxHTTPHandler) logIfEnabled(a any) {
	if l.logging {
		fmt.Fprintf(os.Stderr, "http.handler: %v\n", a)
	}
}

func (l *LoxHTTPHandler) strIsIgnoredEndPath(s string) bool {
	if l.ignoredEndPaths == nil {
		return false
	}
	_, ok := l.ignoredEndPaths[s]
	return ok
}

func (l *LoxHTTPHandler) strWouldBeIgnored(s string) (string, bool) {
	if l.ignoredEndPaths == nil {
		return "", false
	}
	_, end := path.Split(s)
	_, ok := l.ignoredEndPaths[end]
	return end, ok
}

func (l *LoxHTTPHandler) urlContainsIgnoredEndPath(u *url.URL) (string, bool) {
	if l.ignoredEndPaths == nil || u == nil {
		return "", false
	}
	_, end := path.Split(u.Path)
	_, ok := l.ignoredEndPaths[end]
	return end, ok
}

func (l *LoxHTTPHandler) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	handlerFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native http handler fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'http handler.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "delAllIgnoredEndPaths":
		return handlerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			clear(l.ignoredEndPaths)
			l.ignoredEndPaths = nil
			return l, nil
		})
	case "delIgnoredEndPath":
		return handlerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				delete(l.ignoredEndPaths, loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "hasIgnoredEndPath":
		return handlerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return l.strIsIgnoredEndPath(loxStr.str), nil
			}
			return argMustBeType("string")
		})
	case "ignoredEndPaths":
		return handlerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			set := EmptyLoxSet()
			for element := range l.ignoredEndPaths {
				_, errStr := set.add(element)
				if len(errStr) > 0 {
					return nil, loxerror.RuntimeError(name, errStr)
				}
			}
			return set, nil
		})
	case "ignoredEndPathsList":
		return handlerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCap[any](int64(len(l.ignoredEndPaths)))
			for element := range l.ignoredEndPaths {
				newList.Add(NewLoxStringQuote(element))
			}
			return NewLoxList(newList), nil
		})
	case "ignoreEndPath":
		return handlerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				l.ignoredEndPathsAdd(loxStr.str)
				return l, nil
			}
			return argMustBeType("string")
		})
	case "ignoreFaviconPath":
		return handlerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.ignoredEndPathsAdd("favicon.ico")
			return l, nil
		})
	case "logging":
		return handlerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if logging, ok := args[0].(bool); ok {
				l.logging = logging
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "wouldIgnorePath":
		return handlerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				_, b := l.strWouldBeIgnored(loxStr.str)
				return b, nil
			}
			return argMustBeType("string")
		})
	}
	return nil, loxerror.RuntimeError(name, "HTTP handlers have no property called '"+methodName+"'.")
}

var (
	loxhttphandler_req *LoxHTTPRequest   = nil
	loxhttphandler_res *LoxHTTPResWriter = nil
)

func (l *LoxHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if end, ok := l.urlContainsIgnoredEndPath(r.URL); ok {
		l.logIfEnabled(
			"ignored end path " + end + " in " + r.URL.Path,
		)
		return
	}

	if loxhttphandler_req == nil {
		loxhttphandler_req = NewLoxHTTPRequest(r)
		defer func() {
			loxhttphandler_req = nil
		}()
	}
	if loxhttphandler_res == nil {
		loxhttphandler_res = NewLoxHTTPResWriter(w)
		defer func() {
			loxhttphandler_res = nil
		}()
	}

	argList := getArgList(l, 2)
	defer argList.Clear()

	argList[0] = loxhttphandler_req
	resWriter := loxhttphandler_res
	argList[1] = resWriter

	result, resultErr := l.call(l.in, argList)
	if resultReturn, ok := result.(Return); ok {
		result = resultReturn.FinalValue
	} else if resultErr != nil {
		l.handleErr(w, resultErr)
		return
	}

	//User called write() method on LoxHTTPResWriter object
	//Ignore the return value of the callback
	if resWriter.calledWrite {
		return
	}

	//w.Write() method will detect the content type
	//and set the appropriate Content-Type header
	switch result := result.(type) {
	case nil:
	case *LoxBuffer:
		data := make([]byte, 0, len(result.elements))
		for _, element := range result.elements {
			data = append(data, byte(element.(int64)))
		}
		if _, err := w.Write(data); err != nil {
			l.printErr(err)
			return
		}
	case *LoxDict:
		jsonStr, jsonStrErr, test := jsonDictToStr(l.in, result)
		if jsonStrErr != nil {
			if test {
				l.handleErr(w, loxerror.RuntimeError(l.callToken, jsonStrErr.Error()))
			} else {
				l.handleErr(w, jsonStrErr)
			}
			return
		}
		resWriter.setContentTypeIfNotSet("application/json")
		if _, err := io.WriteString(w, jsonStr); err != nil {
			l.printErr(err)
			return
		}
	case *LoxFile:
		if !result.isRead() {
			l.handleErrStr(w, "Returned file must be in read mode.")
			return
		}
		if _, err := io.Copy(w, result.file); err != nil {
			l.printErr(err)
			return
		}
	case *LoxHTMLNode:
		resWriter.setContentTypeIfNotSet("text/html; charset=utf-8")
		if _, err := io.WriteString(w, "<!DOCTYPE html>"); err != nil {
			//We can still attempt to write the rest of the html
			//even if there's an error here, so no need to return early
			l.printErr(err)
		}
		if err := html.Render(w, result.current); err != nil {
			l.printErr(err)
			return
		}
	case *LoxString:
		resWriter.setContentTypeIfNotSet("text/plain")
		if _, err := io.WriteString(w, result.str); err != nil {
			l.printErr(err)
			return
		}
	case http.Handler:
		result.ServeHTTP(w, r)
	default:
		resWriter.setContentTypeIfNotSet("text/plain")
		if _, err := io.WriteString(w, getResult(result, result, true)); err != nil {
			l.printErr(err)
			return
		}
	}
}

func (l *LoxHTTPHandler) String() string {
	if l.name == "" {
		return fmt.Sprintf("<http handler fn at %p>", l)
	}
	return fmt.Sprintf("<http handler fn %v at %p>", l.name, l)
}

func (l *LoxHTTPHandler) Type() string {
	return "http handler"
}
