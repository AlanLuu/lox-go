package ast

import (
	"fmt"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func RangeIndexMustBeWholeNum(index any) string {
	return IndexMustBeWholeNum("Range", index)
}

func RangeIndexOutOfRange(index int64) string {
	return fmt.Sprintf("Range index %v out of range.", index)
}

type LoxRange struct {
	start   int64
	stop    int64
	step    int64
	methods map[string]*struct{ ProtoLoxCallable }
}

type LoxRangeIterator struct {
	theRange *LoxRange
	current  int64
}

func (l *LoxRangeIterator) HasNext() bool {
	if l.theRange.step == 0 {
		return false
	}
	if l.theRange.step < 0 {
		return l.current > l.theRange.stop
	}
	return l.current < l.theRange.stop
}

func (l *LoxRangeIterator) Next() any {
	current := l.current
	l.current += l.theRange.step
	return current
}

func NewLoxRange(start int64, stop int64, step int64) *LoxRange {
	return &LoxRange{
		start:   start,
		stop:    stop,
		step:    step,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxRangeStop(stop int64) *LoxRange {
	return NewLoxRange(0, stop, 1)
}

func NewLoxRangeStartStop(start int64, stop int64) *LoxRange {
	return NewLoxRange(start, stop, 1)
}

func (l *LoxRange) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxRange:
		return l.start == obj.start &&
			l.stop == obj.stop &&
			l.step == obj.step
	default:
		return false
	}
}

func (l *LoxRange) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	rangeFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native range fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'range.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'range.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "all":
		return rangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				var index int64 = 0
				it := l.Iterator()
				for it.HasNext() {
					argList[0] = it.Next()
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if !i.isTruthy(result) {
						return false, nil
					}
					index++
				}
				return true, nil
			}
			return argMustBeType("function")
		})
	case "any":
		return rangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				var index int64 = 0
				it := l.Iterator()
				for it.HasNext() {
					argList[0] = it.Next()
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						return nil, resultErr
					}
					if i.isTruthy(result) {
						return true, nil
					}
					index++
				}
				return false, nil
			}
			return argMustBeType("function")
		})
	case "contains":
		return rangeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if value, ok := args[0].(int64); ok {
				return l.contains(value), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "filter":
		return rangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				newList := list.NewListCap[any](l.Length())
				var index int64 = 0
				it := l.Iterator()
				for it.HasNext() {
					element := it.Next()
					argList[0] = element
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						result = resultReturn.FinalValue
					} else if resultErr != nil {
						newList.Clear()
						return nil, resultErr
					}
					if i.isTruthy(result) {
						newList.Add(element)
					}
					index++
				}
				return NewLoxList(newList), nil
			}
			return argMustBeType("function")
		})
	case "forEach":
		return rangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				var index int64 = 0
				it := l.Iterator()
				for it.HasNext() {
					argList[0] = it.Next()
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return nil, resultErr
					}
					index++
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "index":
		return rangeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if value, ok := args[0].(int64); ok {
				return l.index(value), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "map":
		return rangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				newList := list.NewListCap[any](l.Length())
				var index int64 = 0
				it := l.Iterator()
				for it.HasNext() {
					argList[0] = it.Next()
					argList[1] = index
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						newList.Add(resultReturn.FinalValue)
					} else if resultErr != nil {
						newList.Clear()
						return nil, resultErr
					} else {
						newList.Add(result)
					}
					index++
				}
				return NewLoxList(newList), nil
			}
			return argMustBeType("function")
		})
	case "reduce":
		return rangeFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 || argsLen > 2 {
				return nil, loxerror.RuntimeError(name, fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
			if callback, ok := args[0].(*LoxFunction); ok {
				it := l.Iterator()
				var index int64 = 0
				var value any
				switch argsLen {
				case 1:
					if !it.HasNext() {
						return nil, loxerror.RuntimeError(name, "Cannot call 'range.reduce' on empty range without initial value.")
					}
					value = it.Next()
					index++
				case 2:
					value = args[1]
				}

				argList := getArgList(callback, 4)
				defer argList.Clear()
				argList[3] = l
				for it.HasNext() {
					argList[0] = value
					argList[1] = it.Next()
					argList[2] = index

					var valueErr error
					value, valueErr = callback.call(i, argList)
					if valueReturn, ok := value.(Return); ok {
						value = valueReturn.FinalValue
					} else if valueErr != nil {
						return nil, valueErr
					}
					index++
				}
				return value, nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'range.reduce' must be a function.")
		})
	case "start":
		return l.start, nil
	case "step":
		return l.step, nil
	case "stop":
		return l.stop, nil
	case "sum":
		return rangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var sum int64 = 0
			it := l.Iterator()
			for it.HasNext() {
				sum += it.Next().(int64)
			}
			return sum, nil
		})
	case "toBuffer":
		return rangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			capacity := l.Length()
			if capacity > 256 {
				capacity = 256
			}
			buffer := EmptyLoxBufferCap(capacity)
			it := l.Iterator()
			for it.HasNext() {
				addErr := buffer.add(it.Next())
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "toList":
		return rangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			nums := list.NewListCap[any](l.Length())
			it := l.Iterator()
			for it.HasNext() {
				nums.Add(it.Next())
			}
			return NewLoxList(nums), nil
		})
	case "toSet":
		return rangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newSet := EmptyLoxSet()
			it := l.Iterator()
			for it.HasNext() {
				_, errStr := newSet.add(it.Next())
				if len(errStr) > 0 {
					return nil, loxerror.RuntimeError(name, errStr)
				}
			}
			return newSet, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Ranges have no property called '"+methodName+"'.")
}

func (l *LoxRange) contains(value int64) bool {
	if l.step == 0 {
		return false
	}
	if l.step < 0 {
		return value <= l.start &&
			value > l.stop &&
			(value-l.start)%l.step == 0
	}
	return value >= l.start &&
		value < l.stop &&
		(value-l.start)%l.step == 0
}

func (l *LoxRange) get(index int64) int64 {
	return l.start + (index * l.step)
}

func (l *LoxRange) getRange(start int64, stop int64) *LoxRange {
	newStart := l.start + start*l.step
	if newStart > l.stop {
		newStart = l.stop
	}
	newStop := l.start + stop*l.step
	if newStop < l.start {
		newStop = l.start
	}
	return NewLoxRange(newStart, newStop, l.step)
}

func (l *LoxRange) index(value int64) int64 {
	if !l.contains(value) {
		return -1
	}
	return (value - l.start) / l.step
}

func (l *LoxRange) Iterator() interfaces.Iterator {
	return &LoxRangeIterator{l, l.start}
}

func (l *LoxRange) Length() int64 {
	if l.step > 0 && l.start < l.stop {
		return ((l.stop - l.start - 1) / l.step) + 1
	} else if l.step < 0 && l.stop < l.start {
		return ((l.start - l.stop - 1) / -l.step) + 1
	}
	return 0
}

func (l *LoxRange) String() string {
	if l.step == 1 {
		return fmt.Sprintf("range(%v, %v)", l.start, l.stop)
	}
	return fmt.Sprintf("range(%v, %v, %v)", l.start, l.stop, l.step)
}

func (l *LoxRange) Type() string {
	return "range"
}
