package ast

import (
	"fmt"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/net/html"
)

type LoxHTMLToken struct {
	token      html.Token
	properties map[string]any
}

func NewLoxHTMLToken(token html.Token) *LoxHTMLToken {
	return &LoxHTMLToken{
		token:      token,
		properties: make(map[string]any),
	}
}

func (l *LoxHTMLToken) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	htmlTokenField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	switch lexemeName {
	case "attributes":
		attributesLen := len(l.token.Attr)
		attributesList := list.NewListCap[any](int64(attributesLen))
		for i := 0; i < attributesLen; i++ {
			attributesList.Add(NewLoxHTMLAttribute(l.token.Attr[i]))
		}
		return htmlTokenField(NewLoxList(attributesList))
	case "data":
		return htmlTokenField(NewLoxStringQuote(l.token.Data))
	case "tag":
		return htmlTokenField(NewLoxString(l.token.DataAtom.String(), '\''))
	case "type":
		return htmlTokenField(int64(l.token.Type))
	case "typeStr":
		typeStr := l.token.Type.String()
		if strings.HasPrefix(typeStr, "Invalid") {
			typeStr = "Unknown"
		}
		return htmlTokenField(NewLoxString(typeStr, '\''))
	}
	return nil, loxerror.RuntimeError(name, "HTML tokens have no property called '"+lexemeName+"'.")
}

func (l *LoxHTMLToken) String() string {
	tokenType := l.token.Type
	tokenTypeStr := tokenType.String()
	switch tokenType {
	case html.StartTagToken, html.EndTagToken:
		tagName := l.token.Data
		return fmt.Sprintf("<HTML %v token \"%v\" at %p>", tokenTypeStr, tagName, l)
	}
	return fmt.Sprintf("<HTML %v token at %p>", tokenTypeStr, l)
}

func (l *LoxHTMLToken) Type() string {
	return "html token"
}
