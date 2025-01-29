package ast

import (
	"fmt"
	"math/big"

	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func BigRangeIndexMustBeWholeNum(index any) string {
	msg := IndexMustBeWholeNum("Bigrange", index)
	return msg[:len(msg)-1] + " or bigint."
}

func BigRangeIndexOutOfRange(index *big.Int) string {
	return fmt.Sprintf("Bigrange index %v out of range.", index)
}

type LoxBigRange struct {
	start   *big.Int
	stop    *big.Int
	step    *big.Int
	methods map[string]*struct{ ProtoLoxCallable }
}

type LoxBigRangeIterator struct {
	bigRange *LoxBigRange
	current  *big.Int
}

func (l *LoxBigRangeIterator) HasNext() bool {
	if l.bigRange.step.Cmp(bigint.Zero) == 0 {
		return false
	}
	if l.bigRange.step.Cmp(bigint.Zero) < 0 {
		return l.current.Cmp(l.bigRange.stop) > 0
	}
	return l.current.Cmp(l.bigRange.stop) < 0
}

func (l *LoxBigRangeIterator) Next() any {
	current := new(big.Int).Set(l.current)
	l.current.Add(l.current, l.bigRange.step)
	return current
}

func NewLoxBigRange(start *big.Int, stop *big.Int, step *big.Int) *LoxBigRange {
	return &LoxBigRange{
		start:   start,
		stop:    stop,
		step:    step,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxBigRangeStop(stop *big.Int) *LoxBigRange {
	return NewLoxBigRange(big.NewInt(0), stop, big.NewInt(1))
}

func NewLoxBigRangeStartStop(start *big.Int, stop *big.Int) *LoxBigRange {
	return NewLoxBigRange(start, stop, big.NewInt(1))
}

func (l *LoxBigRange) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxBigRange:
		return l.start.Cmp(obj.start) == 0 &&
			l.stop.Cmp(obj.stop) == 0 &&
			l.step.Cmp(obj.step) == 0
	default:
		return false
	}
}

func (l *LoxBigRange) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	bigRangeFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native bigrange fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bigrange.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bigrange.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "all":
		return bigRangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
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
		return bigRangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
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
		return bigRangeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case int64:
				return l.contains(big.NewInt(arg)), nil
			case *big.Int:
				return l.contains(arg), nil
			}
			return argMustBeTypeAn("integer or bigint")
		})
	case "filter":
		return bigRangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
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
		return bigRangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
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
		return bigRangeFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch arg := args[0].(type) {
			case int64:
				return l.index(big.NewInt(arg)), nil
			case *big.Int:
				return l.index(arg), nil
			}
			return argMustBeTypeAn("integer or bigint")
		})
	case "map":
		return bigRangeFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
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
		return bigRangeFunc(-1, func(i *Interpreter, args list.List[any]) (any, error) {
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
						return nil, loxerror.RuntimeError(name, "Cannot call 'bigrange.reduce' on empty range without initial value.")
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
			return nil, loxerror.RuntimeError(name, "First argument to 'bigrange.reduce' must be a function.")
		})
	case "start":
		return new(big.Int).Set(l.start), nil
	case "step":
		return new(big.Int).Set(l.step), nil
	case "stop":
		return new(big.Int).Set(l.stop), nil
	case "sum":
		return bigRangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			sum := big.NewInt(0)
			it := l.Iterator()
			for it.HasNext() {
				sum.Add(sum, it.Next().(*big.Int))
			}
			return sum, nil
		})
	case "toBuffer":
		return bigRangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			capacity := l.Length()
			if capacity > 256 {
				capacity = 256
			}
			buffer := EmptyLoxBufferCap(capacity)
			it := l.Iterator()
			for it.HasNext() {
				addErr := buffer.add(it.Next().(*big.Int).Int64())
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "toList":
		return bigRangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			nums := list.NewListCap[any](l.Length())
			it := l.Iterator()
			for it.HasNext() {
				nums.Add(it.Next())
			}
			return NewLoxList(nums), nil
		})
	case "toSet":
		return bigRangeFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
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
	return nil, loxerror.RuntimeError(name, "Bigranges have no property called '"+methodName+"'.")
}

func (l *LoxBigRange) contains(value *big.Int) bool {
	if l.step.Cmp(bigint.Zero) == 0 {
		return false
	}
	if l.step.Cmp(bigint.Zero) < 0 {
		if !(value.Cmp(l.start) <= 0) || !(value.Cmp(l.stop) > 0) {
			return false
		}
	} else {
		if !(value.Cmp(l.start) >= 0) || !(value.Cmp(l.stop) < 0) {
			return false
		}
	}
	bigInt := &big.Int{}
	bigInt.Sub(value, l.start)
	bigInt.Mod(bigInt, l.step)
	return bigInt.Cmp(bigint.Zero) == 0
}

func (l *LoxBigRange) get(index *big.Int) *big.Int {
	bigInt := &big.Int{}
	bigInt.Mul(index, l.step)
	bigInt.Add(bigInt, l.start)
	return bigInt
}

func (l *LoxBigRange) getRange(start *big.Int, stop *big.Int) *LoxBigRange {
	newStart := &big.Int{}
	newStart.Mul(start, l.step)
	newStart.Add(newStart, l.start)
	if newStart.Cmp(l.stop) > 0 {
		newStart.Set(l.stop)
	}
	newStop := &big.Int{}
	newStop.Mul(stop, l.step)
	newStop.Add(newStop, l.start)
	if newStop.Cmp(l.start) < 0 {
		newStop.Set(l.start)
	}
	return NewLoxBigRange(newStart, newStop, new(big.Int).Set(l.step))
}

func (l *LoxBigRange) index(value *big.Int) *big.Int {
	if !l.contains(value) {
		return big.NewInt(-1)
	}
	bigInt := &big.Int{}
	bigInt.Sub(value, l.start)
	bigInt.Div(bigInt, l.step)
	return bigInt
}

func (l *LoxBigRange) Iterator() interfaces.Iterator {
	return &LoxBigRangeIterator{l, new(big.Int).Set(l.start)}
}

func (l *LoxBigRange) Length() int64 {
	if l.step.Cmp(bigint.Zero) > 0 && l.start.Cmp(l.stop) < 0 {
		bigInt := new(big.Int).Set(l.stop)
		bigInt.Sub(bigInt, l.start)
		bigInt.Sub(bigInt, bigint.One)
		bigInt.Div(bigInt, l.step)
		bigInt.Add(bigInt, bigint.One)
		return bigInt.Int64()
	} else if l.step.Cmp(bigint.Zero) < 0 && l.stop.Cmp(l.start) < 0 {
		bigInt := new(big.Int).Set(l.start)
		bigInt.Sub(bigInt, l.stop)
		bigInt.Sub(bigInt, bigint.One)
		bigInt.Div(bigInt, new(big.Int).Neg(l.step))
		bigInt.Add(bigInt, bigint.One)
		return bigInt.Int64()
	}
	return 0
}

func (l *LoxBigRange) String() string {
	if l.step.Cmp(bigint.One) == 0 {
		return fmt.Sprintf("bigrange(%v, %v)", l.start, l.stop)
	}
	return fmt.Sprintf("bigrange(%v, %v, %v)", l.start, l.stop, l.step)
}

func (l *LoxBigRange) Type() string {
	return "bigrange"
}
