package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxClass struct {
	name                string
	superClass          *LoxClass
	methods             map[string]*LoxFunction
	bindedStaticMethods map[string]*LoxFunction
	classProperties     map[string]any
	instanceFields      map[string]any
	canInstantiate      bool
	isBuiltin           bool
}

type LoxBuiltInProtoCallable struct {
	instance *LoxInstance
	callable *struct{ ProtoLoxCallable }
}

func (l LoxBuiltInProtoCallable) arity() int {
	return l.callable.arity()
}

func (l LoxBuiltInProtoCallable) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	return l.callable.call(interpreter, arguments)
}

func (l LoxBuiltInProtoCallable) String() string {
	return l.callable.String()
}

func (l LoxBuiltInProtoCallable) Type() string {
	return l.callable.Type()
}

func NewLoxClass(name string, superClass *LoxClass, canInstantiate bool) *LoxClass {
	return &LoxClass{
		name:            name,
		superClass:      superClass,
		methods:         make(map[string]*LoxFunction),
		classProperties: make(map[string]any),
		instanceFields:  make(map[string]any),
		canInstantiate:  canInstantiate,
		isBuiltin:       false,
	}
}

func (c *LoxClass) arity() int {
	if !c.canInstantiate {
		return -1
	}
	initializer, ok := c.findMethod("init")
	if ok {
		return initializer.arity()
	} else if c.isChildOfBuiltInClass() {
		initializer, ok := c.findInstanceField("init")
		if ok {
			switch initializer := initializer.(type) {
			case LoxCallable:
				return initializer.arity()
			}
		}
	}
	return 0
}

func (c *LoxClass) call(interpreter *Interpreter, arguments list.List[any]) (any, error) {
	if !c.canInstantiate {
		return nil, loxerror.RuntimeError(interpreter.callToken,
			fmt.Sprintf("Cannot instantiate static class '%v'.", c.name))
	}
	instance := NewLoxInstance(c)
	for cls := c; cls != nil; cls = cls.superClass {
		for name, field := range cls.instanceFields {
			if _, ok := instance.fields[name]; !ok {
				switch field := field.(type) {
				case *struct{ ProtoLoxCallable }:
					instance.fields[name] = LoxBuiltInProtoCallable{instance, field}
				default:
					instance.fields[name] = field
				}
			}
		}
	}
	initializer, ok := c.findMethod("init")
	if ok {
		call, callErr := initializer.bind(instance).call(interpreter, arguments)
		if callErr != nil && call == nil {
			return nil, callErr
		}
	} else if c.isChildOfBuiltInClass() {
		initializer, ok := c.findInstanceField("init")
		if ok {
			switch initializer := initializer.(type) {
			case LoxCallable:
				arguments.AddAt(0, instance)
				call, callErr := initializer.call(interpreter, arguments)
				if callErr != nil && call == nil {
					return nil, callErr
				}
			}
		}
	}
	return instance, nil
}

func (c *LoxClass) Get(name *token.Token) (any, error) {
	staticMethod, foundMethod, methodDepth := c.findBindedStaticMethod(name.Lexeme)
	if foundMethod && methodDepth == 0 {
		return staticMethod, nil
	}
	item, foundClassProp, classPropDepth := c.findClassProperty(name.Lexeme)
	if foundClassProp {
		switch method := item.(type) {
		case *LoxFunction:
			if classPropDepth == methodDepth && foundMethod {
				return staticMethod, nil
			}
			bindedMethod := method.bind(c)
			cls := c
			for i := 0; i < classPropDepth; i++ {
				cls = cls.superClass
			}
			cls.bindedStaticMethods[name.Lexeme] = bindedMethod
			return bindedMethod, nil
		}
		return item, nil
	}
	return nil, loxerror.RuntimeError(name, "Undefined property '"+name.Lexeme+"'.")
}

func (c *LoxClass) findBindedStaticMethod(name string) (*LoxFunction, bool, int) {
	return c.findBindedStaticMethodHelper(name, 0)
}

func (c *LoxClass) findBindedStaticMethodHelper(name string, depth int) (*LoxFunction, bool, int) {
	value, ok := c.bindedStaticMethods[name]
	if ok {
		return value, ok, depth
	}
	if c.superClass != nil {
		return c.superClass.findBindedStaticMethodHelper(name, depth+1)
	}
	return value, ok, depth
}

func (c *LoxClass) findClassProperty(name string) (any, bool, int) {
	return c.findClassPropertyHelper(name, 0)
}

func (c *LoxClass) findClassPropertyHelper(name string, depth int) (any, bool, int) {
	value, ok := c.classProperties[name]
	if ok {
		return value, ok, depth
	}
	if c.superClass != nil {
		return c.superClass.findClassPropertyHelper(name, depth+1)
	}
	return value, ok, depth
}

func (c *LoxClass) findInstanceField(name string) (any, bool) {
	value, ok := c.instanceFields[name]
	if ok {
		return value, ok
	}
	if c.superClass != nil {
		return c.superClass.findInstanceField(name)
	}
	return value, ok
}

func (c *LoxClass) findMethod(name string) (*LoxFunction, bool) {
	value, ok := c.methods[name]
	if ok {
		return value, ok
	}
	if c.superClass != nil {
		return c.superClass.findMethod(name)
	}
	return value, ok
}

func (c *LoxClass) isChildOfBuiltInClass() bool {
	for cls := c; cls != nil; cls = cls.superClass {
		if cls.isBuiltin {
			return true
		}
	}
	return false
}

func (c *LoxClass) String() string {
	return fmt.Sprintf("<class %v at %p>", c.name, c)
}

func (c *LoxClass) Type() string {
	return "class"
}
