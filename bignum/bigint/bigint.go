package bigint

import "math/big"

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
