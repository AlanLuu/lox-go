# bigint and bigfloat methods

The following methods are defined in the built-in `bigint` class:
- `bigint.bitSize(bigintArg)`, which returns the number of bits in the absolute value of the specified bigint argument as an integer
    - The bit size of `0n` is 0
- `bigint.bytes(bigintArg)`, which returns a buffer of the byte representation of the absolute value of the specified bigint argument in big-endian order
- `bigint.new(arg)`, which returns a bigint from the specified argument, which is either an integer, float, or string
    - If the argument is a float, the returned bigint's value is the truncated form of the float
- `bigint.isInt(bigintArg)`, which returns `true` if the specified bigint argument's value can be represented as an integer without overflow and `false` otherwise
- `bigint.probablyPrime(bigintArg, [n])`, which applies the Miller-Rabin primality test to the specified bigint argument with `n` pseudorandomly chosen bases followed by the Baillie-PSW primality test and returns `true` if the bigint argument is prime and possibly `false` otherwise
    - If `n` is negative, this method throws a runtime error
    - If `n` is omitted, the value of `n` is set to `10`
    - A larger `n` value lowers the chances of this method returning `true` for a non-prime argument
- `bigint.toBigFloat(bigintArg)`, which returns a bigfloat with the value of the specified bigint argument
- `bigint.toFloat(bigintArg)`, which returns the integer representation of the specified bigint argument as a float
- `bigint.toHexStr(bigintArg)`, which returns the hexadecimal representation of the specified bigint argument as a string
- `bigint.toInt(bigintArg)`, which returns the integer representation of the specified bigint argument
- `bigint.toString(bigintArg)`, which returns the string representation of the specified bigint argument

The following methods are defined in the built-in `bigfloat` class:
- `bigfloat.new(arg)`, which returns a bigfloat from the specified argument, which is either an integer, float, or string
- `bigfloat.toBigInt(bigFloatArg)`, which returns a bigint with the value of the truncated form of the specified bigfloat argument
- `bigfloat.toFloat(bigFloatArg)`, which returns the float representation of the specified bigfloat argument
- `bigfloat.toInt(bigFloatArg)`, which returns the truncated form of the specified bigfloat argument as an integer
- `bigfloat.toString(bigFloatArg)`, which returns the string representation of the specified bigfloat argument
