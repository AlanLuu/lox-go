package bigint

import "math/big"

var BoolMap = map[bool]*big.Int{
	true:  big.NewInt(1),
	false: big.NewInt(0),
}

func IsNegative(x *big.Int) bool {
	return x.Sign() < 0
}

func IsPositive(x *big.Int) bool {
	return x.Sign() > 0
}

func IsZero(x *big.Int) bool {
	return len(x.Bits()) == 0
}

func String(x *big.Int) string {
	return x.String() + "n"
}
