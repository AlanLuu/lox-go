# bigint and bigfloat methods

The following methods are defined in the built-in `bigint` class:
- `bigint.new(arg)`, which returns a bigint from the specified argument, which is either an integer, float, or string
    - If the argument is a float, the returned bigint's value is the truncated form of the float
- `bigint.toBigFloat(bigintArg)`, which returns a bigfloat with the value of the specified bigint argument
- `bigint.toFloat(bigintArg)`, which returns the integer representation of the specified bigint argument as a float
- `bigint.toInt(bigintArg)`, which returns the integer representation of the specified bigint argument
- `bigint.toString(bigintArg)`, which returns the string representation of the specified bigint argument

The following methods are defined in the built-in `bigfloat` class:
- `bigfloat.new(arg)`, which returns a bigfloat from the specified argument, which is either an integer, float, or string
- `bigfloat.toBigInt(bigFloatArg)`, which returns a bigint with the value of the truncated form of the specified bigfloat argument
- `bigfloat.toFloat(bigFloatArg)`, which returns the float representation of the specified bigfloat argument
- `bigfloat.toInt(bigFloatArg)`, which returns the truncated form of the specified bigfloat argument as an integer
- `bigfloat.toString(bigFloatArg)`, which returns the string representation of the specified bigfloat argument
