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

const (
	loxHTMLNodeParent uint8 = iota
	loxHTMLNodeFirstChild
	loxHTMLNodeLastChild
	loxHTMLNodePrevSibling
	loxHTMLNodeNextSibling
)

type loxHTMLNodeType html.NodeType

func (l loxHTMLNodeType) String() string {
	switch html.NodeType(l) {
	case html.ErrorNode:
		return "Error"
	case html.TextNode:
		return "Text"
	case html.DocumentNode:
		return "Document"
	case html.ElementNode:
		return "Element"
	case html.CommentNode:
		return "Comment"
	case html.DoctypeNode:
		return "Doctype"
	case html.RawNode:
		return "Raw"
	}
	return "Unknown"
}

type loxHTMLNodeTypeStr string

func (l loxHTMLNodeTypeStr) nodeType() int8 {
	switch strings.ToLower(string(l)) {
	case "error":
		return int8(html.ErrorNode)
	case "text":
		return int8(html.TextNode)
	case "document":
		return int8(html.DocumentNode)
	case "element":
		return int8(html.ElementNode)
	case "comment":
		return int8(html.CommentNode)
	case "doctype":
		return int8(html.DoctypeNode)
	case "raw":
		return int8(html.RawNode)
	}
	return -1
}

type LoxHTMLNode struct {
	current *html.Node
	family  [5]struct {
		cachedNode   *LoxHTMLNode
		familyMember *html.Node
	}
	properties map[string]any
}

func NewLoxHTMLNode(htmlNode *html.Node) *LoxHTMLNode {
	if htmlNode == nil {
		panic("in NewLoxHTMLNode: node argument is nil")
	}
	node := &LoxHTMLNode{
		current:    htmlNode,
		properties: make(map[string]any),
	}
	node.family[loxHTMLNodeParent].familyMember = node.current.Parent
	node.family[loxHTMLNodeFirstChild].familyMember = node.current.FirstChild
	node.family[loxHTMLNodeLastChild].familyMember = node.current.LastChild
	node.family[loxHTMLNodePrevSibling].familyMember = node.current.PrevSibling
	node.family[loxHTMLNodeNextSibling].familyMember = node.current.NextSibling
	return node
}

func (l *LoxHTMLNode) forEachDescendent(callback func(*html.Node)) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		callback(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(l.current)
}

func (l *LoxHTMLNode) getNode(index uint8) *LoxHTMLNode {
	familyStruct := l.family[index]
	if familyStruct.familyMember == nil {
		return nil
	}
	if familyStruct.cachedNode != nil {
		return familyStruct.cachedNode
	}
	newNode := NewLoxHTMLNode(familyStruct.familyMember)
	l.family[index].cachedNode = newNode
	return newNode
}

func (l *LoxHTMLNode) isRootNode() bool {
	return l.family[loxHTMLNodeParent].familyMember == nil
}

