# Iterator utility methods and fields

The following methods and fields are defined in the built-in `Iterator` class:
- `Iterator.zeroes`, which is an iterator that produces endless amounts of the integer `0`
- `Iterator.zip(iterables)`, which takes a variable amount of iterables as arguments and returns an iterator that produces lists containing each element from each iterable and stops when the shortest iterable is out of elements
    - If 1 iterable is passed, the resulting iterator produces lists of 1 element
    - If no arguments are passed, this method returns an empty iterator
