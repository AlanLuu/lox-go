package ast

import (
	"fmt"
	"math"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxBitFieldIteratorLSB struct {
	field uint8
	pos   int8
}

func (l *LoxBitFieldIteratorLSB) HasNext() bool {
	return l.pos <= 7
}

func (l *LoxBitFieldIteratorLSB) Next() any {
	value := l.field & (1 << l.pos)
	l.pos++
	if value != 0 {
		return int64(1)
	}
	return int64(0)
}

type LoxBitFieldIteratorMSB struct {
	field uint8
	pos   int8
}

func (l *LoxBitFieldIteratorMSB) HasNext() bool {
	return l.pos >= 0
}

func (l *LoxBitFieldIteratorMSB) Next() any {
	value := l.field & (1 << l.pos)
	l.pos--
	if value != 0 {
		return int64(1)
	}
	return int64(0)
}

type LoxBitField struct {
	field   uint8
	methods map[string]*struct{ ProtoLoxCallable }
}

func EmptyLoxBitField() *LoxBitField {
	return NewLoxBitField(0)
}

func NewLoxBitField(field uint8) *LoxBitField {
	return &LoxBitField{
		field:   field,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxBitFieldBools(bools [8]bool) *LoxBitField {
	var field uint8 = 0
	for i, b := range bools {
		if b {
			field += uint8(math.Exp2(float64(7 - i)))
		}
	}
	return NewLoxBitField(field)
}

func (l *LoxBitField) fieldStr() string {
	return fmt.Sprintf("%.8b", l.field)
}

func (l *LoxBitField) fieldStr0b() string {
	return "0b" + l.fieldStr()
}

func (l *LoxBitField) getBit(pos int64) uint8 {
	return l.getBitIndex(pos - 1)
}

func (l *LoxBitField) getBitIndex(pos int64) uint8 {
	if l.field&(1<<pos) != 0 {
		return 1
	}
	return 0
}

func (l *LoxBitField) getBitIndexMSB(pos int64) uint8 {
	return l.getBitIndex(7 - pos)
}

func (l *LoxBitField) getBitMSB(pos int64) uint8 {
	return l.getBit(7 - pos)
}

func (l *LoxBitField) reversed() uint8 {
	return reverseUint8(l.field)
}

func (l *LoxBitField) reversedStr() string {
	return fmt.Sprintf("%.8b", l.reversed())
}

func (l *LoxBitField) reversedStr0b() string {
	return "0b" + l.reversedStr()
}

func (l *LoxBitField) setBit(pos int64, setBit bool) {
	l.setBitIndex(pos-1, setBit)
}

func (l *LoxBitField) setBitIndex(pos int64, setBit bool) {
	if setBit {
		l.field |= (1 << pos)
	} else {
		l.field &= ^(1 << pos)
	}
}

func (l *LoxBitField) signed() int8 {
	return int8(l.field)
}

func (l *LoxBitField) toggleBit(pos int64) {
	l.toggleBitIndex(pos - 1)
}

func (l *LoxBitField) toggleBitIndex(pos int64) {
	l.field ^= (1 << pos)
}

func (l *LoxBitField) toggleBitIndexMSB(pos int64) {
	l.toggleBitIndex(7 - pos)
}

func (l *LoxBitField) toggleBitMSB(pos int64) {
	l.toggleBit(7 - pos)
}

func (l *LoxBitField) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if field, ok := l.methods[methodName]; ok {
		return field, nil
	}
	bitFieldFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native bitfield fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bitfield.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bitfield.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "add":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(l.field + other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "addSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field += other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "and":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(l.field & other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "andSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field &= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "andNot":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(l.field &^ other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "andNotSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field &^= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "cmp":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				switch {
				case l.field < other_l.field:
					return int64(-1), nil
				case l.field > other_l.field:
					return int64(1), nil
				default:
					return int64(0), nil
				}
			}
			return argMustBeType("bitfield")
		})
	case "div":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				if other_l.field == 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.div' cannot have a field that is 0.")
				}
				return NewLoxBitField(l.field / other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "divSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				if other_l.field == 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.divSelf' cannot have a field that is 0.")
				}
				l.field /= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "eq":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return l.field == other_l.field, nil
			}
			return argMustBeType("bitfield")
		})
	case "field", "value":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.field), nil
		})
	case "fieldBool", "valueBool":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.field != 0, nil
		})
	case "fieldReversed", "valueReversed":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.reversed()), nil
		})
	case "fieldSigned", "valueSigned":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.signed()), nil
		})
	case "fieldStr", "valueStr":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.fieldStr()), nil
		})
	case "fieldStr0b", "valueStr0b":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.fieldStr0b()), nil
		})
	case "fieldStrReversed", "valueStrReversed":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.reversedStr()), nil
		})
	case "fieldStrReversed0b", "valueStrReversed0b":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.reversedStr0b()), nil
		})
	case "getBit":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 1 || num > 8 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBit' must be from 1 to 8.")
				}
				return int64(l.getBit(num)), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getBitMSB":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 1 || num > 8 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBitMSB' must be from 1 to 8.")
				}
				return int64(l.getBitMSB(num)), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getBitBool":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 1 || num > 8 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBitBool' must be from 1 to 8.")
				}
				return l.getBit(num) != 0, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getBitBoolMSB":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 1 || num > 8 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBitBoolMSB' must be from 1 to 8.")
				}
				return l.getBitMSB(num) != 0, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getBitIndex":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 || num > 7 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBitIndex' must be from 0 to 7.")
				}
				return int64(l.getBitIndex(num)), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getBitIndexMSB":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 || num > 7 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBitIndexMSB' must be from 0 to 7.")
				}
				return int64(l.getBitIndexMSB(num)), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getBitIndexBool":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 || num > 7 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBitIndexBool' must be from 0 to 7.")
				}
				return l.getBitIndex(num) != 0, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "getBitIndexBoolMSB":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 || num > 7 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.getBitIndexBoolMSB' must be from 0 to 7.")
				}
				return l.getBitIndexMSB(num) != 0, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "gt":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return l.field > other_l.field, nil
			}
			return argMustBeType("bitfield")
		})
	case "gte":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return l.field >= other_l.field, nil
			}
			return argMustBeType("bitfield")
		})
	case "imply":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(^l.field | other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "implySelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field = ^l.field | other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "isMax":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.field == 255, nil
		})
	case "isMin":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.field == 0, nil
		})
	case "lt":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return l.field < other_l.field, nil
			}
			return argMustBeType("bitfield")
		})
	case "lte":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return l.field <= other_l.field, nil
			}
			return argMustBeType("bitfield")
		})
	case "mod":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				if other_l.field == 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.mod' cannot have a field that is 0.")
				}
				return NewLoxBitField(l.field % other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "modSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				if other_l.field == 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.modSelf' cannot have a field that is 0.")
				}
				l.field %= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "mul":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(l.field * other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "mulSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field *= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "nand":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(^(l.field & other_l.field)), nil
			}
			return argMustBeType("bitfield")
		})
	case "nandSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field = ^(l.field & other_l.field)
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "ne":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return l.field != other_l.field, nil
			}
			return argMustBeType("bitfield")
		})
	case "nimply":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(^(^l.field | other_l.field)), nil
			}
			return argMustBeType("bitfield")
		})
	case "nimplySelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field = ^(^l.field | other_l.field)
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "nor":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(^(l.field | other_l.field)), nil
			}
			return argMustBeType("bitfield")
		})
	case "norSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field = ^(l.field | other_l.field)
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "not":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxBitField(^l.field), nil
		})
	case "notSelf":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.field = ^l.field
			return l, nil
		})
	case "or":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(l.field | other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "orSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field |= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "reversed":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxBitField(l.reversed()), nil
		})
	case "reversedSelf":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.field = l.reversed()
			return l, nil
		})
	case "setBit":
		return bitFieldFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'bitfield.setBit' must be an integer.")
			}
			switch args[1].(type) {
			case bool:
			case int64:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'bitfield.setBit' must be a boolean or integer.")
			}
			num := args[0].(int64)
			if num < 1 || num > 8 {
				return nil, loxerror.RuntimeError(name,
					"First integer argument to 'bitfield.setBit' must be from 1 to 8.")
			}
			switch arg := args[1].(type) {
			case bool:
				l.setBit(num, arg)
			case int64:
				switch arg {
				case 0:
					l.setBit(num, false)
				case 1:
					l.setBit(num, true)
				default:
					return nil, loxerror.RuntimeError(
						name,
						"Second integer argument to 'bitfield.setBit' "+
							"must be either the value 0 or 1.",
					)
				}
			}
			return l, nil
		})
	case "setBitIndex":
		return bitFieldFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'bitfield.setBitIndex' must be an integer.")
			}
			switch args[1].(type) {
			case bool:
			case int64:
			default:
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'bitfield.setBitIndex' must be a boolean or integer.")
			}
			num := args[0].(int64)
			if num < 1 || num > 8 {
				return nil, loxerror.RuntimeError(name,
					"First integer argument to 'bitfield.setBitIndex' must be from 1 to 8.")
			}
			switch arg := args[1].(type) {
			case bool:
				l.setBitIndex(num, arg)
			case int64:
				switch arg {
				case 0:
					l.setBitIndex(num, false)
				case 1:
					l.setBitIndex(num, true)
				default:
					return nil, loxerror.RuntimeError(
						name,
						"Second integer argument to 'bitfield.setBitIndex' "+
							"must be either the value 0 or 1.",
					)
				}
			}
			return l, nil
		})
	case "setField":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 || num > 255 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.setField' must be an integer between 0 and 255.")
				}
				l.field = uint8(num)
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "shl":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.shl' cannot be negative.")
				}
				return NewLoxBitField(l.field << num), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "shlSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.shlSelf' cannot be negative.")
				}
				l.field <<= num
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "shr":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.shr' cannot be negative.")
				}
				return NewLoxBitField(l.field >> num), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "shrSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.shrSelf' cannot be negative.")
				}
				l.field >>= num
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "sub":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(l.field - other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "subSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field -= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "toBuffer":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			buffer := EmptyLoxBufferCap(8)
			for i := int64(0); i <= 7; i++ {
				addErr := buffer.add(int64(l.getBitIndex(i)))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "toBufferMSB":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			buffer := EmptyLoxBufferCap(8)
			for i := int64(7); i >= 0; i-- {
				addErr := buffer.add(int64(l.getBitIndex(i)))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "toggleBit":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 1 || num > 8 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.toggleBit' must be from 1 to 8.")
				}
				l.toggleBit(num)
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "toggleBitMSB":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 1 || num > 8 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.toggleBitMSB' must be from 1 to 8.")
				}
				l.toggleBitMSB(num)
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "toggleBitIndex":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 || num > 7 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.toggleBitIndex' must be from 0 to 7.")
				}
				l.toggleBitIndex(num)
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "toggleBitIndexMSB":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if num, ok := args[0].(int64); ok {
				if num < 0 || num > 7 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'bitfield.toggleBitIndexMSB' must be from 0 to 7.")
				}
				l.toggleBitIndexMSB(num)
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "toListBools":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			boolsList := list.NewListCap[any](8)
			for i := int64(0); i <= 7; i++ {
				boolsList.Add(l.getBitIndex(i) != 0)
			}
			return NewLoxList(boolsList), nil
		})
	case "toListBoolsMSB":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			boolsList := list.NewListCap[any](8)
			for i := int64(7); i >= 0; i-- {
				boolsList.Add(l.getBitIndex(i) != 0)
			}
			return NewLoxList(boolsList), nil
		})
	case "toListInts":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			intsList := list.NewListCap[any](8)
			for i := int64(0); i <= 7; i++ {
				intsList.Add(int64(l.getBitIndex(i)))
			}
			return NewLoxList(intsList), nil
		})
	case "toListIntsMSB":
		return bitFieldFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			intsList := list.NewListCap[any](8)
			for i := int64(7); i >= 0; i-- {
				intsList.Add(int64(l.getBitIndex(i)))
			}
			return NewLoxList(intsList), nil
		})
	case "xnor":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(^(l.field ^ other_l.field)), nil
			}
			return argMustBeType("bitfield")
		})
	case "xnorSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field = ^(l.field ^ other_l.field)
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	case "xor":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				return NewLoxBitField(l.field ^ other_l.field), nil
			}
			return argMustBeType("bitfield")
		})
	case "xorSelf":
		return bitFieldFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if other_l, ok := args[0].(*LoxBitField); ok {
				l.field ^= other_l.field
				return l, nil
			}
			return argMustBeType("bitfield")
		})
	}
	return nil, loxerror.RuntimeError(name, "Bitfields have no property called '"+methodName+"'.")
}

func (l *LoxBitField) Iterator() interfaces.Iterator {
	return &LoxBitFieldIteratorLSB{l.field, 0}
}

func (l *LoxBitField) Length() int64 {
	return 8
}

func (l *LoxBitField) ReverseIterator() interfaces.Iterator {
	return &LoxBitFieldIteratorMSB{l.field, 7}
}

func (l *LoxBitField) String() string {
	return fmt.Sprintf(
		"<bitfield field=%.8b at %p>",
		l.field,
		l,
	)
}

func (l *LoxBitField) Type() string {
	return "bitfield"
}
