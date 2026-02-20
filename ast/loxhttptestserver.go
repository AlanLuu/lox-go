package ast

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

const panicRecoverUnknownErr = "unknown error"

type LoxHTTPTestServer struct {
	server     *httptest.Server
	closed     bool
	started    bool
	properties map[string]any
}

func NewLoxHTTPTestServer(
	handler http.Handler,
	server *httptest.Server,
	started bool,
) *LoxHTTPTestServer {
	return &LoxHTTPTestServer{
		server:     server,
		started:    started,
		properties: make(map[string]any),
	}
}

func NewLoxHTTPTestServerStarted(
	handler http.Handler,
) (srv *LoxHTTPTestServer, err error) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = loxerror.Error(s)
			} else {
				err = loxerror.Error(
					fmt.Sprintf(
						"%v in NewLoxHTTPTestServerStarted",
						panicRecoverUnknownErr,
					),
				)
			}
		}
	}()
	srv = NewLoxHTTPTestServer(
		handler,
		httptest.NewServer(handler),
		true,
	)
	return
}

func NewLoxHTTPTestServerUnstarted(
	handler http.Handler,
) (srv *LoxHTTPTestServer, err error) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = loxerror.Error(s)
			} else {
				err = loxerror.Error(
					fmt.Sprintf(
						"%v in NewLoxHTTPTestServerUnstarted",
						panicRecoverUnknownErr,
					),
				)
			}
		}
	}()
	srv = NewLoxHTTPTestServer(
		handler,
		httptest.NewUnstartedServer(handler),
		false,
	)
	return
}

func (l *LoxHTTPTestServer) close() (err error) {
	if l.closed {
		return loxerror.Error("server is already closed.")
	}
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = loxerror.Error(s)
			} else {
				err = loxerror.Error(
					fmt.Sprintf(
						"%v in LoxHTTPTestServer.close",
						panicRecoverUnknownErr,
					),
				)
			}
		}
	}()
	l.server.Close()
	l.closed = true
	return
}

func (l *LoxHTTPTestServer) start() (err error) {
	if l.started {
		return loxerror.Error("server is already started.")
	}
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = loxerror.Error(s)
			} else {
				err = loxerror.Error(
					fmt.Sprintf(
						"%v in LoxHTTPTestServer.start",
						panicRecoverUnknownErr,
					),
				)
			}
		}
	}()
	l.server.Start()
	l.started = true
	return
}

func (l *LoxHTTPTestServer) url() string {
	return l.server.URL
}

func (l *LoxHTTPTestServer) Get(name *token.Token) (any, error) {
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
	serverStrField := func(strFunc func() string) (any, error) {
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
	serverFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native http test server fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	switch lexemeName {
	case "close":
		return serverFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.close(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"http test server.close: %v",
						err.Error(),
					),
				)
			}
			return nil, nil
		})
	case "isClosed":
		return serverFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.closed, nil
		})
	case "isStarted":
		return serverFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.started, nil
		})
	case "start":
		return serverFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if err := l.start(); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"http test server.start: %v",
						err.Error(),
					),
				)
			}
			return nil, nil
		})
	case "url":
		return serverStrField(l.url)
	case "wait":
		return serverFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.closed {
				if !l.started {
					l.start()
				}
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt)
				fmt.Println("Test server listening at " + l.url())
				<-sigChan
				l.close()
			}
			return nil, nil
		})
	case "waitForever":
		return serverFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.closed {
				if !l.started {
					l.start()
				}
				fmt.Println("Test server listening at " + l.url())
				select {}
			}
			return nil, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "HTTP test servers have no property called '"+lexemeName+"'.")
}

func (l *LoxHTTPTestServer) String() string {
	return fmt.Sprintf("<http test server at %p>", l)
}

func (l *LoxHTTPTestServer) Type() string {
	return "http test server"
}
