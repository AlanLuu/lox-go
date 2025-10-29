package ast

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

const BufferElementErrMsg = "Buffer elements must be integers between 0 and 255."
const BufferNestedElementErrMsg = "Buffers do not support nested elements."

func BufferIndexMustBeWholeNum(index any) string {
	return IndexMustBeWholeNum("Buffer", index)
}

func BufferIndexOutOfRange(index int64) string {
	return fmt.Sprintf("Buffer index %v out of range.", index)
}

type LoxBuffer struct {
	*LoxList
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxBuffer(elements list.List[any]) *LoxBuffer {
	return &LoxBuffer{
		LoxList: NewLoxList(elements),
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxBuffer() *LoxBuffer {
	return NewLoxBuffer(list.NewList[any]())
}

func EmptyLoxBufferCap(cap int64) *LoxBuffer {
	return NewLoxBuffer(list.NewListCap[any](cap))
}

func EmptyLoxBufferCapDouble(cap int64) *LoxBuffer {
	return NewLoxBuffer(list.NewListCapDouble[any](cap))
}

func (l *LoxBuffer) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxBuffer:
		if l == obj {
			return true
		}
		return reflect.DeepEqual(l.elements, obj.elements)
	default:
		return false
	}
}

func (l *LoxBuffer) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	bufferFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native buffer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'buffer.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'buffer.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	unknownTypeErr := func(element any) error {
		return loxerror.RuntimeError(name,
			fmt.Sprintf("Unknown type '%v' found in buffer.", getType(element)))
	}
	memfrob := func(start int64, stop int64) error {
		for i := start; i < stop; i++ {
			switch element := l.elements[i].(type) {
			case int64:
				l.elements[i] = element ^ 42
			default: //Should never happen
				for j := start; j < i; j++ {
					l.elements[j] = l.elements[j].(int64) ^ 42
				}
				return unknownTypeErr(element)
			}
		}
		return nil
	}
	memfrobCopy := func(start int64, stop int64) (*LoxBuffer, error) {
		elementsLen := int64(len(l.elements))
		buffer := EmptyLoxBufferCap(elementsLen)
		for i := int64(0); i < elementsLen; i++ {
			if i >= start && i < stop {
				switch element := l.elements[i].(type) {
				case int64:
					addErr := buffer.add(element ^ 42)
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				default: //Should never happen
					return nil, unknownTypeErr(element)
				}
			} else {
				addErr := buffer.add(l.elements[i])
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
		}
		return buffer, nil
	}
	switch methodName {
	case "append":
		return bufferFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			rangeErr := bufferElementRangeCheck(args[0])
			if rangeErr != nil {
				return nil, loxerror.RuntimeError(name, rangeErr.Error())
			}
			l.elements.Add(args[0])
			return nil, nil
		})
	case "extend":
		return bufferFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if extendList, ok := args[0].(*LoxBuffer); ok {
				for _, element := range extendList.elements {
					l.elements.Add(element)
				}
				return nil, nil
			}
			return argMustBeType("buffer")
		})
	case "filter":
		return bufferFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				newList := list.NewListCap[any](int64(len(l.elements)))
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = int64(index)
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
				}
				return NewLoxBuffer(newList), nil
			}
			return argMustBeType("function")
		})
	case "flatten":
		return bufferFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCap[any](int64(len(l.elements)))
			for _, element := range l.elements {
				newList.Add(element)
			}
			return NewLoxBuffer(newList), nil
		})
	case "insert":
		return bufferFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if index, ok := args[0].(int64); ok {
				originalIndex := index
				if index < 0 {
					index += int64(len(l.elements))
				}
				if index < 0 || index > int64(len(l.elements)) {
					return nil, loxerror.RuntimeError(name, BufferIndexOutOfRange(originalIndex))
				}
				newElement := args[1]
				rangeErr := bufferElementRangeCheck(newElement)
				if rangeErr != nil {
					return nil, loxerror.RuntimeError(name, rangeErr.Error())
				}
				l.elements.AddAt(index, newElement)
				return nil, nil
			}
			return nil, loxerror.RuntimeError(name, BufferIndexMustBeWholeNum(args[0]))
		})
	case "map":
		return bufferFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				argList[2] = l
				newList := list.NewListCap[any](int64(len(l.elements)))
				for index, element := range l.elements {
					argList[0] = element
					argList[1] = int64(index)
					result, resultErr := callback.call(i, argList)
					if resultReturn, ok := result.(Return); ok {
						rangeErr := bufferElementRangeCheck(resultReturn.FinalValue)
						if rangeErr != nil {
							newList.Clear()
							return nil, loxerror.RuntimeError(name, rangeErr.Error())
						}
						newList.Add(resultReturn.FinalValue)
					} else if resultErr != nil {
						newList.Clear()
						return nil, resultErr
					} else {
						rangeErr := bufferElementRangeCheck(resultReturn)
						if rangeErr != nil {
							newList.Clear()
							return nil, loxerror.RuntimeError(name, rangeErr.Error())
						}
						newList.Add(result)
					}
				}
				return NewLoxBuffer(newList), nil
			}
			return argMustBeType("function")
		})
	case "memfrob":
		return bufferFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var err error
			argsLen := len(args)
			switch argsLen {
			case 0:
				err = memfrob(0, int64(len(l.elements)))
			case 1:
				if num, ok := args[0].(int64); ok {
					if num < 0 {
						return nil, loxerror.RuntimeError(name,
							"Argument to 'buffer.memfrob' cannot be negative.")
					}
					if num > int64(len(l.elements)) {
						return nil, loxerror.RuntimeError(name,
							"Argument to 'buffer.memfrob' cannot be larger than the buffer size.")
					}
					err = memfrob(0, num)
				} else {
					return argMustBeTypeAn("integer")
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			return nil, err
		})
	case "memfrobCopy":
		return bufferFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var buffer *LoxBuffer
			var err error
			argsLen := len(args)
			switch argsLen {
			case 0:
				buffer, err = memfrobCopy(0, int64(len(l.elements)))
			case 1:
				if num, ok := args[0].(int64); ok {
					if num < 0 {
						return nil, loxerror.RuntimeError(name,
							"Argument to 'buffer.memfrobCopy' cannot be negative.")
					}
					if num > int64(len(l.elements)) {
						return nil, loxerror.RuntimeError(name,
							"Argument to 'buffer.memfrobCopy' cannot be larger than the buffer size.")
					}
					buffer, err = memfrobCopy(0, num)
				} else {
					return argMustBeTypeAn("integer")
				}
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
			}
			return buffer, err
		})
	case "memfrobRange":
		return bufferFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var start, stop int64
			elementsLen := int64(len(l.elements))
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'buffer.memfrobRange' must be an integer.")
				}
				if argsLen == 2 {
					if _, ok := args[1].(int64); !ok {
						return nil, loxerror.RuntimeError(name,
							"Second argument to 'buffer.memfrobRange' must be an integer.")
					}
					stop = args[1].(int64)
				} else {
					stop = elementsLen
				}
				start = args[0].(int64)
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
			if stop < start {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'buffer.memfrobRange' cannot be less than first argument.")
			}
			if start < 0 {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'buffer.memfrobRange' cannot be negative.")
			}
			if start > elementsLen {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'buffer.memfrobRange' cannot be larger than the buffer size.")
			}
			if stop < 0 {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'buffer.memfrobRange' cannot be negative.")
			}
			if stop > elementsLen {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'buffer.memfrobRange' cannot be larger than the buffer size.")
			}
			return nil, memfrob(start, stop)
		})
	case "memfrobRangeCopy":
		return bufferFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			var start, stop int64
			elementsLen := int64(len(l.elements))
			argsLen := len(args)
			switch argsLen {
			case 1, 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'buffer.memfrobRangeCopy' must be an integer.")
				}
				if argsLen == 2 {
					if _, ok := args[1].(int64); !ok {
						return nil, loxerror.RuntimeError(name,
							"Second argument to 'buffer.memfrobRangeCopy' must be an integer.")
					}
					stop = args[1].(int64)
				} else {
					stop = elementsLen
				}
				start = args[0].(int64)
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
			if stop < start {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'buffer.memfrobRangeCopy' cannot be less than first argument.")
			}
			if start < 0 {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'buffer.memfrobRangeCopy' cannot be negative.")
			}
			if start > elementsLen {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'buffer.memfrobRangeCopy' cannot be larger than the buffer size.")
			}
			if stop < 0 {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'buffer.memfrobRangeCopy' cannot be negative.")
			}
			if stop > elementsLen {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'buffer.memfrobRangeCopy' cannot be larger than the buffer size.")
			}
			return memfrobCopy(start, stop)
		})
	case "tobigint":
		return bufferFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			elementsLen := len(l.elements)
			if elementsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Cannot convert empty buffer to bigint.")
			}
			bigInt := big.NewInt(0)
			j := big.NewInt(0)
			twoFiveSix := big.NewInt(256)
			for i := elementsLen - 1; i >= 0; i-- {
				num := big.NewInt(l.elements[i].(int64))
				twoFiveSix.Exp(twoFiveSix, j, nil)
				twoFiveSix.Mul(twoFiveSix, num)
				bigInt.Add(bigInt, twoFiveSix)
				twoFiveSix.SetInt64(256)
				j.Add(j, bigint.One)
			}
			return bigInt, nil
		})
	case "toList":
		return bufferFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCap[any](int64(len(l.elements)))
			for _, element := range l.elements {
				newList.Add(element)
			}
			return NewLoxList(newList), nil
		})
	case "toString":
		return bufferFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var builder strings.Builder
			var b [4]byte
			bIndex := 0
			bToStr := func() string {
				var builder strings.Builder
				builder.WriteByte('[')
				for i := 0; i < bIndex; i++ {
					builder.WriteString(fmt.Sprint(b[i]))
					if i < bIndex-1 {
						builder.WriteString(", ")
					}
				}
				builder.WriteByte(']')
				return builder.String()
			}
			convertErrMsg := "Failed to convert buffer elements '%v' to string."
			useDoubleQuote := false
			for _, element := range l.elements {
				b[bIndex] = byte(element.(int64))
				if r, _ := utf8.DecodeRune(b[:bIndex+1]); r != utf8.RuneError {
					bIndex = 0
					if !useDoubleQuote && r == '\'' {
						useDoubleQuote = true
					}
					builder.WriteRune(r)
				} else {
					if bIndex == len(b)-1 {
						return nil, loxerror.RuntimeError(name, fmt.Sprintf(convertErrMsg, bToStr()))
					}
					bIndex++
				}
			}
			if bIndex != 0 {
				return nil, loxerror.RuntimeError(name, fmt.Sprintf(convertErrMsg, bToStr()))
			}
			if useDoubleQuote {
				return NewLoxString(builder.String(), '"'), nil
			}
			return NewLoxString(builder.String(), '\''), nil
		})
	case "with":
		return bufferFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if newIndex, ok := args[0].(int64); ok {
				originalNewIndex := newIndex
				if newIndex < 0 {
					newIndex += int64(len(l.elements))
				}
				if newIndex < 0 || newIndex >= int64(len(l.elements)) {
					return nil, loxerror.RuntimeError(name, BufferIndexOutOfRange(originalNewIndex))
				}
				newElement := args[1]
				rangeErr := bufferElementRangeCheck(newElement)
				if rangeErr != nil {
					return nil, loxerror.RuntimeError(name, rangeErr.Error())
				}
				newList := list.NewListCap[any](int64(len(l.elements)))
				for oldIndex, oldElement := range l.elements {
					if int64(oldIndex) != newIndex {
						newList.Add(oldElement)
					} else {
						newList.Add(newElement)
					}
				}
				return NewLoxBuffer(newList), nil
			}
			return nil, loxerror.RuntimeError(name, "First argument to 'buffer.with' must be an integer.")
		})
	default:
		element, elementErr := l.LoxList.Get(name)
		if elementErr != nil {
			errMsg := elementErr.Error()
			errMsg = strings.ReplaceAll(errMsg, "list", "buffer")
			errMsg = strings.ReplaceAll(errMsg, "Lists", "Buffers")
			return nil, loxerror.Error(errMsg)
		}
		switch element := element.(type) {
		case *struct{ ProtoLoxCallable }:
			originalCallMethod := element.callMethod
			element.callMethod = func(i *Interpreter, args list.List[any]) (any, error) {
				result, err := originalCallMethod(i, args)
				if err != nil {
					errMsg := err.Error()
					errMsg = strings.ReplaceAll(errMsg, "list", "buffer")
					err = loxerror.Error(errMsg)
				}
				return result, err
			}
			element.stringMethod = func() string {
				return fmt.Sprintf("<native buffer fn %v at %p>", methodName, element)
			}
			if _, ok := l.methods[methodName]; !ok {
				l.methods[methodName] = element
			}
		}
		return element, nil
	}
}

func bufferElementRangeCheck(element any) error {
	switch element := element.(type) {
	case int64:
		if element < 0 || element > 255 {
			return loxerror.Error(BufferElementErrMsg)
		}
		return nil
	default:
		return loxerror.Error(BufferElementErrMsg)
	}
}

func (l *LoxBuffer) add(element any) error {
	rangeCheckErr := bufferElementRangeCheck(element)
	if rangeCheckErr != nil {
		return rangeCheckErr
	}
	l.elements.Add(element)
	return nil
}

func (l *LoxBuffer) setIndex(index int64, element any) error {
	if index < 0 || index > int64(len(l.elements)) {
		return loxerror.Error(BufferIndexOutOfRange(index))
	}
	rangeCheckErr := bufferElementRangeCheck(element)
	if rangeCheckErr != nil {
		return rangeCheckErr
	}
	l.elements[index] = element
	return nil
}

func (l *LoxBuffer) String() string {
	return getResult(l, l, true)
}

func (l *LoxBuffer) Type() string {
	return "buffer"
}
