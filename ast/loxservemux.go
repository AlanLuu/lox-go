package ast

import (
	"log"
	"net/http"
	"sync"
)

type LoxResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewLoxResponseWriter(w http.ResponseWriter) *LoxResponseWriter {
	return &LoxResponseWriter{w, http.StatusOK}
}

func (l *LoxResponseWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

type LoxServeMux struct {
	mux      *http.ServeMux
	handlers map[string]http.Handler
	mu       sync.RWMutex
}

func NewLoxServeMux() *LoxServeMux {
	return &LoxServeMux{
		mux:      http.NewServeMux(),
		handlers: make(map[string]http.Handler),
	}
}

func (l *LoxServeMux) Handle(pattern string, handler http.Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers[pattern] = handler
	l.mux.Handle(pattern, handler)
}

func (l *LoxServeMux) RemoveHandler(pattern string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.handlers[pattern]; ok {
		delete(l.handlers, pattern)
		l.mux = http.NewServeMux()
		for p, h := range l.handlers {
			l.mux.Handle(p, h)
		}
	}
}

func (l *LoxServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	loxResWriter := NewLoxResponseWriter(w)
	l.mux.ServeHTTP(loxResWriter, r)
	userAgent := r.UserAgent()
	if len(userAgent) > 0 {
		log.Printf(
			"%s %s %s %s %d \"%s\"",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			loxResWriter.status,
			userAgent,
		)
	} else {
		log.Printf(
			"%s %s %s %s %d",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			loxResWriter.status,
		)
	}
}
