## Rand class methods

None of the following methods are suitable for security or cryptographic purposes.

The following methods and fields are defined in the built-in `Rand` class:
- (constructor) `Rand([seed])`, which creates a new `Rand` instance with the specified integer seed. If the seed is omitted, the returned `Rand` instance will have a random seed
- (instance method) `Rand().choice(sequence)`, which returns a random element from `sequence`, where `sequence` is a bigrange, bitfield, buffer, deque, list, queue, range, ring, or string
    - If `sequence` is a string, the random element is a random character from the string as a new string
    - If `sequence` is empty, a runtime error is thrown
- (instance method) `Rand().choices(sequence, numChoices)`, which returns a list of `numChoices` random elements from `sequence` with replacement, where `sequence` is a bigrange, bitfield, buffer, deque, list, queue, range, ring, or string and `numChoices` is an integer
    - If `sequence` is a string, the random element is a random character from the string as a new string
    - If `numChoices` is negative or `sequence` is empty and `numChoices` is not `0`, a runtime error is thrown
- (instance method) `Rand().flip()`, which simulates flipping a coin and returns `true` if the coin landed on heads and `false` if the coin landed on tails
- (instance method) `Rand().perm(arg1, [arg2])`, which returns a list of a random permutation of all the integers from `arg1` to `arg2` inclusive. If `arg2` is omitted, a random permutation of all the integers from `0` to `arg1` exclusive is returned
    - If only `arg1` is specified and `arg1` is 0 or negative, or `arg1` and `arg2` are specified and `arg2 < arg1`, a runtime error is thrown
- (instance method) `Rand().rand()`, which returns a random float between `0` and `1` exclusive
- (instance method) `Rand().randBytes(size)`, which returns a buffer of `size` random bytes, where `size` is an integer
    - If `size` is negative, a runtime error is thrown
- (instance method) `Rand().randFloat(arg1, [arg2])`, which takes in either 1 or 2 float or integer arguments and returns a random float based on the arguments passed in
    - If only one argument is passed (only `arg1` is specified), then a random float between `0` and `arg1` exclusive is returned, otherwise a random float between `arg1` and `arg2` inclusive is returned
    - If only `arg1` is specified and `arg1` is 0 or negative, or `arg1` and `arg2` are specified and `arg2 < arg1`, a runtime error is thrown
- (instance method) `Rand().randInt(arg1, [arg2])`, which takes in either 1 or 2 integer arguments and returns a random integer based on the arguments passed in
    - If only one argument is passed (only `arg1` is specified), then a random integer between `0` and `arg1` exclusive is returned, otherwise a random integer between `arg1` and `arg2` inclusive is returned
    - If only `arg1` is specified and `arg1` is 0 or negative, or `arg1` and `arg2` are specified and (`arg2 < arg1` or `(arg2 - arg1 + 1) <= 0`), a runtime error is thrown
- (instance method) `Rand().randRange(stop)`, which returns a random integer from a range object with a start value of `0`, the specified stop value, and a step value of `1`
    - If the range object with the specified parameters has a length of 0, a runtime error is thrown
- (instance method) `Rand().randRange(start, stop, [step])`, which returns a random integer from a range object with the specified start, stop, and step values
    - If `step` is omitted, the range object will have a step value of `1`
    - If the range object with the specified parameters has a length of 0, a runtime error is thrown
- (instance method) `Rand().sample(sequence, k)`, which returns a list of `k` random elements from `sequence` without replacement, where `sequence` is a bigrange, bitfield, buffer, deque, list, queue, range, ring, or string and `k` is an integer
    - If `sequence` is a string, the random element is a random character from the string as a new string
    - If `k` is negative or `k` is greater than the number of elements in `sequence` or `sequence` is empty and `k` is not `0`, a runtime error is thrown
- (instance method) `Rand().sampleAll(sequence)`, which returns a list of `len(sequence)` random elements from `sequence` without replacement, where `sequence` is a bigrange, bitfield, buffer, deque, list, queue, range, ring, or string
    - If `sequence` is a string, the random element is a random character from the string as a new string
- (instance method) `Rand().success(num)`, which takes in an integer or float from 0 to 100 inclusive and runs a random simulation that succeeds `num` percent of times and fails `100 - num` percent of times, returning `true` if this method succeeds and `false` otherwise
    - If `num < 0 || num > 100`, this method throws a runtime error
    - Example: `Rand().success(60)` returns `true` 60% of the time and `false` 40% of the time
- (instance method) `Rand().successes(numTimes, num)`, which takes in `numTimes` as an integer and `num` as an integer or float and runs `numTimes` random simulations, where each of them succeeds `num` percent of times and fails `100 - num` percent of times, returning a list of `numTimes` booleans that represent the outcome of those simulations, where `true` is success and `false` is failure
    - If `numTimes < 0`, this method throws a runtime error
    - If `num < 0 || num > 100`, this method throws a runtime error
- (instance method) `Rand().successesPercent(numTimes, percentage)`, which takes in `numTimes` as an integer and a percentage as a float from 0.0 to 1.0 inclusive and runs `numTimes` random simulations, where each of them succeeds `percentage * 100` percent of times and fails `(1 - percentage) * 100` percent of times, returning a list of `numTimes` booleans that represent the outcome of those simulations, where `true` is success and `false` is failure
    - If `numTimes < 0`, this method throws a runtime error
    - If `percentage < 0.0 || percentage > 1.0`, this method throws a runtime error
- (instance method) `Rand().successPercent(percentage)`, which takes in a percentage as a float from 0.0 to 1.0 inclusive and runs a random simulation that succeeds `percentage * 100` percent of times and fails `(1 - percentage) * 100` percent of times, returning `true` if this method succeeds and `false` otherwise
    - If `percentage < 0.0 || percentage > 1.0`, this method throws a runtime error
    - Example: `Rand().successPercent(0.60)` returns `true` 60% of the time and `false` 40% of the time
