package ast

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func (i *Interpreter) defineBitFuncs() {
	className := "bit"
	bitClass := NewLoxClass(className, nil, false)
	bitFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native bit class fn %v at %p>", name, &s)
		}
		bitClass.classProperties[name] = s
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'bit.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	bitFunc("add", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bit.add' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bit.add' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.add' must be an integer.")
		}
		x := args[0].(int64)
		y := args[1].(int64)
		carry := args[2].(int64)
		if carry != 0 && carry != 1 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.add' must be equal to 0 or 1.")
		}
		sum, carryOut := bits.Add64(
			uint64(x),
			uint64(y),
			uint64(carry),
		)
		pair := list.NewListCap[any](2)
		pair.Add(int64(sum))
		pair.Add(int64(carryOut))
		return NewLoxList(pair), nil
	})
	bitFunc("div", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bit.div' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bit.div' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.div' must be an integer.")
		}
		hi := uint64(args[0].(int64))
		lo := uint64(args[1].(int64))
		y := uint64(args[2].(int64))
		if y == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.div' cannot be 0.")
		}
		if y <= hi {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.div' cannot be less than or equal to first argument.")
		}
		quo, rem := bits.Div64(hi, lo, y)
		pair := list.NewListCap[any](2)
		pair.Add(int64(quo))
		pair.Add(int64(rem))
		return NewLoxList(pair), nil
	})
	bitFunc("len", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return int64(bits.Len64(uint64(num))), nil
		}
		return argMustBeTypeAn(in.callToken, "len", "integer")
	})
	bitFunc("lenabs", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			if num < 0 {
				num = -num
			}
			return int64(bits.Len64(uint64(num))), nil
		}
		return argMustBeTypeAn(in.callToken, "lenabs", "integer")
	})
	bitFunc("mul", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bit.mul' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bit.mul' must be an integer.")
		}
		x := uint64(args[0].(int64))
		y := uint64(args[1].(int64))
		hi, lo := bits.Mul64(x, y)
		pair := list.NewListCap[any](2)
		pair.Add(int64(hi))
		pair.Add(int64(lo))
		return NewLoxList(pair), nil
	})
	bitFunc("ones", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return int64(bits.OnesCount64(uint64(num))), nil
		}
		return argMustBeTypeAn(in.callToken, "ones", "integer")
	})
	bitFunc("rem", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bit.rem' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bit.rem' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.rem' must be an integer.")
		}
		hi := uint64(args[0].(int64))
		lo := uint64(args[1].(int64))
		y := uint64(args[2].(int64))
		if y == 0 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.rem' cannot be 0.")
		}
		return int64(bits.Rem64(hi, lo, y)), nil
	})
	bitFunc("reverseBits", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return int64(bits.Reverse64(uint64(num))), nil
		}
		return argMustBeTypeAn(in.callToken, "reverseBits", "integer")
	})
	bitFunc("reverseBytes", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return int64(bits.ReverseBytes64(uint64(num))), nil
		}
		return argMustBeTypeAn(in.callToken, "reverseBytes", "integer")
	})
	bitFunc("rotateLeft", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bit.rotateLeft' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bit.rotateLeft' must be an integer.")
		}
		x := uint64(args[0].(int64))
		k := int(args[1].(int64))
		return int64(bits.RotateLeft64(x, k)), nil
	})
	bitFunc("str", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			var str string
			if num < 0 {
				numUnsigned := uint64(num)
				var builder strings.Builder
				for i := uint64(1 << 63); i > 0; i >>= 1 {
					if numUnsigned&i != 0 {
						builder.WriteByte('1')
					} else {
						builder.WriteByte('0')
					}
				}
				str = builder.String()
			} else {
				str = strconv.FormatInt(num, 2)
			}
			return NewLoxString(str, '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "str", "integer")
	})
	bitFunc("strAllBits", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			numUnsigned := uint64(num)
			var builder strings.Builder
			for i := uint64(1 << 63); i > 0; i >>= 1 {
				if numUnsigned&i != 0 {
					builder.WriteByte('1')
				} else {
					builder.WriteByte('0')
				}
			}
			return NewLoxString(builder.String(), '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "strAllBits", "integer")
	})
	bitFunc("sub", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'bit.sub' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'bit.sub' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.sub' must be an integer.")
		}
		x := args[0].(int64)
		y := args[1].(int64)
		borrow := args[2].(int64)
		if borrow != 0 && borrow != 1 {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'bit.sub' must be equal to 0 or 1.")
		}
		diff, borrowOut := bits.Sub64(
			uint64(x),
			uint64(y),
			uint64(borrow),
		)
		pair := list.NewListCap[any](2)
		pair.Add(int64(diff))
		pair.Add(int64(borrowOut))
		return NewLoxList(pair), nil
	})
	bitFunc("zerosLead", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return int64(bits.LeadingZeros64(uint64(num))), nil
		}
		return argMustBeTypeAn(in.callToken, "zerosLead", "integer")
	})
	bitFunc("zerosTrail", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if num, ok := args[0].(int64); ok {
			return int64(bits.TrailingZeros64(uint64(num))), nil
		}
		return argMustBeTypeAn(in.callToken, "zerosTrail", "integer")
	})

	i.globals.Define(className, bitClass)
}
