package ast

import (
	"fmt"
	"os"

	"github.com/AlanLuu/lox/browser"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineWebBrowserFuncs() {
	className := "webbrowser"
	webBrowserClass := NewLoxClass(className, nil, false)
	webBrowserFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native webbrowser fn %v at %p>", name, &s)
		}
		webBrowserClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'webbrowser.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	webBrowserFunc("browsers", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		browserStrings := browser.Browsers()
		browsersList := list.NewListCapDouble[any](int64(len(browserStrings)))
		for _, browserString := range browserStrings {
			browsersList.Add(NewLoxStringQuote(browserString))
		}
		return NewLoxList(browsersList), nil
	})
	webBrowserFunc("commands", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		commands := browser.Commands()
		outerList := list.NewListCapDouble[any](int64(len(commands)))
		for index, outer := range commands {
			if index == 0 && os.Getenv("BROWSER") != "" {
				continue
			}
			innerList := list.NewListCap[any](int64(len(outer)))
			for _, inner := range outer {
				innerList.Add(NewLoxStringQuote(inner))
			}
			outerList.Add(NewLoxList(innerList))
		}
		return NewLoxList(outerList), nil
	})
	webBrowserFunc("mustOpen", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if url, ok := args[0].(*LoxString); ok {
			err := browser.MustOpen(url.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "mustOpen", "string")
	})
	webBrowserFunc("open", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if url, ok := args[0].(*LoxString); ok {
			return browser.Open(url.str), nil
		}
		return argMustBeType(in.callToken, "open", "string")
	})
	webBrowserFunc("other", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		otherStrings := browser.Other()
		otherList := list.NewListCapDouble[any](int64(len(otherStrings)))
		for _, otherString := range otherStrings {
			otherList.Add(NewLoxStringQuote(otherString))
		}
		return NewLoxList(otherList), nil
	})

	i.globals.Define(className, webBrowserClass)
}
