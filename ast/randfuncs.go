package ast

import (
	"fmt"
	"math/rand"
	"unicode/utf8"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxRand struct {
	rand *rand.Rand
}

func (r LoxRand) String() string {
	return "private field"
}

func (i *Interpreter) defineRandFuncs() {
	className := "Rand"
	randClass := NewLoxClass(className, nil, true)
	randClass.isBuiltin = true
	randInstanceFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<fn %v at %p>", name, &s)
		}
		randClass.instanceFields[name] = s
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Rand().%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	randElement := func(randStruct LoxRand, arg any) (any, error) {
		emptyErr := func(theType string) (any, error) {
			return nil, loxerror.Error(
				"Cannot get random element from empty " + theType + ".",
			)
		}
		switch arg := arg.(type) {
		case *LoxBuffer:
			if arg.elements.IsEmpty() {
				return emptyErr("buffer")
			}
			var randIndex int
			if randStruct.rand != nil {
				randIndex = randStruct.rand.Intn(len(arg.elements))
			} else {
				randIndex = rand.Intn(len(arg.elements))
			}
			return arg.elements[randIndex], nil
		case *LoxList:
			if arg.elements.IsEmpty() {
				return emptyErr("list")
			}
			var randIndex int
			if randStruct.rand != nil {
				randIndex = randStruct.rand.Intn(len(arg.elements))
			} else {
				randIndex = rand.Intn(len(arg.elements))
			}
			return arg.elements[randIndex], nil
		case *LoxRange:
			rangeLen := arg.Length()
			if rangeLen == 0 {
				return emptyErr("range")
			}
			var randIndex int64
			if randStruct.rand != nil {
				randIndex = randStruct.rand.Int63n(rangeLen)
			} else {
				randIndex = rand.Int63n(rangeLen)
			}
			return arg.get(randIndex), nil
		case *LoxString:
			if len(arg.str) == 0 {
				return emptyErr("string")
			}
			var randIndex int
			if randStruct.rand != nil {
				randIndex = randStruct.rand.Intn(utf8.RuneCountInString(arg.str))
			} else {
				randIndex = rand.Intn(utf8.RuneCountInString(arg.str))
			}
			return NewLoxStringQuote(string([]rune(arg.str)[randIndex])), nil
		default:
			return nil, loxerror.Error(
				fmt.Sprintf("Cannot get random element from type '%v'.", getType(arg)),
			)
		}
	}

	randStr := "randObj"
	randInstanceFunc("init", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args) - 1
		switch argsLen {
		case 0:
			instance := args[0].(*LoxInstance)
			instance.fields[randStr] = LoxRand{nil}
			return nil, nil
		case 1:
			if seed, ok := args[1].(int64); ok {
				instance := args[0].(*LoxInstance)
				instance.fields[randStr] = LoxRand{rand.New(rand.NewSource(seed))}
				return nil, nil
			}
			return argMustBeTypeAn(in.callToken, "init", "integer")
		default:
			return nil, loxerror.RuntimeError(in.callToken, fmt.Sprintf("Expected 0 or 1 arguments but got %v.", argsLen))
		}
	})

	randFieldTypeErrMsg := "'Rand().rand' field is not the correct type."
	randInstanceFunc("choice", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			element, err := randElement(randStruct, args[1])
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return element, nil
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("choices", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			if _, ok := args[2].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'Rand().choices' must be an integer.")
			}
			numChoices := args[2].(int64)
			if numChoices < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'Rand().choices' cannot be negative.")
			}
			arg := args[1]
			choices := list.NewListCap[any](numChoices)
			for i := int64(0); i < numChoices; i++ {
				element, err := randElement(randStruct, arg)
				if err != nil {
					choices.Clear()
					return nil, loxerror.RuntimeError(in.callToken, err.Error())
				}
				choices.Add(element)
			}
			return NewLoxList(choices), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("perm", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			argsLen := len(args) - 1
			switch argsLen {
			case 1:
				if num, ok := args[1].(int64); ok {
					if num <= 0 {
						return nil, loxerror.RuntimeError(in.callToken,
							"Argument to 'Rand().perm' cannot be 0 or negative.")
					}
					var randPerms []int
					if randStruct.rand != nil {
						randPerms = randStruct.rand.Perm(int(num))
					} else {
						randPerms = rand.Perm(int(num))
					}
					permsList := list.NewListCap[any](int64(len(randPerms)))
					for _, perm := range randPerms {
						permsList.Add(int64(perm))
					}
					return NewLoxList(permsList), nil
				}
				return argMustBeTypeAn(in.callToken, "perm", "integer")
			case 2:
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken,
						"First argument to 'Rand().perm' must be an integer.")
				}
				if _, ok := args[2].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'Rand().perm' must be an integer.")
				}

				start := args[1].(int64)
				stop := args[2].(int64)
				if stop < start {
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'Rand().perm' cannot be less than first argument.")
				}
				loxRange := NewLoxRangeStartStop(start, stop+1)
				permsList := list.NewListCap[any](loxRange.Length())
				it := loxRange.Iterator()
				for it.HasNext() {
					permsList.Add(it.Next())
				}
				shuffleFunc := func(a int, b int) {
					permsList[a], permsList[b] = permsList[b], permsList[a]
				}
				if randStruct.rand != nil {
					randStruct.rand.Shuffle(len(permsList), shuffleFunc)
				} else {
					rand.Shuffle(len(permsList), shuffleFunc)
				}
				return NewLoxList(permsList), nil
			default:
				return nil, loxerror.RuntimeError(in.callToken, fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("rand", 0, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			if randStruct.rand != nil {
				return randStruct.rand.Float64(), nil
			}
			return rand.Float64(), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("randBytes", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			if numBytes, ok := args[1].(int64); ok {
				if numBytes < 0 {
					return nil, loxerror.RuntimeError(in.callToken,
						"Argument to 'Rand().randBytes' cannot be negative.")
				}
				buffer := EmptyLoxBufferCap(numBytes)
				if randStruct.rand != nil {
					for i := int64(0); i < numBytes; i++ {
						addErr := buffer.add(randStruct.rand.Int63n(256))
						if addErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
						}
					}
				} else {
					for i := int64(0); i < numBytes; i++ {
						addErr := buffer.add(rand.Int63n(256))
						if addErr != nil {
							return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
						}
					}
				}
				return buffer, nil
			}
			return nil, loxerror.RuntimeError(in.callToken,
				"Argument to 'Rand().randBytes' must be an integer.")
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("randFloat", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			argsLen := len(args) - 1
			switch argsLen {
			case 1:
				switch max := args[1].(type) {
				case int64:
					if max <= 0 {
						return nil, loxerror.RuntimeError(in.callToken, "Argument to 'Rand().randFloat' cannot be 0 or negative.")
					}

					if randStruct.rand != nil {
						return randStruct.rand.Float64() * float64(max), nil
					}
					return rand.Float64() * float64(max), nil
				case float64:
					if max <= 0 {
						return nil, loxerror.RuntimeError(in.callToken, "Argument to 'Rand().randFloat' cannot be 0.0 or negative.")
					}

					if randStruct.rand != nil {
						return randStruct.rand.Float64() * max, nil
					}
					return rand.Float64() * max, nil
				default:
					return nil, loxerror.RuntimeError(in.callToken, "Argument to 'Rand().randFloat' must be an integer or float.")
				}
			case 2:
				secondArgTypeErrMsg := "Second argument to 'Rand().randFloat' must be an integer or float."
				secondArgLessErrMsg := "Second argument to 'Rand().randFloat' cannot be less than first argument."
				switch min := args[1].(type) {
				case int64:
					switch max := args[2].(type) {
					case int64:
						if max < min {
							return nil, loxerror.RuntimeError(in.callToken, secondArgLessErrMsg)
						}
						floatMin := float64(min)
						floatMax := float64(max)

						if randStruct.rand != nil {
							return randStruct.rand.Float64()*(floatMax-floatMin) + floatMin, nil
						}
						return rand.Float64()*(floatMax-floatMin) + floatMin, nil
					case float64:
						floatMin := float64(min)
						if max < floatMin {
							return nil, loxerror.RuntimeError(in.callToken, secondArgLessErrMsg)
						}

						if randStruct.rand != nil {
							return randStruct.rand.Float64()*(max-floatMin) + floatMin, nil
						}
						return rand.Float64()*(max-floatMin) + floatMin, nil
					default:
						return nil, loxerror.RuntimeError(in.callToken, secondArgTypeErrMsg)
					}
				case float64:
					switch max := args[2].(type) {
					case int64:
						floatMax := float64(max)
						if floatMax < min {
							return nil, loxerror.RuntimeError(in.callToken, secondArgLessErrMsg)
						}

						if randStruct.rand != nil {
							return randStruct.rand.Float64()*(floatMax-min) + min, nil
						}
						return rand.Float64()*(floatMax-min) + min, nil
					case float64:
						if max < min {
							return nil, loxerror.RuntimeError(in.callToken, secondArgLessErrMsg)
						}

						if randStruct.rand != nil {
							return randStruct.rand.Float64()*(max-min) + min, nil
						}
						return rand.Float64()*(max-min) + min, nil
					default:
						return nil, loxerror.RuntimeError(in.callToken, secondArgTypeErrMsg)
					}
				default:
					return nil, loxerror.RuntimeError(in.callToken, "First argument to 'Rand().randFloat' must be an integer or float.")
				}
			default:
				return nil, loxerror.RuntimeError(in.callToken, fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("randInt", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			argsLen := len(args) - 1
			switch argsLen {
			case 1:
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken, "Argument to 'Rand().randInt' must be an integer.")
				}
				max := args[1].(int64)
				if max <= 0 {
					return nil, loxerror.RuntimeError(in.callToken, "Argument to 'Rand().randInt' cannot be 0 or negative.")
				}
				if randStruct.rand != nil {
					return randStruct.rand.Int63n(max + 1), nil
				}
				return rand.Int63n(max), nil
			case 2:
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken, "First argument to 'Rand().randInt' must be an integer.")
				}
				if _, ok := args[2].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken, "Second argument to 'Rand().randInt' must be an integer.")
				}
				min := args[1].(int64)
				max := args[2].(int64)
				if max < min {
					return nil, loxerror.RuntimeError(in.callToken, "Second argument to 'Rand().randInt' cannot be less than first argument.")
				}
				if randStruct.rand != nil {
					return randStruct.rand.Int63n(max-min+1) + min, nil
				}
				return rand.Int63n(max-min+1) + min, nil
			default:
				return nil, loxerror.RuntimeError(in.callToken, fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("randRange", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			invalidRangeLenMsg := "Empty range in Rand().randRange."
			argsLen := len(args) - 1
			switch argsLen {
			case 1:
				if stop, ok := args[1].(int64); ok {
					r := NewLoxRangeStop(stop)
					rangeLen := r.Length()
					if rangeLen == 0 {
						return nil, loxerror.RuntimeError(in.callToken, invalidRangeLenMsg)
					}
					if randStruct.rand != nil {
						return r.get(randStruct.rand.Int63n(rangeLen)), nil
					}
					return r.get(rand.Int63n(rangeLen)), nil
				}
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'Rand().randRange' must be an integer.")
			case 2, 3:
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken,
						"First argument to 'Rand().randRange' must be an integer.")
				}
				if _, ok := args[2].(int64); !ok {
					return nil, loxerror.RuntimeError(in.callToken,
						"Second argument to 'Rand().randRange' must be an integer.")
				}
				start := args[1].(int64)
				stop := args[2].(int64)
				var step int64
				if argsLen == 3 {
					if _, ok := args[3].(int64); !ok {
						return nil, loxerror.RuntimeError(in.callToken,
							"Third argument to 'Rand().randRange' must be an integer.")
					}
					step = args[3].(int64)
					if step == 0 {
						return nil, loxerror.RuntimeError(in.callToken,
							"Third argument to 'Rand().randRange' cannot be 0.")
					}
				} else {
					step = 1
				}
				r := NewLoxRange(start, stop, step)
				rangeLen := r.Length()
				if rangeLen == 0 {
					return nil, loxerror.RuntimeError(in.callToken, invalidRangeLenMsg)
				}
				if randStruct.rand != nil {
					return r.get(randStruct.rand.Int63n(rangeLen)), nil
				}
				return r.get(rand.Int63n(rangeLen)), nil
			default:
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Expected 1, 2, or 3 arguments but got %v.", argsLen))
			}
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})
	randInstanceFunc("sample", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		instance := args[0].(*LoxInstance)
		switch randStruct := instance.fields[randStr].(type) {
		case LoxRand:
			if _, ok := args[2].(int64); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'Rand().sample' must be an integer.")
			}
			numSamples := args[2].(int64)
			if numSamples < 0 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'Rand().sample' cannot be negative.")
			}
			if _, ok := args[1].(interfaces.Length); !ok {
				return nil, loxerror.RuntimeError(in.callToken,
					fmt.Sprintf("Cannot get random element from type '%v'.", getType(args[1])))
			}
			arg := args[1].(interfaces.Length)
			argLen := arg.Length()
			if numSamples > argLen {
				return nil, loxerror.RuntimeError(in.callToken,
					"Second argument to 'Rand().sample' cannot be greater than the first argument's length.")
			}
			getIndex := func(index int) any {
				switch arg := arg.(type) {
				case *LoxBuffer:
					return arg.elements[index]
				case *LoxList:
					return arg.elements[index]
				case *LoxRange:
					return arg.get(int64(index))
				case *LoxString:
					return NewLoxStringQuote(string([]rune(arg.str)[index]))
				default:
					return loxerror.Error(
						fmt.Sprintf("Cannot get random element from type '%v'.", getType(arg)),
					)
				}
			}
			samples := list.NewListCap[any](numSamples)
			var randIndexes []int
			if randStruct.rand != nil {
				randIndexes = randStruct.rand.Perm(int(argLen))
			} else {
				randIndexes = rand.Perm(int(argLen))
			}
			for i := int64(0); i < numSamples; i++ {
				element := getIndex(randIndexes[i])
				if i == 0 {
					switch element := element.(type) {
					case error:
						samples.Clear()
						return nil, loxerror.RuntimeError(in.callToken, element.Error())
					}
				}
				samples.Add(element)
			}
			return NewLoxList(samples), nil
		default:
			return nil, loxerror.RuntimeError(in.callToken, randFieldTypeErrMsg)
		}
	})

	i.globals.Define(className, randClass)
}
