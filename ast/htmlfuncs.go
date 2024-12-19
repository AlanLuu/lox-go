package ast

import (
	"fmt"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/net/html"
)

func defineHTMLFields(htmlClass *LoxClass) {
	tokenTypeClass := NewLoxClass("HTML.token", nil, false)
	tokenTypes := map[string]html.TokenType{
		"error":       html.ErrorToken,
		"text":        html.TextToken,
		"startTag":    html.StartTagToken,
		"endTag":      html.EndTagToken,
		"selfClosing": html.SelfClosingTagToken,
		"comment":     html.CommentToken,
		"doctype":     html.DoctypeToken,
	}
	for key, value := range tokenTypes {
		tokenTypeClass.classProperties[key] = int64(value)
	}
	htmlClass.classProperties["token"] = tokenTypeClass
}

func (i *Interpreter) defineHTMLFuncs() {
	className := "HTML"
	htmlClass := NewLoxClass(className, nil, false)
	htmlFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native HTML fn %v at %p>", name, &s)
		}
		htmlClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'HTML.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'HTML.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineHTMLFields(htmlClass)
	htmlFunc("escape", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(html.EscapeString(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "escape", "string")
	})
	htmlFunc("tokenize", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		switch arg := args[0].(type) {
		case *LoxFile:
			if !arg.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot call 'HTML.tokenize' for file not in read mode.")
			}
			return NewLoxHTMLTokenizer(arg.file), nil
		case *LoxString:
			return NewLoxHTMLTokenizerStr(arg.str), nil
		default:
			return argMustBeType(in.callToken, "tokenize", "file or string")
		}
	})
	htmlFunc("tokenType", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			typeStr := html.TokenType(arg).String()
			if strings.HasPrefix(typeStr, "Invalid") {
				typeStr = "Unknown"
			}
			return NewLoxString(typeStr, '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "tokenType", "integer")
	})
	htmlFunc("unescape", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			return NewLoxStringQuote(html.UnescapeString(loxStr.str)), nil
		}
		return argMustBeType(in.callToken, "unescape", "string")
	})

	i.globals.Define(className, htmlClass)
}