func (l *LoxHTMLNode) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	htmlNodeFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native HTML node fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'HTML node.%v' must be a %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'HTML node.%v' must be an %v.", lexemeName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	fieldAliases := map[string]string{
		"firstChild": "fc",
		"fc":         "firstChild",

		"lastChild": "lc",
		"lc":        "lastChild",

		"nextSibling": "ns",
		"ns":          "nextSibling",

		"parent": "p",
		"p":      "parent",

		"prevSibling": "ps",
		"ps":          "prevSibling",
	}
	htmlNodeField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
			if aliasName, ok := fieldAliases[lexemeName]; ok {
				if _, ok := l.properties[aliasName]; !ok {
					l.properties[aliasName] = field
				}
			}
		}
		return field, nil
	}
	getHTMLNode := func(index uint8) (any, error) {
		htmlNode := l.getNode(index)
		if htmlNode == nil {
			return htmlNodeField(nil)
		}
		return htmlNode, nil
	}
	switch lexemeName {
	case "ancestors":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			ancestors := list.NewList[any]()
			for p := l.current.Parent; p != nil; p = p.Parent {
				ancestors.Add(NewLoxHTMLNode(p))
			}
			return NewLoxList(ancestors), nil
		})
	case "ancestorsIter":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			p := l.current
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				p = p.Parent
				return p != nil
			}
			iterator.nextMethod = func() any {
				return NewLoxHTMLNode(p)
			}
			return NewLoxIterator(iterator), nil
		})
	case "attributes":
		attributesLen := len(l.current.Attr)
		attributesList := list.NewListCap[any](int64(attributesLen))
		for i := 0; i < attributesLen; i++ {
			attributesList.Add(NewLoxHTMLAttribute(l.current.Attr[i]))
		}
		return htmlNodeField(NewLoxList(attributesList))
	case "children":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			children := list.NewList[any]()
			for c := l.current.FirstChild; c != nil; c = c.NextSibling {
				children.Add(NewLoxHTMLNode(c))
			}
			return NewLoxList(children), nil
		})
	case "childrenIter":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			c := l.current
			firstIteration := true
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				if firstIteration {
					c = c.FirstChild
					firstIteration = false
				} else {
					c = c.NextSibling
				}
				return c != nil
			}
			iterator.nextMethod = func() any {
				return NewLoxHTMLNode(c)
			}
			return NewLoxIterator(iterator), nil
		})
	case "commentNodesByContent":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				commentNodes := list.NewList[any]()
				l.forEachDescendent(func(n *html.Node) {
					if n.Type == html.CommentNode && n.Data == str {
						commentNodes.Add(NewLoxHTMLNode(n))
					}
				})
				return NewLoxList(commentNodes), nil
			}
			return argMustBeType("string")
		})
	case "commentNodesByContentIter":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				stack := list.NewList[*html.Node]()
				stack.Add(l.current)
				firstIteration := true
				condition := func() bool {
					e := stack.Peek()
					return e.Type == html.CommentNode && e.Data == str
				}
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if !firstIteration && condition() {
						stack[len(stack)-1] = stack.Peek().NextSibling
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					for len(stack) > 0 && !condition() {
						if !firstIteration {
							stack.Add(stack.Peek().FirstChild)
						} else {
							firstIteration = false
						}
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					if firstIteration {
						firstIteration = false
					}
					return len(stack) > 0
				}
				iterator.nextMethod = func() any {
					htmlNode := stack.Peek()
					return NewLoxHTMLNode(htmlNode)
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeType("string")
		})
	case "data":
		return htmlNodeField(NewLoxStringQuote(l.current.Data))
	case "descendents":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			descendents := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				descendents.Add(NewLoxHTMLNode(n))
			})
			return NewLoxList(descendents), nil
		})
	case "descendentsIter", "dfsIter":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			stack := list.NewList[*html.Node]()
			stack.Add(l.current)
			firstIteration := true
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				if !firstIteration {
					stack.Add(stack.Peek().FirstChild)
				} else {
					firstIteration = false
				}
				for len(stack) > 0 && stack.Peek() == nil {
					stack.Pop()
					if len(stack) > 0 && stack.Peek() != nil {
						stack[len(stack)-1] = stack.Peek().NextSibling
					}
				}
				return len(stack) > 0
			}
			iterator.nextMethod = func() any {
				htmlNode := stack.Peek()
				return NewLoxHTMLNode(htmlNode)
			}
			return NewLoxIterator(iterator), nil
		})
	case "firstChild", "fc":
		return getHTMLNode(loxHTMLNodeFirstChild)
	case "isRootNode":
		return htmlNodeField(l.isRootNode())
	case "lastChild", "lc":
		return getHTMLNode(loxHTMLNodeLastChild)
	case "nextSibling", "ns":
		return getHTMLNode(loxHTMLNodeNextSibling)
	case "nodesByType":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				if arg < int64(html.ErrorNode) || arg > int64(html.RawNode) {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Integer argument to 'HTML node.nodesByType' must be from %v to %v.",
							html.ErrorNode,
							html.RawNode,
						),
					)
				}
				nodeType := html.NodeType(arg)
				nodes := list.NewList[any]()
				l.forEachDescendent(func(n *html.Node) {
					if n.Type == nodeType {
						nodes.Add(NewLoxHTMLNode(n))
					}
				})
				return NewLoxList(nodes), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "nodesByTypeIter":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				if arg < int64(html.ErrorNode) || arg > int64(html.RawNode) {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Integer argument to 'HTML node.nodesByTypeIter' must be from %v to %v.",
							html.ErrorNode,
							html.RawNode,
						),
					)
				}
				nodeType := html.NodeType(arg)
				stack := list.NewList[*html.Node]()
				stack.Add(l.current)
				firstIteration := true
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if !firstIteration && stack.Peek().Type == nodeType {
						if nodeType == html.ElementNode {
							stack.Add(stack.Peek().FirstChild)
						} else {
							stack[len(stack)-1] = stack.Peek().NextSibling
						}
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					for len(stack) > 0 && stack.Peek().Type != nodeType {
						if !firstIteration {
							stack.Add(stack.Peek().FirstChild)
						} else {
							firstIteration = false
						}
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					if firstIteration {
						firstIteration = false
					}
					return len(stack) > 0
				}
				iterator.nextMethod = func() any {
					htmlNode := stack.Peek()
					return NewLoxHTMLNode(htmlNode)
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "nodesByTypeStr":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				nodeType := loxHTMLNodeTypeStr(str).nodeType()
				if nodeType < 0 {
					return nil, loxerror.RuntimeError(name,
						fmt.Sprintf("HTML node.nodesByTypeStr: invalid type '%v'.", str))
				}
				nodes := list.NewList[any]()
				l.forEachDescendent(func(n *html.Node) {
					if n.Type == html.NodeType(nodeType) {
						nodes.Add(NewLoxHTMLNode(n))
					}
				})
				return NewLoxList(nodes), nil
			}
			return argMustBeType("string")
		})
	case "nodesByTypeStrIter":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				nodeTypeInt8 := loxHTMLNodeTypeStr(str).nodeType()
				if nodeTypeInt8 < 0 {
					return nil, loxerror.RuntimeError(name,
						fmt.Sprintf("HTML node.nodesByTypeStrIter: invalid type '%v'.", str))
				}
				nodeType := html.NodeType(nodeTypeInt8)
				stack := list.NewList[*html.Node]()
				stack.Add(l.current)
				firstIteration := true
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if !firstIteration && stack.Peek().Type == nodeType {
						if nodeType == html.ElementNode {
							stack.Add(stack.Peek().FirstChild)
						} else {
							stack[len(stack)-1] = stack.Peek().NextSibling
						}
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					for len(stack) > 0 && stack.Peek().Type != nodeType {
						if !firstIteration {
							stack.Add(stack.Peek().FirstChild)
						} else {
							firstIteration = false
						}
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					if firstIteration {
						firstIteration = false
					}
					return len(stack) > 0
				}
				iterator.nextMethod = func() any {
					htmlNode := stack.Peek()
					return NewLoxHTMLNode(htmlNode)
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeType("string")
		})
	case "parent", "p":
		return getHTMLNode(loxHTMLNodeParent)
	case "prevSibling", "ps":
		return getHTMLNode(loxHTMLNodePrevSibling)
	case "render":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var builder strings.Builder
			err := html.Render(&builder, l.current)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return NewLoxStringQuote(builder.String()), nil
		})
	case "renderToBuf":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			bytesBuffer := new(bytes.Buffer)
			err := html.Render(bytesBuffer, l.current)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			bytes := bytesBuffer.Bytes()
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, b := range bytes {
				addErr := buffer.add(int64(b))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "renderToFile":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxFile, ok := args[0].(*LoxFile); ok {
				if !loxFile.isWrite() && !loxFile.isAppend() {
					return nil, loxerror.RuntimeError(name,
						"File argument to 'HTML node.renderToFile' must be in write or append mode.")
				}
				err := html.Render(loxFile.file, l.current)
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return nil, nil
			}
			return argMustBeType("file")
		})
	case "tag":
		return htmlNodeField(NewLoxString(l.current.DataAtom.String(), '\''))
	case "tagNodesByAttrKey":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := strings.ToLower(loxStr.str)
				tagNodes := list.NewList[any]()
				l.forEachDescendent(func(n *html.Node) {
					if n.Type == html.ElementNode {
						for _, attr := range n.Attr {
							if attr.Key == str {
								tagNodes.Add(NewLoxHTMLNode(n))
								break
							}
						}
					}
				})
				return NewLoxList(tagNodes), nil
			}
			return argMustBeType("string")
		})
	case "tagNodesByAttrKeyIter":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := strings.ToLower(loxStr.str)
				stack := list.NewList[*html.Node]()
				stack.Add(l.current)
				firstIteration := true
				condition := func() bool {
					e := stack.Peek()
					if e.Type != html.ElementNode {
						return false
					}
					for _, attr := range e.Attr {
						if attr.Key == str {
							return true
						}
					}
					return false
				}
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if !firstIteration && condition() {
						stack.Add(stack.Peek().FirstChild)
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					for len(stack) > 0 && !condition() {
						if !firstIteration {
							stack.Add(stack.Peek().FirstChild)
						} else {
							firstIteration = false
						}
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					if firstIteration {
						firstIteration = false
					}
					return len(stack) > 0
				}
				iterator.nextMethod = func() any {
					htmlNode := stack.Peek()
					return NewLoxHTMLNode(htmlNode)
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeType("string")
		})
	case "tagNodesByAttrKeysAll":
		return htmlNodeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			attrKeyNames := make(map[string]uint8)
			for i := 0; i < argsLen; i++ {
				if loxStr, ok := args[i].(*LoxString); ok {
					attrKeyNames[strings.ToLower(loxStr.str)] = 0
				} else {
					attrKeyNames = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument number %v in 'HTML node.tagNodesByAttrKeysAll' must be a string.",
							i+1,
						),
					)
				}
			}
			tagNodes := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				if n.Type == html.ElementNode && len(n.Attr) > 0 {
					defer func() {
						for key := range attrKeyNames {
							attrKeyNames[key] = 0
						}
					}()
					for _, attr := range n.Attr {
						if _, ok := attrKeyNames[attr.Key]; !ok {
							return
						}
						attrKeyNames[attr.Key] = 1
					}
					for _, value := range attrKeyNames {
						if value != 1 {
							return
						}
					}
					tagNodes.Add(NewLoxHTMLNode(n))
				}
			})
			return NewLoxList(tagNodes), nil
		})
	case "tagNodesByAttrKeysAny":
		return htmlNodeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			attrKeyNames := make(map[string]struct{})
			for i := 0; i < argsLen; i++ {
				if loxStr, ok := args[i].(*LoxString); ok {
					attrKeyNames[strings.ToLower(loxStr.str)] = struct{}{}
				} else {
					attrKeyNames = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument number %v in 'HTML node.tagNodesByAttrKeysAny' must be a string.",
							i+1,
						),
					)
				}
			}
			tagNodes := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				if n.Type == html.ElementNode {
					for _, attr := range n.Attr {
						if _, ok := attrKeyNames[attr.Key]; ok {
							tagNodes.Add(NewLoxHTMLNode(n))
							break
						}
					}
				}
			})
			return NewLoxList(tagNodes), nil
		})
	case "tagNodesByAttrKeysNotAll":
		return htmlNodeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			attrKeyNames := make(map[string]struct{})
			for i := 0; i < argsLen; i++ {
				if loxStr, ok := args[i].(*LoxString); ok {
					attrKeyNames[strings.ToLower(loxStr.str)] = struct{}{}
				} else {
					attrKeyNames = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument number %v in 'HTML node.tagNodesByAttrKeysNotAll' must be a string.",
							i+1,
						),
					)
				}
			}
			tagNodes := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				if n.Type == html.ElementNode {
					if len(n.Attr) == 0 {
						tagNodes.Add(NewLoxHTMLNode(n))
					} else {
						for _, attr := range n.Attr {
							if _, ok := attrKeyNames[attr.Key]; !ok {
								tagNodes.Add(NewLoxHTMLNode(n))
								break
							}
						}
					}
				}
			})
			return NewLoxList(tagNodes), nil
		})
	case "tagNodesByAttrKeysNotAny":
		return htmlNodeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			attrKeyNames := make(map[string]struct{})
			for i := 0; i < argsLen; i++ {
				if loxStr, ok := args[i].(*LoxString); ok {
					attrKeyNames[strings.ToLower(loxStr.str)] = struct{}{}
				} else {
					attrKeyNames = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument number %v in 'HTML node.tagNodesByAttrKeysNotAny' must be a string.",
							i+1,
						),
					)
				}
			}
			tagNodes := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				if n.Type == html.ElementNode {
					for _, attr := range n.Attr {
						if _, ok := attrKeyNames[attr.Key]; ok {
							return
						}
					}
					tagNodes.Add(NewLoxHTMLNode(n))
				}
			})
			return NewLoxList(tagNodes), nil
		})
	case "tagNodesByAttrKeyVal":
		return htmlNodeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'HTML node.tagNodesByAttrKeyVal' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'HTML node.tagNodesByAttrKeyVal' must be a string.")
			}
			attrKey := strings.ToLower(args[0].(*LoxString).str)
			attrVal := args[1].(*LoxString).str
			tagNodes := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				if n.Type == html.ElementNode {
					for _, attr := range n.Attr {
						if attr.Key == attrKey && attr.Val == attrVal {
							tagNodes.Add(NewLoxHTMLNode(n))
							break
						}
					}
				}
			})
			return NewLoxList(tagNodes), nil
		})
	case "tagNodesByAttrKeyValIter":
		return htmlNodeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'HTML node.tagNodesByAttrKeyValIter' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'HTML node.tagNodesByAttrKeyValIter' must be a string.")
			}
			attrKey := strings.ToLower(args[0].(*LoxString).str)
			attrVal := args[1].(*LoxString).str
			stack := list.NewList[*html.Node]()
			stack.Add(l.current)
			firstIteration := true
			condition := func() bool {
				e := stack.Peek()
				if e.Type != html.ElementNode {
					return false
				}
				for _, attr := range e.Attr {
					if attr.Key == attrKey && attr.Val == attrVal {
						return true
					}
				}
				return false
			}
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				if !firstIteration && condition() {
					stack.Add(stack.Peek().FirstChild)
					for len(stack) > 0 && stack.Peek() == nil {
						stack.Pop()
						if len(stack) > 0 && stack.Peek() != nil {
							stack[len(stack)-1] = stack.Peek().NextSibling
						}
					}
				}
				for len(stack) > 0 && !condition() {
					if !firstIteration {
						stack.Add(stack.Peek().FirstChild)
					} else {
						firstIteration = false
					}
					for len(stack) > 0 && stack.Peek() == nil {
						stack.Pop()
						if len(stack) > 0 && stack.Peek() != nil {
							stack[len(stack)-1] = stack.Peek().NextSibling
						}
					}
				}
				if firstIteration {
					firstIteration = false
				}
				return len(stack) > 0
			}
			iterator.nextMethod = func() any {
				htmlNode := stack.Peek()
				return NewLoxHTMLNode(htmlNode)
			}
			return NewLoxIterator(iterator), nil
		})
	case "tagNodesByName":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := strings.ToLower(loxStr.str)
				tagNodes := list.NewList[any]()
				if str == "!doctype" {
					l.forEachDescendent(func(n *html.Node) {
						if n.Type == html.DoctypeNode {
							tagNodes.Add(NewLoxHTMLNode(n))
						}
					})
				} else {
					l.forEachDescendent(func(n *html.Node) {
						if n.Type == html.ElementNode && n.Data == str {
							tagNodes.Add(NewLoxHTMLNode(n))
						}
					})
				}
				return NewLoxList(tagNodes), nil
			}
			return argMustBeType("string")
		})
	case "tagNodesByNameIter":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := strings.ToLower(loxStr.str)
				stack := list.NewList[*html.Node]()
				stack.Add(l.current)
				firstIteration := true
				iterator := ProtoIterator{}
				if str == "!doctype" {
					nodeType := html.DoctypeNode
					iterator.hasNextMethod = func() bool {
						if !firstIteration && stack.Peek().Type == nodeType {
							stack[len(stack)-1] = stack.Peek().NextSibling
							for len(stack) > 0 && stack.Peek() == nil {
								stack.Pop()
								if len(stack) > 0 && stack.Peek() != nil {
									stack[len(stack)-1] = stack.Peek().NextSibling
								}
							}
						}
						for len(stack) > 0 && stack.Peek().Type != nodeType {
							if !firstIteration {
								stack.Add(stack.Peek().FirstChild)
							} else {
								firstIteration = false
							}
							for len(stack) > 0 && stack.Peek() == nil {
								stack.Pop()
								if len(stack) > 0 && stack.Peek() != nil {
									stack[len(stack)-1] = stack.Peek().NextSibling
								}
							}
						}
						if firstIteration {
							firstIteration = false
						}
						return len(stack) > 0
					}
				} else {
					condition := func() bool {
						e := stack.Peek()
						return e.Type == html.ElementNode && e.Data == str
					}
					iterator.hasNextMethod = func() bool {
						if !firstIteration && condition() {
							stack.Add(stack.Peek().FirstChild)
							for len(stack) > 0 && stack.Peek() == nil {
								stack.Pop()
								if len(stack) > 0 && stack.Peek() != nil {
									stack[len(stack)-1] = stack.Peek().NextSibling
								}
							}
						}
						for len(stack) > 0 && !condition() {
							if !firstIteration {
								stack.Add(stack.Peek().FirstChild)
							} else {
								firstIteration = false
							}
							for len(stack) > 0 && stack.Peek() == nil {
								stack.Pop()
								if len(stack) > 0 && stack.Peek() != nil {
									stack[len(stack)-1] = stack.Peek().NextSibling
								}
							}
						}
						if firstIteration {
							firstIteration = false
						}
						return len(stack) > 0
					}
				}
				iterator.nextMethod = func() any {
					htmlNode := stack.Peek()
					return NewLoxHTMLNode(htmlNode)
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeType("string")
		})
	case "tagNodesNoAttrs":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			tagNodes := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				if n.Type == html.ElementNode && len(n.Attr) == 0 {
					tagNodes.Add(NewLoxHTMLNode(n))
				}
			})
			return NewLoxList(tagNodes), nil
		})
	case "textNodes":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			textNodes := list.NewList[any]()
			l.forEachDescendent(func(n *html.Node) {
				if n.Type == html.TextNode {
					textNodes.Add(NewLoxHTMLNode(n))
				}
			})
			return NewLoxList(textNodes), nil
		})
	case "textNodesByContent":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				textNodes := list.NewList[any]()
				l.forEachDescendent(func(n *html.Node) {
					if n.Type == html.TextNode && n.Data == str {
						textNodes.Add(NewLoxHTMLNode(n))
					}
				})
				return NewLoxList(textNodes), nil
			}
			return argMustBeType("string")
		})
	case "textNodesByContentIter":
		return htmlNodeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				stack := list.NewList[*html.Node]()
				stack.Add(l.current)
				firstIteration := true
				condition := func() bool {
					e := stack.Peek()
					return e.Type == html.TextNode && e.Data == str
				}
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if !firstIteration && condition() {
						stack[len(stack)-1] = stack.Peek().NextSibling
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					for len(stack) > 0 && !condition() {
						if !firstIteration {
							stack.Add(stack.Peek().FirstChild)
						} else {
							firstIteration = false
						}
						for len(stack) > 0 && stack.Peek() == nil {
							stack.Pop()
							if len(stack) > 0 && stack.Peek() != nil {
								stack[len(stack)-1] = stack.Peek().NextSibling
							}
						}
					}
					if firstIteration {
						firstIteration = false
					}
					return len(stack) > 0
				}
				iterator.nextMethod = func() any {
					htmlNode := stack.Peek()
					return NewLoxHTMLNode(htmlNode)
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeType("string")
		})
	case "textNodesIter":
		return htmlNodeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			stack := list.NewList[*html.Node]()
			stack.Add(l.current)
			firstIteration := true
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				if !firstIteration && stack.Peek().Type == html.TextNode {
					stack[len(stack)-1] = stack.Peek().NextSibling
					for len(stack) > 0 && stack.Peek() == nil {
						stack.Pop()
						if len(stack) > 0 && stack.Peek() != nil {
							stack[len(stack)-1] = stack.Peek().NextSibling
						}
					}
				}
				for len(stack) > 0 && stack.Peek().Type != html.TextNode {
					if !firstIteration {
						stack.Add(stack.Peek().FirstChild)
					} else {
						firstIteration = false
					}
					for len(stack) > 0 && stack.Peek() == nil {
						stack.Pop()
						if len(stack) > 0 && stack.Peek() != nil {
							stack[len(stack)-1] = stack.Peek().NextSibling
						}
					}
				}
				if firstIteration {
					firstIteration = false
				}
				return len(stack) > 0
			}
			iterator.nextMethod = func() any {
				htmlNode := stack.Peek()
				return NewLoxHTMLNode(htmlNode)
			}
			return NewLoxIterator(iterator), nil
		})
	case "type":
		return htmlNodeField(int64(l.current.Type))
	case "typeStr":
		typeStr := loxHTMLNodeType(l.current.Type).String()
		return htmlNodeField(NewLoxString(typeStr, '\''))
	}
	return nil, loxerror.RuntimeError(name, "HTML nodes have no property called '"+lexemeName+"'.")
}

func (l *LoxHTMLNode) String() string {
	if l.isRootNode() && l.current.Data == "" {
		return fmt.Sprintf("<HTML root node at %p>", l)
	}
	tokenTypeStr := strings.ToLower(loxHTMLNodeType(l.current.Type).String())
	switch l.current.Type {
	case html.ElementNode:
		tagName := l.current.Data
		return fmt.Sprintf("<HTML %v node \"%v\" at %p>", tokenTypeStr, tagName, l)
	}
	return fmt.Sprintf("<HTML %v node at %p>", tokenTypeStr, l)
}

func (l *LoxHTMLNode) Type() string {
	return "html node"
}
