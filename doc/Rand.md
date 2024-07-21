## Rand class methods

None of the following methods are suitable for security or cryptographic purposes.

The following methods and fields are defined in the built-in `Rand` class:
- (constructor) `Rand([seed])`, which creates a new `Rand` instance with the specified integer seed. If the seed is omitted, the returned `Rand` instance will have a random seed
- (instance method) `Rand().choice(sequence)`, which returns a random element from `sequence`, where `sequence` is a buffer, list, or string
    - If `sequence` is a string, the random element is a random character from the string as a new string
    - If `sequence` is empty, a runtime error is thrown
- (instance method) `Rand().choices(sequence, numChoices)`, which returns a list of `numChoices` random elements from `sequence` with replacement, where `sequence` is a buffer, list, or string and `numChoices` is an integer
    - If `sequence` is a string, the random element is a random character from the string as a new string
    - If `numChoices` is negative or `sequence` is empty and `numChoices` is not `0`, a runtime error is thrown
- (instance method) `Rand().rand()`, which returns a random float between `0` and `1` exclusive
- (instance method) `Rand().randBytes(size)`, which returns a buffer of `size` random bytes, where `size` is an integer
    - If `size` is negative, a runtime error is thrown
- (instance method) `Rand().randFloat(arg1, [arg2])`, which takes in either 1 or 2 float or integer arguments and returns a random float based on the arguments passed in
    - If only one argument is passed (only `arg1` is specified), then a random float between `0` and `arg1` exclusive is returned, otherwise a random float between `arg1` and `arg2` inclusive is returned
    - If only `arg1` is specified and `arg1` is 0 or negative, or `arg1` and `arg2` are specified and `arg2 < arg1`, a runtime error is thrown
- (instance method) `Rand().randInt(arg1, [arg2])`, which takes in either 1 or 2 integer arguments and returns a random integer based on the arguments passed in
    - If only one argument is passed (only `arg1` is specified), then a random integer between `0` and `arg1` exclusive is returned, otherwise a random integer between `arg1` and `arg2` inclusive is returned
    - If only `arg1` is specified and `arg1` is 0 or negative, or `arg1` and `arg2` are specified and `arg2 < arg1`, a runtime error is thrown
- (instance method) `Rand().randRange(stop)`, which returns a random integer from a range object with a start value of `0`, the specified stop value, and a step value of `1`
    - If the range object with the specified parameters has a length of 0, a runtime error is thrown
- (instance method) `Rand().randRange(start, stop, [step])`, which returns a random integer from a range object with the specified start, stop, and step values
    - If `step` is omitted, the range object will have a step value of `1`
    - If the range object with the specified parameters has a length of 0, a runtime error is thrown
