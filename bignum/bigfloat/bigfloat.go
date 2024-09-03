package bigfloat

import (
	"math/big"

	"github.com/AlanLuu/lox/util"
)

var (
	One = big.NewFloat(1.0)
)

var BoolMap = map[bool]*big.Float{
	true:  New(1),
	false: New(0),
}

func New(x float64) *big.Float {
	bigFloat := &big.Float{}
	bigFloat.SetString(util.FormatFloat(x))
	return bigFloat
}

func IsNegative(x *big.Float) bool {
	return x.Sign() < 0
}

func IsPositive(x *big.Float) bool {
	return x.Sign() > 0
}

func IsZero(x *big.Float) bool {
	return x.Sign() == 0
}

func String(x *big.Float) string {
	if x.IsInt() {
		return x.String() + ".0n"
	}
	return x.String() + "n"
}
