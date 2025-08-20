package ast

import (
	"container/ring"
	"fmt"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxRingIterator struct {
	ring      *ring.Ring
	current   *ring.Ring
	firstIter bool
	reverse   bool
}

func (l *LoxRingIterator) HasNext() bool {
	hasNext := l.firstIter || l.current != l.ring
	l.firstIter = false
	return hasNext
}

func (l *LoxRingIterator) Next() any {
	value := l.current.Value
	if l.reverse {
		l.current = l.current.Prev()
	} else {
		l.current = l.current.Next()
	}
	return value
}

type LoxRing struct {
	ring    *ring.Ring
	prev    *LoxRing
	next    *LoxRing
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxRingArgs(args ...any) *LoxRing {
	argsLen := len(args)
	if argsLen == 0 {
		return NewLoxRingNils(1)
	}
	loxRings := list.NewList[*LoxRing]()
	ringObj := ring.New(argsLen)
	ringObj.Value = args[0]
	loxRings.Add(&LoxRing{
		ring:    ringObj,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	})
	i := 1
	for r := ringObj.Next(); r != ringObj; r = r.Next() {
		r.Value = args[i]
		loxRings.Add(&LoxRing{
			ring:    r,
			methods: make(map[string]*struct{ ProtoLoxCallable }),
		})
		loxRings[i].prev = loxRings[i-1]
		loxRings[i-1].next = loxRings[i]
		i++
	}
	if loxRingsLen := len(loxRings); loxRingsLen == 1 {
		loxRings[0].prev = loxRings[0]
		loxRings[0].next = loxRings[0]
	} else {
		loxRings[0].prev = loxRings[loxRingsLen-1]
		loxRings[loxRingsLen-1].next = loxRings[0]
	}
	return loxRings[0]
}

func NewLoxRingNils(n int) *LoxRing {
	if n <= 0 {
		panic("in NewLoxRingNils: n is less than 1")
	}
	loxRings := list.NewList[*LoxRing]()
	ringObj := ring.New(n)
	loxRings.Add(&LoxRing{
		ring:    ringObj,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	})
	i := 1
	for r := ringObj.Next(); r != ringObj; r = r.Next() {
		loxRings.Add(&LoxRing{
			ring:    r,
			methods: make(map[string]*struct{ ProtoLoxCallable }),
		})
		loxRings[i].prev = loxRings[i-1]
		loxRings[i-1].next = loxRings[i]
		i++
	}
	if loxRingsLen := len(loxRings); loxRingsLen == 1 {
		loxRings[0].prev = loxRings[0]
		loxRings[0].next = loxRings[0]
	} else {
		loxRings[0].prev = loxRings[loxRingsLen-1]
		loxRings[loxRingsLen-1].next = loxRings[0]
	}
	return loxRings[0]
}

func (l *LoxRing) getIndexPositive(index int64) any {
	if index < 0 {
		panic("in LoxRing.getIndexPositive: index is negative")
	}
	var current int64 = 0
	firstIter := true
	for r := l.ring; firstIter || r != l.ring; r = r.Next() {
		firstIter = false
		if current == index {
			return r.Value
		} else if current > index {
			break
		}
		current++
	}
	panic("in LoxRing.getIndexPositive: reached end of ring")
}

func (l *LoxRing) link(other_l *LoxRing) *LoxRing {
	if l == other_l {
		return l
	}
	l.ring.Link(other_l.ring)

	l_next := l.next
	other_l_prev := other_l.prev
	l.next = other_l
	other_l.prev = l
	l_next.prev = other_l_prev
	other_l_prev.next = l_next

	return l_next
}

func (l *LoxRing) move(arg int64) *LoxRing {
	r := l
	if arg > 0 {
		for ; arg > 0; arg-- {
			r = r.next
		}
	} else if arg < 0 {
		for ; arg < 0; arg++ {
			r = r.prev
		}
	}
	return r
}

func (l *LoxRing) unlink(arg int64) *LoxRing {
	if arg <= 0 {
		return l
	}
	return l.link(l.move(arg + 1))
}

func (l *LoxRing) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if field, ok := l.methods[methodName]; ok {
		return field, nil
	}
	ringFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native ring fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'ring.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'ring.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "extend":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				if arg <= 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'ring.extend' cannot be less than 1.")
				}
				l.link(NewLoxRingNils(int(arg)))
				return nil, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "extendRet":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				if arg <= 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'ring.extendRet' cannot be less than 1.")
				}
				l.link(NewLoxRingNils(int(arg)))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "filter":
		return ringFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 2)
				defer argList.Clear()
				elements := list.NewList[any]()
				var index int64 = 0
				firstIter := true
				for r := l.ring; firstIter || r != l.ring; r = r.Next() {
					firstIter = false
					argList[0] = r.Value
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						elements = nil
						return nil, resultErr
					}
					if i.isTruthy(result) {
						elements.Add(r.Value)
					}
					index++
				}
				return NewLoxRingArgs(elements...), nil
			}
			return argMustBeType("function")
		})
	case "forEach":
		return ringFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 2)
				defer argList.Clear()
				errorChan := make(chan error, 1)
				go func() {
					var index int64 = 0
					foundError := false
					l.ring.Do(func(value any) {
						if !foundError {
							argList[0] = value
							argList[1] = index
							result, resultErr := callback.call(i, argList)
							if resultErr != nil && result == nil {
								errorChan <- resultErr
								foundError = true
							} else {
								index++
							}
						}
					})
					close(errorChan)
				}()
				err, ok := <-errorChan
				if ok && err != nil {
					return nil, err
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "getValue":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				return l.move(arg).ring.Value, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "infIter":
		return ringFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			r := l.ring
			iterator := InfiniteIterator{}
			iterator.nextMethod = func() any {
				value := r.Value
				r = r.Next()
				return value
			}
			return NewLoxIterator(iterator), nil
		})
	case "isLinked":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxRing); ok {
				if other_l == l {
					return true, nil
				}
				for r := l.next; r != l; r = r.next {
					if r == other_l {
						return true, nil
					}
				}
				return false, nil
			}
			return argMustBeType("ring")
		})
	case "limitIter":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				var i int64 = -1
				r := l.ring
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					i++
					return i < num
				}
				iterator.nextMethod = func() any {
					value := r.Value
					r = r.Next()
					return value
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "limitIterNoWrap":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				var i int64 = -1
				r := l.ring
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					i++
					if i == 0 {
						return i < num
					}
					return i < num && r != l.ring
				}
				iterator.nextMethod = func() any {
					value := r.Value
					r = r.Next()
					return value
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "link":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxRing); ok {
				return l.link(other_l), nil
			}
			return argMustBeType("ring")
		})
	case "loopIter":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				var i int64 = -1
				r := l.ring
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if r == l.ring {
						i++
					}
					return i < arg
				}
				iterator.nextMethod = func() any {
					value := r.Value
					r = r.Next()
					return value
				}
				return NewLoxIterator(iterator), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "map":
		return ringFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 2)
				defer argList.Clear()
				elements := list.NewList[any]()
				var index int64 = 0
				firstIter := true
				for r := l.ring; firstIter || r != l.ring; r = r.Next() {
					firstIter = false
					argList[0] = r.Value
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						elements.Add(resultReturn.FinalValue)
					} else if resultErr != nil {
						elements = nil
						return nil, resultErr
					} else {
						elements.Add(result)
					}
					index++
				}
				return NewLoxRingArgs(elements...), nil
			}
			return argMustBeType("function")
		})
	case "mapInPlace":
		return ringFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 2)
				defer argList.Clear()
				var index int64 = 0
				firstIter := true
				for r := l.ring; firstIter || r != l.ring; r = r.Next() {
					firstIter = false
					argList[0] = r.Value
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						r.Value = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					} else {
						r.Value = result
					}
					index++
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "move":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				return l.move(arg), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "next":
		return l.next, nil
	case "prev":
		return l.prev, nil
	case "printLines":
		return ringFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.ring.Do(func(value any) {
				fmt.Println(getResult(value, value, true))
			})
			return nil, nil
		})
	case "printLinesSep":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				firstIter := true
				l.ring.Do(func(value any) {
					if firstIter {
						firstIter = false
					} else if str != "" {
						fmt.Println(str)
					}
					fmt.Println(getResult(value, value, true))
				})
				return nil, nil
			}
			return argMustBeType("string")
		})
	case "printList":
		return ringFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var builder strings.Builder
			builder.WriteString("Ring [")
			firstIter := true
			l.ring.Do(func(value any) {
				if firstIter {
					firstIter = false
				} else {
					builder.WriteString(", ")
				}
				builder.WriteString(getResult(value, value, false))
			})
			builder.WriteRune(']')
			fmt.Println(builder.String())
			return nil, nil
		})
	case "printListToStr":
		return ringFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var builder strings.Builder
			builder.WriteString("Ring [")
			firstIter := true
			l.ring.Do(func(value any) {
				if firstIter {
					firstIter = false
				} else {
					builder.WriteString(", ")
				}
				builder.WriteString(getResult(value, value, false))
			})
			builder.WriteRune(']')
			return NewLoxStringQuote(builder.String()), nil
		})
	case "reduce":
		return ringFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 || argsLen > 2 {
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
			if callback, ok := args[0].(*LoxFunction); ok {
				var value any
				switch argsLen {
				case 1:
					value = l.ring.Value
				case 2:
					value = args[1]
				}

				argList := getArgList(callback, 3)
				defer argList.Clear()
				var index int64 = 0
				firstIter := true
				for r := l.ring; firstIter || r != l.ring; r = r.Next() {
					firstIter = false
					if !(index == 0 && argsLen == 1) {
						argList[0] = value
						argList[1] = r.Value
						argList[2] = index

						var valueErr error
						value, valueErr = callback.call(i, argList)
						if valueReturn, ok := value.(Return); ok {
							value = valueReturn.FinalValue
						} else if valueErr != nil {
							return nil, valueErr
						}
					}
					index++
				}
				return value, nil
			}
			return nil, loxerror.RuntimeError(name,
				"First argument to 'ring.reduce' must be a function.")
		})
	case "reduceRight":
		return ringFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 || argsLen > 2 {
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
			if callback, ok := args[0].(*LoxFunction); ok {
				var value any
				switch argsLen {
				case 1:
					value = l.ring.Prev().Value
				case 2:
					value = args[1]
				}

				argList := getArgList(callback, 3)
				defer argList.Clear()
				var index int64 = 0
				for r := l.ring.Prev(); ; r = r.Prev() {
					if !(index == 0 && argsLen == 1) {
						argList[0] = value
						argList[1] = r.Value
						argList[2] = index

						var valueErr error
						value, valueErr = callback.call(i, argList)
						if valueReturn, ok := value.(Return); ok {
							value = valueReturn.FinalValue
						} else if valueErr != nil {
							return nil, valueErr
						}
					}
					if r == l.ring {
						break
					}
					index++
				}
				return value, nil
			}
			return nil, loxerror.RuntimeError(name,
				"First argument to 'ring.reduceRight' must be a function.")
		})
	case "setValue":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.ring.Value = args[0]
			return nil, nil
		})
	case "setValueRet":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			l.ring.Value = args[0]
			return l, nil
		})
	case "toList":
		return ringFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			elementsList := list.NewList[any]()
			firstIter := true
			for r := l.ring; firstIter || r != l.ring; r = r.Next() {
				firstIter = false
				elementsList.Add(r.Value)
			}
			return NewLoxList(elementsList), nil
		})
	case "unlink":
		return ringFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				return l.unlink(arg), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "value":
		return l.ring.Value, nil
	}
	return nil, loxerror.RuntimeError(name, "Rings have no property called '"+methodName+"'.")
}

func (l *LoxRing) Iterator() interfaces.Iterator {
	return &LoxRingIterator{
		ring:      l.ring,
		current:   l.ring,
		firstIter: true,
		reverse:   false,
	}
}

func (l *LoxRing) Length() int64 {
	return int64(l.ring.Len())
}

func (l *LoxRing) ReverseIterator() interfaces.Iterator {
	return &LoxRingIterator{
		ring:      l.ring,
		current:   l.ring,
		firstIter: true,
		reverse:   true,
	}
}

func (l *LoxRing) String() string {
	return getResult(l, l, true)
}

func (l *LoxRing) Type() string {
	return "ring"
}
