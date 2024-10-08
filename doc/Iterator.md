# Iterator utility methods and fields

The following methods and fields are defined in the built-in `Iterator` class:
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
- `Iterator.countFloat(start, [step])`, which returns an iterator that returns `start`, with `start` being incremented by `step` after each iteration
    - `start` can be an integer, bigint, float, or bigfloat, and `step` can be a float or bigfloat
    - If `step` is omitted, `1.0` is used as the step value
- `Iterator.countInt(start, [step])`, which returns an iterator that returns `start`, with `start` being incremented by `step` after each iteration
    - `start` and `step` can be integers or bigints
    - If `step` is omitted, `1` is used as the step value
- `Iterator.cycle(iterable)`, which returns an iterator that produces elements from the specified iterable, saving each element from the iterable internally. When the iterable is out of elements, this iterator continues to return the saved elements over and over again
- `Iterator.enumerate(iterable, [start])`, which returns an iterator that produces lists with two elements, where the first element is the `start` argument, which is either an integer, bigint, float, or bigfloat and is incremented by 1 for each element in the iterable, and the second element is the current element in the iterable. If `start` is omitted, it defaults to `0`
- `Iterator.infiniteArg(arg)`, which returns an iterator that returns endless amounts of the specified argument
- `Iterator.infiniteArgs(args)`, which takes in a variable amount of arguments and returns an iterator that endlessly returns those arguments in order from left to right, going back to the leftmost argument after returning the rightmost argument
- `Iterator.pairwise(iterable)`, which returns an iterator that returns lists of successive overlapping pairs of elements from the specified iterable. If the specified iterable is empty or only has one element when iterated over, this method returns an empty iterator
- `Iterator.repeat(element, [count])`, which returns an iterator that returns `element` over and over again. If `count` is specified, the returned iterator will only return `element` for a total of `count` times, where `count` is an integer or bigint
- `Iterator.reversed(reverseIterable)`, which takes in an iterable that can be reversed and returns an iterator that produces elements from the iterable in reverse order
    - The following iterables can be reversed: buffers, lists, and strings
- `Iterator.urandom`, which is an iterator that produces endless amounts of cryptographically secure bytes as integers
- `Iterator.zeroes`, which is an iterator that produces endless amounts of the integer `0`
- `Iterator.zip(iterables)`, which takes a variable amount of iterables as arguments and returns an iterator that produces lists containing each element from each iterable and stops when the shortest iterable is out of elements
    - If 1 iterable is passed, the resulting iterator produces lists of 1 element
    - If no arguments are passed, this method returns an empty iterator
