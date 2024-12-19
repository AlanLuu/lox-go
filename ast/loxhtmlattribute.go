package ast

import (
	"fmt"
	"strings"

	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	"golang.org/x/net/html"
)

type LoxHTMLAttribute struct {
	attribute  html.Attribute
	properties map[string]any
}

func NewLoxHTMLAttribute(attribute html.Attribute) *LoxHTMLAttribute {
	return &LoxHTMLAttribute{
		attribute:  attribute,
		properties: make(map[string]any),
	}
}

func (l *LoxHTMLAttribute) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	htmlAttributeField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	switch lexemeName {
	case "key":
		return htmlAttributeField(NewLoxStringQuote(l.attribute.Key))
	case "value":
		return htmlAttributeField(NewLoxStringQuote(l.attribute.Val))
	}
	return nil, loxerror.RuntimeError(name, "HTML attributes have no property called '"+lexemeName+"'.")
}

func (l *LoxHTMLAttribute) String() string {
	key := l.attribute.Key
	if strings.Contains(key, " ") {
		key = "'" + key + "'"
	}
	value := l.attribute.Val
	if strings.Contains(value, " ") {
		value = "'" + value + "'"
	}
	return fmt.Sprintf("<HTML attribute %v=%v at %p>", key, value, l)
}

func (l *LoxHTMLAttribute) Type() string {
	return "html attribute"
}
