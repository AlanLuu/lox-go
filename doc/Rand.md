## Rand class methods

The following methods and fields are defined in the built-in `Rand` class:
- (constructor) `Rand([seed])`, which creates a new `Rand` instance with the specified integer seed. If the seed is omitted, the returned `Rand` instance will have a random seed
- (instance method) `Rand().rand()`, which returns a random float between `0` and `1` exclusive
- (instance method) `Rand().randFloat(arg1, [arg2])`, which takes in either 1 or 2 float or integer arguments and returns a random float based on the arguments passed in
    - If only one argument is passed (only `arg1` is specified), then a random float between `0` and `arg1` exclusive is returned, otherwise a random float between `arg1` and `arg2` inclusive is returned
    - If only `arg1` is specified and `arg1` is 0 or negative, or `arg1` and `arg2` are specified and `arg2 < arg1`, a runtime error is thrown
- (instance method) `Rand().randInt(arg1, [arg2])`, which takes in either 1 or 2 integer arguments and returns a random integer based on the arguments passed in
    - If only one argument is passed (only `arg1` is specified), then a random integer between `0` and `arg1` exclusive is returned, otherwise a random integer between `arg1` and `arg2` inclusive is returned
    - If only `arg1` is specified and `arg1` is 0 or negative, or `arg1` and `arg2` are specified and `arg2 < arg1`, a runtime error is thrown
