# Iterator utility methods and fields

The following methods and fields are defined in the built-in `Iterator` class:
- `Iterator.accumulate(iterable, callback, [initialValue])`, which returns an iterator that returns the intermediate values of applying the specified reducer callback function on every element from the specified iterable
- `Iterator.accumulateAdd(iterable, [initialValue])`, which returns an iterator that returns the intermediate values of applying a reducer callback function on every element from the specified iterable, where the callback function used in this method is the following function: `fun(a, b) => a + b`
- `Iterator.all(iterable)`, which takes in an iterable and returns `true` if all elements from the iterable are truthy values and `false` otherwise
- `Iterator.allFunc(iterable, callback)`, which takes in an iterable and a callback function and returns `true` if the callback function returns `true` for all elements from the iterable and `false` otherwise
- `Iterator.any(iterable)`, which takes in an iterable and returns `true` if any element from the iterable is a truthy value and `false` otherwise
- `Iterator.anyFunc(iterable, callback)`, which takes in an iterable and a callback function and returns `true` if the callback function returns `true` for any element from the iterable and `false` otherwise
- `Iterator.args(args)`, which takes in a variable amount of arguments and returns an iterator that returns those arguments in order from left to right and stops after returning the rightmost argument
- `Iterator.batched(iterable, length)`, which returns an iterator that produces lists of elements from the iterable of the specified length, which is an integer
    ```js
    var list = [1, 2, 3, 4, 5, 6];
    var iterator1 = Iterator.batched(list, 1);
    var iterator2 = Iterator.batched(list, 2);
    var iterator3 = Iterator.batched(list, 4);
    var iterator4 = Iterator.batched(list, 6);
    var iterator5 = Iterator.batched(list, 8);
    print iterator1.toList(); //[[1], [2], [3], [4], [5], [6]]
    print iterator2.toList(); //[[1, 2], [3, 4], [5, 6]]
    print iterator3.toList(); //[[1, 2, 3, 4], [5, 6]]
    print iterator4.toList(); //[[1, 2, 3, 4, 5, 6]]
    print iterator5.toList(); //[[1, 2, 3, 4, 5, 6]]
    ```
    - The value of `length` must be at least `1` or else a runtime error is thrown
- `Iterator.chain(iterables)`, which takes a variable amount of iterables as arguments and returns an iterator that produces elements from the first iterable until there are no more elements in that iterable, then moves on to producing elements from the next iterable, until all iterables are out of elements
- `Iterator.count(iterable, callback)`, which returns an integer that represents the number of elements from the specified iterable where the callback function returns a truthy value for them
    - If the iterable contains more than `(1 << 63) - 1` elements where the callback function returns a truthy value for them, this method stops iterating over the iterable and an integer equal to `(1 << 63) - 1` is returned
- `Iterator.countFloat(start, [step])`, which returns an iterator that returns `start`, with `start` being incremented by `step` after each iteration
    - `start` can be an integer, bigint, float, or bigfloat, and `step` can be a float or bigfloat
    - If `step` is omitted, `1.0` is used as the step value
- `Iterator.countInt(start, [step])`, which returns an iterator that returns `start`, with `start` being incremented by `step` after each iteration
    - `start` and `step` can be integers or bigints
    - If `step` is omitted, `1` is used as the step value
- `Iterator.countTrue(iterable)`, which returns an integer that represents the number of elements from the specified iterable that are truthy values
    - If the iterable contains more than `(1 << 63) - 1` elements that are truthy values, this method stops iterating over the iterable and an integer equal to `(1 << 63) - 1` is returned
