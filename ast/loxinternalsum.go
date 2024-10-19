package ast

import (
	"fmt"
	"math/big"

	"github.com/AlanLuu/lox/bignum/bigfloat"
	"github.com/AlanLuu/lox/bignum/bigint"
	"github.com/AlanLuu/lox/loxerror"
)

type LoxInternalSum struct {
	element any
}

func (l *LoxInternalSum) sum(other any) error {
	boolMap := map[bool]float64{
		true:  1,
		false: 0,
	}
	boolMapInt := map[bool]int64{
		true:  1,
		false: 0,
	}
	cannotSum := func(element any) error {
		return loxerror.Error(
			fmt.Sprintf(
				"Cannot sum element '%v'.", getResult(element, element, true),
			),
		)
	}
	switch left := l.element.(type) {
	case int64:
		switch right := other.(type) {
		case int64:
			l.element = left + right
		case float64:
			l.element = float64(left) + right
		case *big.Int:
			l.element = new(big.Int).Add(big.NewInt(left), right)
		case *big.Float:
			l.element = new(big.Float).Add(big.NewFloat(float64(left)), right)
		case bool:
			l.element = left + boolMapInt[right]
		case nil:
		default:
			return cannotSum(right)
		}
	case float64:
		switch right := other.(type) {
		case int64:
			l.element = left + float64(right)
		case float64:
			l.element = left + right
		case *big.Int:
			l.element = new(big.Float).Add(bigfloat.New(left), new(big.Float).SetInt(right))
		case *big.Float:
			l.element = new(big.Float).Add(bigfloat.New(left), right)
		case bool:
			l.element = left + boolMap[right]
		case nil:
		default:
			return cannotSum(right)
		}
	case *big.Int:
		switch right := other.(type) {
		case int64:
			l.element = new(big.Int).Add(left, big.NewInt(right))
		case float64:
			l.element = new(big.Float).Add(new(big.Float).SetInt(left), bigfloat.New(right))
		case *big.Int:
			l.element = new(big.Int).Add(left, right)
		case *big.Float:
			l.element = new(big.Float).Add(new(big.Float).SetInt(left), right)
		case bool:
			l.element = new(big.Int).Add(left, bigint.BoolMap[right])
		case nil:
		default:
			return cannotSum(right)
		}
	case *big.Float:
		switch right := other.(type) {
		case int64:
			l.element = new(big.Float).Add(left, bigfloat.New(float64(right)))
		case float64:
			l.element = new(big.Float).Add(left, bigfloat.New(right))
		case *big.Int:
			l.element = new(big.Float).Add(left, new(big.Float).SetInt(right))
		case *big.Float:
			l.element = new(big.Float).Add(left, right)
		case bool:
			l.element = new(big.Float).Add(left, bigfloat.BoolMap[right])
		case nil:
		default:
			return cannotSum(right)
		}
	default:
		return cannotSum(left)
	}
	return nil
}
