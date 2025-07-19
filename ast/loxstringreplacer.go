package ast

import (
	"fmt"
	"io"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxStringReplacer struct {
	replacer *strings.Replacer
	pairs    []string
	pairsMap map[string]string
	methods  map[string]*struct{ ProtoLoxCallable }
}

func NewLoxStringReplacer(strs ...string) (*LoxStringReplacer, error) {
	if len(strs)%2 == 1 {
		return nil, loxerror.Error(
			"Number of string arguments cannot be an odd number.",
		)
	}
	loxStrReplacer := &LoxStringReplacer{
		replacer: nil,
		pairs:    strs,
		pairsMap: nil,
		methods:  nil,
	}
	if err := loxStrReplacer.initPairsMap(); err != nil {
		return nil, err
	}
	loxStrReplacer.replacer = strings.NewReplacer(strs...)
	loxStrReplacer.methods = make(map[string]*struct{ ProtoLoxCallable })
	return loxStrReplacer, nil
}

func (l *LoxStringReplacer) isEmpty() bool {
	return len(l.pairs) == 0
}

func (l *LoxStringReplacer) pairsLenMustBeEven() error {
	//Should never happen
	if len(l.pairs)%2 == 1 {
		return loxerror.Error(
			"internal error: pairs array length is odd",
		)
	}

	return nil
}

func (l *LoxStringReplacer) pairMapSet(key string, value string) error {
	return l.pairMapSetCheckExists(key, value, true)
}

func (l *LoxStringReplacer) pairMapSetCheckExists(
	key string,
	value string,
	checkExists bool,
) error {
	if checkExists {
		if _, ok := l.pairsMap[key]; ok {
			return loxerror.Error(
				fmt.Sprintf(
					"string '%v' already exists in replacer as key.",
					key,
				),
			)
		}
	}
	l.pairsMap[key] = value
	return nil
}

func (l *LoxStringReplacer) updateReplacer() {
	l.replacer = strings.NewReplacer(l.pairs...)
}

func (l *LoxStringReplacer) addPair(first string, second string) error {
	//Should never happen
	if err := l.pairsLenMustBeEven(); err != nil {
		return err
	}

	if err := l.pairMapSet(first, second); err != nil {
		return err
	}
	l.pairs = append(l.pairs, first, second)
	l.updateReplacer()
	return nil
}

func (l *LoxStringReplacer) deleteAllPairs() bool {
	if len(l.pairs) > 0 {
		l.pairs = []string{}
		l.updateReplacer()
		clear(l.pairsMap)
		return true
	} else if len(l.pairsMap) > 0 {
		clear(l.pairsMap)
	}
	return false
}

func (l *LoxStringReplacer) deletePair(first string, second string) error {
	//Should never happen
	if err := l.pairsLenMustBeEven(); err != nil {
		return err
	}

	pairsLen := int64(len(l.pairs))
	for i := int64(0); i < pairsLen; i += 2 {
		if l.pairs[i] == first && l.pairs[i+1] == second {
			(*list.List[string])(&l.pairs).RemoveIndex(i)
			(*list.List[string])(&l.pairs).RemoveIndex(i)
			l.updateReplacer()
			delete(l.pairsMap, first)
			return nil
		}
	}
	return loxerror.Error(
		fmt.Sprintf(
			"pair '%v' -> '%v' not found.",
			first,
			second,
		),
	)
}

func (l *LoxStringReplacer) deletePairByInt(num int64) error {
	//Should never happen
	if err := l.pairsLenMustBeEven(); err != nil {
		return err
	}

	if !l.strsPosInRange(num) {
		return loxerror.Error(
			fmt.Sprintf(
				"integer argument '%v' out of range.",
				num,
			),
		)
	}
	index := (2 * num) - 2
	removedStr := (*list.List[string])(&l.pairs).RemoveIndex(index)
	if index < l.strsLeni64() {
		(*list.List[string])(&l.pairs).RemoveIndex(index)
	}
	l.updateReplacer()
	delete(l.pairsMap, removedStr)
	return nil
}

func (l *LoxStringReplacer) deletePairByStr(str string) error {
	//Should never happen
	if err := l.pairsLenMustBeEven(); err != nil {
		return err
	}

	pairsLen := int64(len(l.pairs))
	for i := int64(0); i < pairsLen; i += 2 {
		if l.pairs[i] == str {
			(*list.List[string])(&l.pairs).RemoveIndex(i)
			(*list.List[string])(&l.pairs).RemoveIndex(i)
			l.updateReplacer()
			delete(l.pairsMap, str)
			return nil
		}
	}
	return loxerror.Error(
		fmt.Sprintf(
			"no pair found with string '%v' as argument.",
			str,
		),
	)
}

func (l *LoxStringReplacer) deletePairsByValuesBool(values ...string) bool {
	return l.deletePairsByValuesErr(values...) == nil
}

func (l *LoxStringReplacer) deletePairsByValuesErr(values ...string) error {
	if len(values) == 0 {
		return nil
	}
	foundPair := false
	for _, value := range values {
		foundPair2 := false
		for i := int64(len(l.pairs)) - 1; i >= 1; i -= 2 {
			if l.pairs[i] == value {
				foundPair = true
				foundPair2 = true
				(*list.List[string])(&l.pairs).RemoveIndex(i - 1)
				(*list.List[string])(&l.pairs).RemoveIndex(i - 1)
			}
		}
		if foundPair2 {
			for key, pairMapValue := range l.pairsMap {
				if pairMapValue == value {
					delete(l.pairsMap, key)
				}
			}
		}
	}
	if !foundPair {
		if len(values) == 1 {
			return loxerror.Error(
				fmt.Sprintf(
					"no pairs found with string '%v' as value argument.",
					values[0],
				),
			)
		}
		return loxerror.Error(
			"no pairs found with specified strings as value arguments.",
		)
	}
	l.updateReplacer()
	return nil
}

func (l *LoxStringReplacer) getPairByInt(num int64) (pair [2]string, err error) {
	if !l.strsPosInRange(num) {
		err = loxerror.Error(
			fmt.Sprintf(
				"integer argument '%v' out of range.",
				num,
			),
		)
	} else {
		index := (2 * num) - 2
		pair[0] = l.pairs[index]
		if index+1 < l.strsLeni64() {
			pair[1] = l.pairs[index+1]
		}
	}
	return
}

func (l *LoxStringReplacer) getPairByStr(str string) (pair [2]string, ok bool) {
	value, ok := l.getStrByStr(str)
	pair[0] = str
	pair[1] = value
	return
}

func (l *LoxStringReplacer) getStrByStr(str string) (string, bool) {
	value, ok := l.pairsMap[str]
	return value, ok
}

func (l *LoxStringReplacer) initPairsMap() error {
	if l.pairsMap != nil {
		return nil
	}
	l.pairsMap = map[string]string{}
	if len(l.pairs) > 0 {
		var pair [2]string
		for i, str := range l.pairs {
			pair[i%2] = str
			if i%2 == 1 {
				if err := l.pairMapSet(pair[0], pair[1]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (l *LoxStringReplacer) setPair(arg1 int64, arg2 any, arg3 string) error {
	//Should never happen
	if err := l.pairsLenMustBeEven(); err != nil {
		return err
	}

	if !l.strsPosInRange(arg1) {
		return loxerror.Error(
			fmt.Sprintf(
				"first integer argument '%v' out of range.",
				arg1,
			),
		)
	}
	index := (2 * arg1) - 2
	switch arg2 := arg2.(type) {
	case int64:
		switch arg2 {
		case 1:
			if err := l.pairMapSet(arg3, l.pairs[index+1]); err != nil {
				return err
			}
			delete(l.pairsMap, l.pairs[index])
			l.pairs[index] = arg3
		case 2:
			if err := l.pairMapSetCheckExists(l.pairs[index], arg3, false); err != nil {
				return err
			}
			l.pairs[index+1] = arg3
		default:
			return loxerror.Error(
				"second integer argument must be equal to 1 or 2.",
			)
		}
	case string:
		if err := l.pairMapSet(arg2, arg3); err != nil {
			return err
		}
		delete(l.pairsMap, l.pairs[index])
		l.pairs[index] = arg2
		l.pairs[index+1] = arg3
	case *LoxString:
		if err := l.pairMapSet(arg2.str, arg3); err != nil {
			return err
		}
		delete(l.pairsMap, l.pairs[index])
		l.pairs[index] = arg2.str
		l.pairs[index+1] = arg3
	default: //Should never happen
		return loxerror.Error(
			"internal error: second argument must be int64 or string",
		)
	}

	l.updateReplacer()
	return nil
}

func (l *LoxStringReplacer) strsLeni64() int64 {
	return int64(len(l.pairs))
}

func (l *LoxStringReplacer) strsLeni64Half() int64 {
	return l.strsLeni64() / 2
}

func (l *LoxStringReplacer) strsPosInRange(pos int64) bool {
	return pos >= 1 && pos <= l.strsLeni64Half()
}

func (l *LoxStringReplacer) strsIndexInRange(index int64) bool {
	return index >= 0 && index < l.strsLeni64()
}

func (l *LoxStringReplacer) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	stringreplacerFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native string replacer fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string replacer.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'string replacer.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	writeToStr := func() string {
		const article = "a"
		const article2 = "a"
		strs := []string{
			"file",
			"buffered writer",
			"stringbuilder",
			"connection object",
		}
		strsLen := len(strs)
		switch strsLen {
		case 0:
			return ""
		case 1:
			return article + " " + strs[0] + "."
		case 2:
			return article + " " + strs[0] + " or " +
				article2 + " " + strs[1] + "."
		}
		var builder strings.Builder
		builder.WriteString(article)
		builder.WriteRune(' ')
		for i, s := range strs {
			builder.WriteString(s)
			if i < strsLen-1 {
				builder.WriteString(", ")
				if i == strsLen-2 {
					builder.WriteString("or ")
				}
			}
		}
		builder.WriteRune('.')
		return builder.String()
	}
	const internalErrStr = "internal error"
	switch methodName {
	case "addPair":
		return stringreplacerFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string replacer.addPair' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string replacer.addPair' must be a string.")
			}
			first := args[0].(*LoxString).str
			second := args[1].(*LoxString).str
			err := l.addPair(first, second)
			if err != nil {
				return nil, loxerror.RuntimeError(
					name,
					"string replacer.addPair: "+err.Error(),
				)
			}
			return nil, nil
		})
	case "addPairBool":
		return stringreplacerFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string replacer.addPairBool' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string replacer.addPairBool' must be a string.")
			}
			first := args[0].(*LoxString).str
			second := args[1].(*LoxString).str
			err := l.addPair(first, second)
			if err != nil {
				if strings.Contains(err.Error(), internalErrStr) {
					return nil, loxerror.RuntimeError(name,
						"string replacer.addPairBool: "+err.Error())
				}
				return false, nil
			}
			return true, nil
		})
	case "deleteAllPairs", "clearAllPairs":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.deleteAllPairs()
			return nil, nil
		})
	case "deleteAllPairsBool", "clearAllPairsBool":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.deleteAllPairs(), nil
		})
	case "deletePair":
		return stringreplacerFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string replacer.deletePair' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string replacer.deletePair' must be a string.")
			}
			first := args[0].(*LoxString).str
			second := args[1].(*LoxString).str
			err := l.deletePair(first, second)
			if err != nil {
				return nil, loxerror.RuntimeError(name,
					"string replacer.deletePair: "+err.Error())
			}
			return nil, nil
		})
	case "deletePairBool":
		return stringreplacerFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string replacer.deletePairBool' must be a string.")
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string replacer.deletePairBool' must be a string.")
			}
			first := args[0].(*LoxString).str
			second := args[1].(*LoxString).str
			err := l.deletePair(first, second)
			if err != nil {
				if strings.Contains(err.Error(), internalErrStr) {
					return nil, loxerror.RuntimeError(name,
						"string replacer.deletePairBool: "+err.Error())
				}
				return false, nil
			}
			return true, nil
		})
	case "deletePairByInt":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				err := l.deletePairByInt(num)
				if err != nil {
					return nil, loxerror.RuntimeError(name,
						"string replacer.deletePairByInt: "+err.Error())
				}
				return nil, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "deletePairByIntBool":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				err := l.deletePairByInt(num)
				if err != nil {
					if strings.Contains(err.Error(), internalErrStr) {
						return nil, loxerror.RuntimeError(name,
							"string replacer.deletePairByIntBool: "+err.Error())
					}
					return false, nil
				}
				return true, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "deletePairByStr":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				err := l.deletePairByStr(loxStr.str)
				if err != nil {
					return nil, loxerror.RuntimeError(name,
						"string replacer.deletePairByStr: "+err.Error())
				}
				return nil, nil
			}
			return argMustBeType("string")
		})
	case "deletePairByStrBool":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				err := l.deletePairByStr(loxStr.str)
				if err != nil {
					if strings.Contains(err.Error(), internalErrStr) {
						return nil, loxerror.RuntimeError(name,
							"string replacer.deletePairByStrBool: "+err.Error())
					}
					return false, nil
				}
				return true, nil
			}
			return argMustBeType("string")
		})
	case "deletePairsByValues":
		return stringreplacerFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			values := make([]string, 0, argsLen)
			for i, arg := range args {
				switch arg := arg.(type) {
				case *LoxString:
					values = append(values, arg.str)
				default:
					values = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument #%v in "+
								"'string replacer.deletePairsByValues' "+
								"must be a string.",
							i+1,
						),
					)
				}
			}
			err := l.deletePairsByValuesErr(values...)
			if err != nil {
				return nil, loxerror.RuntimeError(name,
					"string replacer.deletePairsByValues: "+err.Error())
			}
			return nil, nil
		})
	case "deletePairsByValuesBool":
		return stringreplacerFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, loxerror.RuntimeError(name,
					"Expected at least 1 argument but got 0.")
			}
			values := make([]string, 0, argsLen)
			for i, arg := range args {
				switch arg := arg.(type) {
				case *LoxString:
					values = append(values, arg.str)
				default:
					values = nil
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument #%v in "+
								"'string replacer.deletePairsByValuesBool' "+
								"must be a string.",
							i+1,
						),
					)
				}
			}
			return l.deletePairsByValuesBool(values...), nil
		})
	case "deletePairsByValuesList":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				values := make([]string, 0, len(loxList.elements))
				for i, element := range loxList.elements {
					switch element := element.(type) {
					case *LoxString:
						values = append(values, element.str)
					default:
						values = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"List element at index #%v in "+
									"'string replacer.deletePairsByValuesListBool' "+
									"must be a string.",
								i,
							),
						)
					}
				}
				err := l.deletePairsByValuesErr(values...)
				if err != nil {
					return nil, loxerror.RuntimeError(name,
						"string replacer.deletePairsByValuesList: "+err.Error())
				}
				return nil, nil
			}
			return argMustBeType("list")
		})
	case "deletePairsByValuesListBool":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxList, ok := args[0].(*LoxList); ok {
				values := make([]string, 0, len(loxList.elements))
				for i, element := range loxList.elements {
					switch element := element.(type) {
					case *LoxString:
						values = append(values, element.str)
					default:
						values = nil
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"List element at index #%v in "+
									"'string replacer.deletePairsByValuesListBool' "+
									"must be a string.",
								i,
							),
						)
					}
				}
				return l.deletePairsByValuesBool(values...), nil
			}
			return argMustBeType("list")
		})
	case "getPairByInt":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				pairArr, err := l.getPairByInt(num)
				if err != nil {
					return nil, loxerror.RuntimeError(name,
						"string replacer.getPairByInt: "+err.Error())
				}
				pairList := list.NewListCap[any](2)
				pairList.Add(NewLoxStringQuote(pairArr[0]))
				pairList.Add(NewLoxStringQuote(pairArr[1]))
				return NewLoxList(pairList), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getPairByIntOrNil":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				pairArr, err := l.getPairByInt(num)
				if err != nil {
					return nil, nil
				}
				pairList := list.NewListCap[any](2)
				pairList.Add(NewLoxStringQuote(pairArr[0]))
				pairList.Add(NewLoxStringQuote(pairArr[1]))
				return NewLoxList(pairList), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getPairByStr":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				pairArr, pairOk := l.getPairByStr(str)
				if !pairOk {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"string replacer.getPairByStr: no pair found with string '%v' as argument.",
							str,
						),
					)
				}
				pair := list.NewListCap[any](2)
				pair.Add(NewLoxStringQuote(pairArr[0]))
				pair.Add(NewLoxStringQuote(pairArr[1]))
				return NewLoxList(pair), nil
			}
			return argMustBeType("string")
		})
	case "getPairByStrOrNil":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				pairArr, pairOk := l.getPairByStr(loxStr.str)
				if !pairOk {
					return nil, nil
				}
				pair := list.NewListCap[any](2)
				pair.Add(NewLoxStringQuote(pairArr[0]))
				pair.Add(NewLoxStringQuote(pairArr[1]))
				return NewLoxList(pair), nil
			}
			return argMustBeType("string")
		})
	case "getStrAtIndex":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if index, ok := args[0].(int64); ok {
				if !l.strsIndexInRange(index) {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"string replacer.getStrAtIndex: index argument '%v' out of range.",
							index,
						),
					)
				}
				return NewLoxStringQuote(l.pairs[index]), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getStrAtIndexOrNil":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if index, ok := args[0].(int64); ok {
				if !l.strsIndexInRange(index) {
					return nil, nil
				}
				return NewLoxStringQuote(l.pairs[index]), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getStrByStr":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				str := loxStr.str
				getStr, getStrOk := l.getStrByStr(str)
				if !getStrOk {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"string replacer.getStrByStr: no string found with string '%v' as argument.",
							str,
						),
					)
				}
				return NewLoxStringQuote(getStr), nil
			}
			return argMustBeType("string")
		})
	case "getStrByStrOrNil":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				getStr, getStrOk := l.getStrByStr(loxStr.str)
				if !getStrOk {
					return nil, nil
				}
				return NewLoxStringQuote(getStr), nil
			}
			return argMustBeType("string")
		})
	case "isEmpty":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isEmpty(), nil
		})
	case "listIter":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			index := 0
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				return index < len(l.pairs)
			}
			iterator.nextMethod = func() any {
				str := l.pairs[index]
				index++
				return NewLoxStringQuote(str)
			}
			return NewLoxIterator(iterator), nil
		})
	case "numPairs":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.strsLeni64Half(), nil
		})
	case "pairsDict":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			for key, value := range l.pairsMap {
				dict.setKeyValue(NewLoxStringQuote(key), NewLoxStringQuote(value))
			}
			return dict, nil
		})
	case "pairsListIter", "pairsIter":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			i1 := 0
			i2 := -1
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				i2 += 2
				return i2 < len(l.pairs)
			}
			iterator.nextMethod = func() any {
				pair := list.NewListCap[any](2)
				pair.Add(NewLoxStringQuote(l.pairs[i1]))
				pair.Add(NewLoxStringQuote(l.pairs[i2]))
				i1 += 2
				return NewLoxList(pair)
			}
			return NewLoxIterator(iterator), nil
		})
	case "pairsList", "pairs":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[any](l.strsLeni64Half())
			pair := list.NewListCap[any](2)
			for i, str := range l.pairs {
				pair.Add(NewLoxStringQuote(str))
				if i%2 == 1 {
					pairs.Add(NewLoxList(pair))
					pair = list.NewListCap[any](2)
				}
			}
			return NewLoxList(pairs), nil
		})
	case "printPairs":
		return stringreplacerFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var builder strings.Builder
			for i, str := range l.pairs {
				if i%2 == 0 {
					builder.WriteString("[\"")
					builder.WriteString(strings.ReplaceAll(str, "\"", "\\\""))
					builder.WriteString("\", ")
				} else {
					builder.WriteRune('"')
					builder.WriteString(strings.ReplaceAll(str, "\"", "\\\""))
					builder.WriteString("\"]\n")
				}
			}
			if builder.Len() > 0 {
				fmt.Print(builder.String())
			}
			return nil, nil
		})
	case "replace":
		return stringreplacerFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxStr, ok := args[0].(*LoxString); ok {
				return NewLoxStringQuote(l.replacer.Replace(loxStr.str)), nil
			}
			return argMustBeType("string")
		})
	case "setPair":
		return stringreplacerFunc(3, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string replacer.setPair' must be an integer.")
			}
			switch args[1].(type) {
			case int64:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string replacer.setPair' must be an integer or string.")
			}
			if _, ok := args[2].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Third argument to 'string replacer.setPair' must be a string.")
			}
			if arg2, ok := args[1].(int64); ok {
				switch arg2 {
				case 1, 2:
				default:
					return nil, loxerror.RuntimeError(name,
						"Second integer argument to 'string replacer.setPair' must be either the value 1 or 2.")
				}
			}

			arg1 := args[0].(int64)
			var arg2 any
			switch arg := args[1].(type) {
			case int64:
				arg2 = arg
			case *LoxString:
				arg2 = arg.str
			}
			arg3 := args[2].(*LoxString).str
			err := l.setPair(arg1, arg2, arg3)
			if err != nil {
				return nil, loxerror.RuntimeError(name,
					"string replacer.setPair: "+err.Error())
			}
			return nil, nil
		})
	case "setPairBool":
		return stringreplacerFunc(3, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'string replacer.setPairBool' must be an integer.")
			}
			switch args[1].(type) {
			case int64:
			case *LoxString:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'string replacer.setPairBool' must be an integer or string.")
			}
			if _, ok := args[2].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(name,
					"Third argument to 'string replacer.setPairBool' must be a string.")
			}
			if arg2, ok := args[1].(int64); ok {
				switch arg2 {
				case 1, 2:
				default:
					return nil, loxerror.RuntimeError(name,
						"Second integer argument to 'string replacer.setPairBool' must be either the value 1 or 2.")
				}
			}

			arg1 := args[0].(int64)
			var arg2 any
			switch arg := args[1].(type) {
			case int64:
				arg2 = arg
			case *LoxString:
				arg2 = arg.str
			}
			arg3 := args[2].(*LoxString).str
			err := l.setPair(arg1, arg2, arg3)
			if err != nil {
				if strings.Contains(err.Error(), internalErrStr) {
					return nil, loxerror.RuntimeError(name,
						"string replacer.setPairBool: "+err.Error())
				}
				return false, nil
			}
			return true, nil
		})
	case "writeTo":
		return stringreplacerFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			switch args[0].(type) {
			case *LoxFile:
			case *LoxBufWriter:
			case *LoxStringBuilder:
			case *LoxConnection:
			default:
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"First argument to 'string replacer.%v' must be "+
							writeToStr(),
						methodName,
					),
				)
			}
			if _, ok := args[1].(*LoxString); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Second argument to 'string replacer.%v' must be a string.",
						methodName,
					),
				)
			}
			var writer io.Writer
			switch arg := args[0].(type) {
			case *LoxFile:
				if !arg.isWrite() && !arg.isAppend() {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"File argument to 'string replacer.%v' must be in write or append mode.",
							methodName,
						),
					)
				}
				writer = arg.file
			case *LoxBufWriter:
				writer = arg.writer
			case *LoxStringBuilder:
				writer = &arg.builder
			case *LoxConnection:
				writer = arg.conn
			}
			loxStr := args[1].(*LoxString)
			numBytes, err := l.replacer.WriteString(writer, loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(name, err.Error())
			}
			return int64(numBytes), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "String replacers have no property called '"+methodName+"'.")
}

func (l *LoxStringReplacer) String() string {
	if l.isEmpty() {
		return fmt.Sprintf("<empty string replacer object at %p>", l)
	}
	return fmt.Sprintf("<string replacer object at %p>", l)
}

func (l *LoxStringReplacer) Type() string {
	return "string replacer"
}
