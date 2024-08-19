package ast

import "math/big"

type LoxBigNumKey struct {
	str        string
	isFloat    bool
	isFloatInt bool
}

func NewLoxBigIntKey(x *big.Int) LoxBigNumKey {
	return LoxBigNumKey{
		str:        x.String(),
		isFloat:    false,
		isFloatInt: false,
	}
}

func NewLoxBigFloatKey(x *big.Float) LoxBigNumKey {
	return LoxBigNumKey{
		str:        x.String(),
		isFloat:    true,
		isFloatInt: x.IsInt(),
	}
}

func (l LoxBigNumKey) getBigNum() any {
	if l.isFloat {
		bigFloat := &big.Float{}
		bigFloat.SetString(l.str)
		return bigFloat
	} else {
		bigInt := &big.Int{}
		bigInt.SetString(l.str, 0)
		return bigInt
	}
}

func (l LoxBigNumKey) String() string {
	if l.isFloat && l.isFloatInt {
		return l.str + ".0n"
	}
	return l.str + "n"
}