- `Iterator.cycle(iterable)`, which returns an iterator that produces elements from the specified iterable, saving each element from the iterable internally. When the iterable is out of elements, this iterator continues to return the saved elements over and over again
- `Iterator.dropuntil(iterable, callback)`, which returns an iterable that starts from the first element, skips over all the elements where the callback function returns a falsy value for them, and returns every element afterwards
- `Iterator.dropwhile(iterable, callback)`, which returns an iterable that starts from the first element, skips over all the elements where the callback function returns a truthy value for them, and returns every element afterwards
- `Iterator.enumerate(iterable, [start])`, which returns an iterator that produces lists with two elements, where the first element is the `start` argument, which is either an integer, bigint, float, or bigfloat and is incremented by 1 for each element in the iterable, and the second element is the current element in the iterable. If `start` is omitted, it defaults to `0`
- `Iterator.filter(iterable, callback)`, which returns an iterator that returns the elements from the specified iterable where the callback function returns a truthy value for them
- `Iterator.filterfalse(iterable, callback)`, which returns an iterator that returns the elements from the specified iterable where the callback function returns a falsy value for them
- `Iterator.filterfalseonly(iterable)`, which returns an iterator that returns the elements from the specified iterable that are falsy values
- `Iterator.filtertrueonly(iterable)`, which returns an iterator that returns the elements from the specified iterable that are truthy values
- `Iterator.getuntil(iterable, callback)`, which returns an iterator that starts from the first element, returns the elements where the callback function returns a falsy value for them, and stops returning elements once the callback function returns a truthy value, without including the element where the callback function returned a truthy value
- `Iterator.getuntillast(iterable, callback)`, which returns an iterator that starts from the first element, returns the elements where the callback function returns a falsy value for them, and stops returning elements once the callback function returns a truthy value, but includes the element where the callback function returned a truthy value
- `Iterator.getwhile(iterable, callback)`, which returns an iterator that starts from the first element, returns the elements where the callback function returns a truthy value for them, and stops returning elements once the callback function returns a falsy value, without including the element where the callback function returned a falsy value
- `Iterator.getwhilelast(iterable, callback)`, which returns an iterator that starts from the first element, returns the elements where the callback function returns a truthy value for them, and stops returning elements once the callback function returns a falsy value, but includes the element where the callback function returned a falsy value
- `Iterator.infiniteArg(arg)`, which returns an iterator that returns endless amounts of the specified argument
- `Iterator.infiniteArgs(args)`, which takes in a variable amount of arguments and returns an iterator that endlessly returns those arguments in order from left to right, going back to the leftmost argument after returning the rightmost argument
- `Iterator.map(iterable, callback)`, which returns an iterator that returns the results of calling a callback function on each element from the specified iterable
- `Iterator.pairwise(iterable)`, which returns an iterator that returns lists of successive overlapping pairs of elements from the specified iterable. If the specified iterable is empty or only has one element when iterated over, this method returns an empty iterator
- `Iterator.reduce(iterable, callback, [initialValue])`, which applies a reducer callback function on every element from the specified iterable and returns a single value
- `Iterator.reduceRight(iterable, callback, [initialValue])`, which applies a reducer callback function on every element from the specified iterable starting with the last element returned from the iterable and ending with the first element returned from the iterable and returns a single value
    - This method uses a temporary list to store all the elements returned from the iterable
- `Iterator.repeat(element, [count])`, which returns an iterator that returns `element` over and over again. If `count` is specified, the returned iterator will only return `element` for a total of `count` times, where `count` is an integer or bigint
- `Iterator.reversed(reverseIterable)`, which takes in an iterable that can be reversed and returns an iterator that produces elements from the iterable in reverse order
    - The following iterables can be reversed: buffers, lists, and strings
- `Iterator.urandom`, which is an iterator that produces endless amounts of cryptographically secure bytes as integers
- `Iterator.zeroes`, which is an iterator that produces endless amounts of the integer `0`
- `Iterator.zip(iterables)`, which takes a variable amount of iterables as arguments and returns an iterator that produces lists containing each element from each iterable and stops when the shortest iterable is out of elements
    - If 1 iterable is passed, the resulting iterator produces lists of 1 element
    - If no arguments are passed, this method returns an empty iterator
