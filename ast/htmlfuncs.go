package ast

import (
	"bytes"
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

	nodeTypeClass := NewLoxClass("HTML.node", nil, false)
	nodeTypes := map[string]html.NodeType{
		"error":    html.ErrorNode,
		"text":     html.TextNode,
		"document": html.DocumentNode,
		"element":  html.ElementNode,
		"comment":  html.CommentNode,
		"doctype":  html.DoctypeNode,
		"raw":      html.RawNode,
	}
	for key, value := range nodeTypes {
		nodeTypeClass.classProperties[key] = int64(value)
	}
	htmlClass.classProperties["node"] = nodeTypeClass
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
	htmlFunc("nodeType", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if arg, ok := args[0].(int64); ok {
			typeStr := loxHTMLNodeType(arg).String()
			return NewLoxString(typeStr, '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "nodeType", "integer")
	})
	htmlFunc("parse", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var htmlNode *html.Node
		var err error
		switch arg := args[0].(type) {
		case *LoxFile:
			if !arg.isRead() {
				return nil, loxerror.RuntimeError(in.callToken,
					"Cannot call 'HTML.parse' for file not in read mode.")
			}
			htmlNode, err = html.Parse(arg.file)
		case *LoxString:
			htmlNode, err = html.Parse(strings.NewReader(arg.str))
		default:
			return argMustBeType(in.callToken, "parse", "file or string")
		}
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxHTMLNode(htmlNode), nil
	})
	htmlFunc("render", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxHtmlNode, ok := args[0].(*LoxHTMLNode); ok {
			var builder strings.Builder
			err := html.Render(&builder, loxHtmlNode.current)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(builder.String()), nil
		}
		return argMustBeTypeAn(in.callToken, "render", "html node")
	})
	htmlFunc("renderToBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxHtmlNode, ok := args[0].(*LoxHTMLNode); ok {
			bytesBuffer := new(bytes.Buffer)
			err := html.Render(bytesBuffer, loxHtmlNode.current)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			bytes := bytesBuffer.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		}
		return argMustBeTypeAn(in.callToken, "renderToBuf", "html node")
	})
	htmlFunc("renderToFile", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxHTMLNode); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'HTML.renderToFile' must be an html node.")
		}
		switch arg := args[1].(type) {
		case *LoxFile:
			if !arg.isWrite() && !arg.isAppend() {
				return nil, loxerror.RuntimeError(in.callToken,
					"File argument to 'HTML.renderToFile' must be in write or append mode.")
			}
			loxHtmlNode := args[0].(*LoxHTMLNode)
			err := html.Render(arg.file, loxHtmlNode.current)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'HTML.renderToFile' must be a file.")
		}
		return nil, nil
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
