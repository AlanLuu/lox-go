# bigint and bigfloat mathematical methods

For the purposes of this document, if `bignum` is specified as an argument, that means a bigint or a bigfloat can be passed in for that argument

The following methods are defined in the built-in `bigmath` class:
- `bigmath.abs(bignum)`, which returns the absolute value of `bignum`
- `bigmath.ceil(bignum)`, which returns the smallest bigint greater than or equal to `bignum`
- `bigmath.dim(bignum1, bignum2)`, which returns the maximum of `bignum1 - bignum2` or `0n` if `bignum1 - bignum2` is negative and `bignum1` and `bignum2` are both bigints, otherwise `0.0n` is returned
- `bigmath.divmod(bigint1, bigint2)`, which returns a list with two elements, with the first being the quotient of `bigint1` and `bigint2` as a bigint and the second being the modulus of `bigint1` and `bigint2` as a bigint
    - If `bigint2` is `0n`, this method throws a runtime error
- `bigmath.factorial(bigint)`, which returns the factorial of the specified bigint argument as a bigint
    - If the bigint argument is negative, this method throws a runtime error
- `bigmath.floor(bignum)`, which returns the largest bigint less than or equal to `bignum`
- `bigmath.gcd(bigint1, bigint2)`, which returns the GCD of the two bigint arguments as a bigint
    - `bigmath.gcd(0n, 0n) == 0n`
- `bigmath.lcm(bigint1, bigint2)`, which returns the LCM of the two bigint arguments as a bigint
    - `bigmath.lcm(0n, 0n) == 0n`
- `bigmath.mantexp(bigfloat)`, which returns a list with two elements, with the first being the mantissa part of the specified bigfloat argument as a bigfloat and the second being the exponent part as an integer
- `bigmath.max(bignum1, bignum2)`, which returns the largest of `bignum1` and `bignum2`
- `bigmath.min(bignum1, bignum2)`, which returns the smallest of `bignum1` and `bignum2`
- `bigmath.mod(bigint1, bigint2)`, which returns the result of the modulus operation on `bigint1` and `bigint2` as a bigint
    - If `bigint2` is `0n`, this method throws a runtime error
- `bigmath.quorem(bigint1, bigint2)`, which returns a list with two elements, with the first being the quotient of `bigint1` and `bigint2` as a bigint and the second being the remainder of `bigint1` and `bigint2` as a bigint
    - If `bigint2` is `0n`, this method throws a runtime error
- `bigmath.round(bignum)`, which returns a bigint of `bignum` rounded to the nearest bigint
- `bigmath.sqrt(bignum)`, which returns the square root of `bignum` as a bigfloat
    - If `bignum` is negative, this method throws a runtime error
- `bigmath.sqrtint(bigint)`, which returns the integer square root of `bigint` as a bigint, which is the largest bigint value `x` such that `x ** 2n <= bigint`
    - If `bigint` is negative, this method throws a runtime error
