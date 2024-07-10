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
	getArgList := func(callback *LoxFunction, numArgs int) list.List[any] {
		argList := list.NewListLen[any](int64(numArgs))
		callbackArity := callback.arity()
		if callbackArity > numArgs {
			for i := 0; i < callbackArity-numArgs; i++ {
				argList.Add(nil)
			}
		}
		return argList
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
	case "index":
		return rangeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if value, ok := args[0].(int64); ok {
				return l.index(value), nil
			}
			return argMustBeTypeAn("integer")
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
			buffer := EmptyLoxBuffer()
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
			nums := list.NewList[any]()
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
