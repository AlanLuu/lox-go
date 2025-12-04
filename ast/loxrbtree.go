//https://pkg.go.dev/github.com/emirpasic/gods@v1.18.1/trees/redblacktree

package ast

import (
	"fmt"
	"os"
	"reflect"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

func loxRBTree_toLoxType(a any) any {
	switch a := a.(type) {
	case int:
		return int64(a)
	case string:
		return NewLoxStringQuote(a)
	default:
		fmt.Fprintf(
			os.Stderr,
			"rbtree warning: retrieving unknown key type '%T'\n",
			a,
		)
		return a
	}
}

type LoxRBTreeIterator struct {
	iter      *rbt.Iterator
	reversed  bool
	firstIter bool
}

func (l *LoxRBTreeIterator) HasNext() bool {
	if l.firstIter {
		l.firstIter = false
		if l.reversed {
			return l.iter.Last()
		}
		return l.iter.First()
	}
	if l.reversed {
		return l.iter.Prev()
	}
	return l.iter.Next()
}

func (l *LoxRBTreeIterator) Next() any {
	pair := list.NewListCap[any](2)
	pair.Add(loxRBTree_toLoxType(l.iter.Key()))
	pair.Add(l.iter.Value())
	return NewLoxList(pair)
}

type LoxRBTree struct {
	tree    *rbt.Tree
	keyType string
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxRBTree(tree *rbt.Tree, keyType string) *LoxRBTree {
	return &LoxRBTree{
		tree:    tree,
		keyType: keyType,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxRBTreeArgs(args ...any) *LoxRBTree {
	loxRBTree := NewLoxRBTreeIntKeys()
	for i, arg := range args {
		loxRBTree.put(i, arg)
	}
	return loxRBTree
}

func NewLoxRBTreeMapInt[T int | int64](m map[T]any) *LoxRBTree {
	loxRBTree := NewLoxRBTreeIntKeys()
	for k, v := range m {
		loxRBTree.put(k, v)
	}
	return loxRBTree
}

func NewLoxRBTreeMapStr(m map[string]any) *LoxRBTree {
	loxRBTree := NewLoxRBTreeStringKeys()
	for k, v := range m {
		loxRBTree.put(k, v)
	}
	return loxRBTree
}

func NewLoxRBTreeIntKeys() *LoxRBTree {
	return NewLoxRBTree(rbt.NewWithIntComparator(), "integer")
}

func NewLoxRBTreeStringKeys() *LoxRBTree {
	return NewLoxRBTree(rbt.NewWithStringComparator(), "string")
}

func (l *LoxRBTree) keyTypeNum() int {
	switch l.keyType {
	case "integer":
		return 0
	case "string":
		return 1
	default:
		return -1
	}
}

func (l *LoxRBTree) delete(key any) (any, bool) {
	if ok := l.setToProperKey(&key); !ok {
		return nil, false
	}
	value, ok := l.tree.Get(key)
	if !ok {
		return nil, false
	}
	l.tree.Remove(key)
	return value, true
}

func (l *LoxRBTree) deleteValues(keys ...any) []struct {
	value any
	ok    bool
} {
	values := make([]struct {
		value any
		ok    bool
	}, 0, len(keys))
	for _, key := range keys {
		value, ok := l.delete(key)
		values = append(values, struct {
			value any
			ok    bool
		}{value, ok})
	}
	return values
}

func (l *LoxRBTree) deleteValuesForce(keys ...any) {
	for _, key := range keys {
		l.deleteForce(key)
	}
}

func (l *LoxRBTree) deleteForce(key any) {
	if ok := l.setToProperKey(&key); !ok {
		return
	}
	l.tree.Remove(key)
}

func (l *LoxRBTree) equals(other *LoxRBTree, strict bool) bool {
	if l == other {
		return true
	}
	i1 := l.tree.Iterator()
	i2 := other.tree.Iterator()
	i1First := i1.First()
	i2First := i2.First()
	if i1First != i2First {
		return false
	}
	if !i1First && !i2First {
		if strict && l.keyType != other.keyType {
			return false
		}
		return true
	}
	firstIter := true
	for {
		if !firstIter {
			i1Next := i1.Next()
			i2Next := i2.Next()
			if i1Next != i2Next {
				return false
			}
			if !i1Next && !i2Next {
				if strict && l.keyType != other.keyType {
					return false
				}
				return true
			}
		}
		var result bool
		i1Key := i1.Key()
		i2Key := i2.Key()
		if i1Key == i2Key {
			result = true
		} else if first, ok := i1Key.(interfaces.Equatable); ok {
			result = first.Equals(i2Key)
		} else if second, ok := i2Key.(interfaces.Equatable); ok {
			result = second.Equals(i1Key)
		} else {
			result = reflect.DeepEqual(i1Key, i2Key)
		}
		i1Value := i1.Value()
		i2Value := i2.Value()
		if i1Value == i2Value {
			result = result && true
		} else if first, ok := i1Value.(interfaces.Equatable); ok {
			result = result && first.Equals(i2Value)
		} else if second, ok := i2Value.(interfaces.Equatable); ok {
			result = result && second.Equals(i1Value)
		} else {
			result = result && reflect.DeepEqual(i1Value, i2Value)
		}
		if !result {
			return false
		}
		firstIter = false
	}
}

func (l *LoxRBTree) forEach(f func(key, value any, index int64)) {
	l.forEachErr(func(key, value any, index int64) error {
		f(key, value, index)
		return nil
	})
}

func (l *LoxRBTree) forEachErr(f func(key, value any, index int64) error) error {
	it := l.tree.Iterator()
	if !it.First() {
		return nil
	}
	firstIter := true
	for i := int64(0); firstIter || it.Next(); i++ {
		firstIter = false
		if err := f(loxRBTree_toLoxType(it.Key()), it.Value(), i); err != nil {
			return err
		}
	}
	return nil
}

func (l *LoxRBTree) forEachPrev(f func(key, value any, index int64)) {
	l.forEachPrevErr(func(key, value any, index int64) error {
		f(key, value, index)
		return nil
	})
}

func (l *LoxRBTree) forEachPrevI0(f func(key, value any, index int64)) {
	l.forEachPrevErrI0(func(key, value any, index int64) error {
		f(key, value, index)
		return nil
	})
}

func (l *LoxRBTree) forEachPrevErr(f func(key, value any, index int64) error) error {
	it := l.tree.Iterator()
	if !it.Last() {
		return nil
	}
	firstIter := true
	for i := l.Length() - 1; firstIter || it.Prev(); i-- {
		firstIter = false
		if err := f(loxRBTree_toLoxType(it.Key()), it.Value(), i); err != nil {
			return err
		}
	}
	return nil
}

func (l *LoxRBTree) forEachPrevErrI0(f func(key, value any, index int64) error) error {
	treeLen := l.Length()
	return l.forEachPrevErr(func(key, value any, index int64) error {
		return f(key, value, treeLen-index-1)
	})
}

func (l *LoxRBTree) getValue(key any) (any, bool) {
	if ok := l.setToProperKey(&key); !ok {
		return nil, false
	}
	return l.tree.Get(key)
}

func (l *LoxRBTree) getValues(keys ...any) []struct {
	value any
	ok    bool
} {
	values := make([]struct {
		value any
		ok    bool
	}, 0, len(keys))
	for _, key := range keys {
		value, ok := l.getValue(key)
		values = append(values, struct {
			value any
			ok    bool
		}{value, ok})
	}
	return values
}

func (l *LoxRBTree) keys() list.List[any] {
	keys := l.tree.Keys()
	keysList := list.NewListCap[any](int64(cap(keys)))
	for _, key := range keys {
		switch key := key.(type) {
		case int:
			keysList.Add(int64(key))
		case string:
			keysList.Add(NewLoxStringQuote(key))
		default:
			keysList.Add(key)
		}
	}
	return keysList
}

func (l *LoxRBTree) put(key any, value any) {
	if ok := l.setToProperKey(&key); !ok {
		return
	}
	l.tree.Put(key, value)
}

func (l *LoxRBTree) setToProperKey(key *any) bool {
	switch key2 := (*key).(type) {
	case int:
	case int64:
		key2Int := int(key2)
		*key = key2Int
		if int64(key2Int) != key2 {
			fmt.Fprintf(
				os.Stderr,
				"rbtree warning: int key truncated from %v to %v\n",
				key2,
				*key,
			)
		}
	case string:
	case *LoxString:
		*key = key2.str
	default:
		fmt.Fprintf(
			os.Stderr,
			"rbtree warning: storing invalid key type '%T'\n",
			key2,
		)
		return false
	}
	return true
}

func (l *LoxRBTree) values() list.List[any] {
	return l.tree.Values()
}

func (l *LoxRBTree) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	rbTreeFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native rbtree fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'rbtree.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "clear":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.tree.Clear()
			return nil, nil
		})
	case "clearRet":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.tree.Clear()
			return l, nil
		})
	case "delete":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			value, ok := l.delete(args[0])
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"rbtree.delete: failed to delete value by key.")
			}
			return value, nil
		})
	case "deleteForce":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.deleteForce(args[0])
			return nil, nil
		})
	case "deleteForceRet":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.deleteForce(args[0])
			return l, nil
		})
	case "deleteOk":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			value, ok := l.delete(args[0])
			elements := list.NewListCap[any](2)
			elements.Add(value)
			elements.Add(ok)
			return NewLoxList(elements), nil
		})
	case "deleteOkOnly":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			_, ok := l.delete(args[0])
			return ok, nil
		})
	case "deleteValues":
		return rbTreeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"rbtree.deleteValues: expected at least 1 argument but got 0.")
			}
			deletedValueStructs := l.deleteValues(args...)
			pairs := list.NewListCap[any](int64(len(deletedValueStructs)))
			for _, v := range deletedValueStructs {
				pair := list.NewListCap[any](2)
				pair.Add(v.value)
				pair.Add(v.ok)
				pairs.Add(NewLoxList(pair))
			}
			return NewLoxList(pairs), nil
		})
	case "deleteValuesForce":
		return rbTreeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"rbtree.deleteValuesForce: expected at least 1 argument but got 0.")
			}
			l.deleteValuesForce(args...)
			return nil, nil
		})
	case "deleteValuesForceRet":
		return rbTreeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"rbtree.deleteValuesForceRet: expected at least 1 argument but got 0.")
			}
			l.deleteValuesForce(args...)
			return l, nil
		})
	case "equals":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if loxRBTree, ok := args[0].(*LoxRBTree); ok {
				return l.equals(loxRBTree, false), nil
			}
			return argMustBeType("rb tree")
		})
	case "forEach":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				return nil, l.forEachErr(func(key, value any, index int64) error {
					argList[0] = key
					argList[1] = value
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return resultErr
					}
					return nil
				})
			}
			return argMustBeType("function")
		})
	case "forEachPrev":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				return nil, l.forEachPrevErr(func(key, value any, index int64) error {
					argList[0] = key
					argList[1] = value
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return resultErr
					}
					return nil
				})
			}
			return argMustBeType("function")
		})
	case "forEachPrevI0":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				return nil, l.forEachPrevErrI0(func(key, value any, index int64) error {
					argList[0] = key
					argList[1] = value
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return resultErr
					}
					return nil
				})
			}
			return argMustBeType("function")
		})
	case "get":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			value, ok := l.getValue(args[0])
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"rbtree.get: failed to get value by key.")
			}
			return value, nil
		})
	case "getOk":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			value, ok := l.getValue(args[0])
			elements := list.NewListCap[any](2)
			elements.Add(value)
			elements.Add(ok)
			return NewLoxList(elements), nil
		})
	case "getOkOnly", "keyExists":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			_, ok := l.getValue(args[0])
			return ok, nil
		})
	case "getOrNil":
		return rbTreeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			value, _ := l.getValue(args[0])
			return value, nil
		})
	case "getValues":
		return rbTreeFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"rbtree.getValues: expected at least 1 argument but got 0.")
			}
			valuesStructs := l.getValues(args...)
			pairs := list.NewListCap[any](int64(len(valuesStructs)))
			for _, v := range valuesStructs {
				pair := list.NewListCap[any](2)
				pair.Add(v.value)
				pair.Add(v.ok)
				pairs.Add(NewLoxList(pair))
			}
			return NewLoxList(pairs), nil
		})
	case "isEmpty":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.tree.Empty(), nil
		})
	case "keys":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxList(l.keys()), nil
		})
	case "keyType":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.keyType), nil
		})
	case "mapList":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				newList := list.NewListCap[any](l.Length())
				err := l.forEachErr(func(key, value any, index int64) error {
					argList[0] = key
					argList[1] = value
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						newList.Add(resultReturn.FinalValue)
					} else if resultErr != nil {
						newList.Clear()
						return resultErr
					} else {
						newList.Add(result)
					}
					return nil
				})
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxList(newList), nil
			}
			return argMustBeType("function")
		})
	case "mapListPrev":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				newList := list.NewListCap[any](l.Length())
				err := l.forEachPrevErr(func(key, value any, index int64) error {
					argList[0] = key
					argList[1] = value
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						newList.Add(resultReturn.FinalValue)
					} else if resultErr != nil {
						newList.Clear()
						return resultErr
					} else {
						newList.Add(result)
					}
					return nil
				})
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxList(newList), nil
			}
			return argMustBeType("function")
		})
	case "mapListPrevI0":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				newList := list.NewListCap[any](l.Length())
				err := l.forEachPrevErrI0(func(key, value any, index int64) error {
					argList[0] = key
					argList[1] = value
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						newList.Add(resultReturn.FinalValue)
					} else if resultErr != nil {
						newList.Clear()
						return resultErr
					} else {
						newList.Add(result)
					}
					return nil
				})
				if err != nil {
					return nil, loxerror.RuntimeError(name, err.Error())
				}
				return NewLoxList(newList), nil
			}
			return argMustBeType("function")
		})
	case "print":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			fmt.Println(l.tree)
			return nil, nil
		})
	case "put", "set":
		return rbTreeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch l.keyTypeNum() {
			case 0:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"First argument to rbtree.%v must be an integer.",
							methodName,
						),
					)
				}
			case 1:
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"First argument to rbtree.%v must be a string.",
							methodName,
						),
					)
				}
			}
			l.put(args[0], args[1])
			return nil, nil
		})
	case "putRet", "setRet":
		return rbTreeFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch l.keyTypeNum() {
			case 0:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"First argument to rbtree.%v must be an integer.",
							methodName,
						),
					)
				}
			case 1:
				if _, ok := args[0].(*LoxString); !ok {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"First argument to rbtree.%v must be a string.",
							methodName,
						),
					)
				}
			}
			l.put(args[0], args[1])
			return l, nil
		})
	case "str":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.tree.String()), nil
		})
	case "strictEquals":
		return rbTreeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if loxRBTree, ok := args[0].(*LoxRBTree); ok {
				return l.equals(loxRBTree, true), nil
			}
			return argMustBeType("rb tree")
		})
	case "toDict":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			l.forEach(func(key, value any, _ int64) {
				dict.setKeyValue(key, value)
			})
			return dict, nil
		})
	case "toDictWithIndex":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			l.forEach(func(key, value any, index int64) {
				pair := list.NewListCap[any](2)
				pair.Add(value)
				pair.Add(index)
				dict.setKeyValue(key, NewLoxList(pair))
			})
			return dict, nil
		})
	case "toList":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[any](l.Length())
			l.forEach(func(key, value any, _ int64) {
				pair := list.NewListCap[any](2)
				pair.Add(key)
				pair.Add(value)
				pairs.Add(NewLoxList(pair))
			})
			return NewLoxList(pairs), nil
		})
	case "toListPrev":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[any](l.Length())
			l.forEachPrev(func(key, value any, _ int64) {
				pair := list.NewListCap[any](2)
				pair.Add(key)
				pair.Add(value)
				pairs.Add(NewLoxList(pair))
			})
			return NewLoxList(pairs), nil
		})
	case "toListWithIndex":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[any](l.Length())
			l.forEach(func(key, value any, index int64) {
				pair := list.NewListCap[any](3)
				pair.Add(key)
				pair.Add(value)
				pair.Add(index)
				pairs.Add(NewLoxList(pair))
			})
			return NewLoxList(pairs), nil
		})
	case "toListWithIndexPrev":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[any](l.Length())
			l.forEachPrev(func(key, value any, index int64) {
				pair := list.NewListCap[any](3)
				pair.Add(key)
				pair.Add(value)
				pair.Add(index)
				pairs.Add(NewLoxList(pair))
			})
			return NewLoxList(pairs), nil
		})
	case "toListWithIndexPrevI0":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[any](l.Length())
			l.forEachPrevI0(func(key, value any, index int64) {
				pair := list.NewListCap[any](3)
				pair.Add(key)
				pair.Add(value)
				pair.Add(index)
				pairs.Add(NewLoxList(pair))
			})
			return NewLoxList(pairs), nil
		})
	case "values":
		return rbTreeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxList(l.values()), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "RB trees have no property called '"+methodName+"'.")
}

func (l *LoxRBTree) Iterator() interfaces.Iterator {
	iter := l.tree.Iterator()
	return &LoxRBTreeIterator{
		iter:      &iter,
		reversed:  false,
		firstIter: true,
	}
}

func (l *LoxRBTree) Length() int64 {
	return int64(l.tree.Size())
}

func (l *LoxRBTree) ReverseIterator() interfaces.Iterator {
	iter := l.tree.Iterator()
	return &LoxRBTreeIterator{
		iter:      &iter,
		reversed:  true,
		firstIter: true,
	}
}

func (l *LoxRBTree) String() string {
	if l.keyType == "" {
		return fmt.Sprintf("<rb tree at %p>", l)
	}
	return fmt.Sprintf("<rb tree with %v keys at %p>", l.keyType, l)
}

func (l *LoxRBTree) Type() string {
	return "rb tree"
}
