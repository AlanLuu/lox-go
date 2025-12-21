package ast

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func defineMiscFuncs() *LoxClass {
	const className = "misc"
	miscClass := NewLoxClass(className, nil, false)
	miscFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native misc fn %v at %p>", name, &s)
		}
		miscClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'misc.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'misc.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	miscFunc("bogosort", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			l := loxList.elements
			lLen := len(l)
			tokenNode := &token.Token{
				TokenType: token.GREATER,
				Lexeme:    ">",
			}
			isSorted := func(l list.List[any]) bool {
				for i := 0; i < lLen-1; i++ {
					result, _ := in.visitBinaryExpr(Binary{
						Literal{l[i]},
						tokenNode,
						Literal{l[i+1]},
					})
					switch result := result.(type) {
					case bool:
						if result {
							return false
						}
					}
				}
				return true
			}
			for !isSorted(l) {
				rand.Shuffle(lLen, func(a int, b int) {
					l[a], l[b] = l[b], l[a]
				})
			}
			return nil, nil
		}
		return argMustBeType(in.callToken, "bogosort", "list")
	})
	miscFunc("bogosorted", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			lLen := len(loxList.elements)
			l := list.NewListCap[any](int64(lLen))
			for _, element := range loxList.elements {
				l.Add(element)
			}
			tokenNode := &token.Token{
				TokenType: token.GREATER,
				Lexeme:    ">",
			}
			isSorted := func(l list.List[any]) bool {
				for i := 0; i < lLen-1; i++ {
					result, _ := in.visitBinaryExpr(Binary{
						Literal{l[i]},
						tokenNode,
						Literal{l[i+1]},
					})
					switch result := result.(type) {
					case bool:
						if result {
							return false
						}
					}
				}
				return true
			}
			for !isSorted(l) {
				rand.Shuffle(lLen, func(a int, b int) {
					l[a], l[b] = l[b], l[a]
				})
			}
			return NewLoxList(l), nil
		}
		return argMustBeType(in.callToken, "bogosorted", "list")
	})
	miscFunc("hello", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		fmt.Println("Hello world!")
		return nil, nil
	})
	miscFunc("multable", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if depth, ok := args[0].(int64); ok {
			if depth <= 0 {
				return nil, nil
			}
			numDigits := func(num int64) int64 {
				if num < 0 {
					num = -num
				}
				var numDigits int64 = 0
				if num == 0 {
					numDigits++
				}
				for num > 0 {
					numDigits++
					num /= 10
				}
				return numDigits
			}
			format := func(result int64) string {
				if result < 0 {
					result = -result
				}
				var numWidth int64 = 4
				if digits := numDigits(result); digits >= numWidth {
					numWidth = digits + 1
					if numWidth < 0 {
						//Integer overflow, use max 64-bit signed value
						numWidth = (1 << 63) - 1
					}
				}
				return "%" + fmt.Sprint(numWidth) + "d"
			}
			for i := int64(1); i <= depth; i++ {
				for j := int64(1); j <= depth; j++ {
					result := i * j
					fmt.Printf(format(result), result)
				}
				fmt.Println()
			}
			return nil, nil
		}
		return argMustBeTypeAn(in.callToken, "multable", "integer")
	})
	miscFunc("multablebig", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var depth *big.Int
		switch arg := args[0].(type) {
		case int64:
			if arg <= 0 {
				return nil, nil
			}
			depth = big.NewInt(arg)
		case *big.Int:
			if arg.Cmp(bigint.Zero) <= 0 {
				return nil, nil
			}
			depth = arg
		default:
			return argMustBeTypeAn(in.callToken, "multablebig", "integer or bigint")
		}
		format := func(result string) string {
			var numWidth int64 = 4
			if l := int64(len(result)); l >= numWidth {
				numWidth = l + 1
				if numWidth < 0 {
					//Integer overflow, use max 64-bit signed value
					numWidth = (1 << 63) - 1
				}
			}
			return "%" + fmt.Sprint(numWidth) + "s"
		}
		one := bigint.One
		result := new(big.Int)
		for i := big.NewInt(1); i.Cmp(depth) <= 0; i.Add(i, one) {
			for j := big.NewInt(1); j.Cmp(depth) <= 0; j.Add(j, one) {
				result = result.Mul(i, j)
				resultStr := result.String()
				fmt.Printf(format(resultStr), resultStr)
			}
			fmt.Println()
		}
		return nil, nil
	})

	return miscClass
}
