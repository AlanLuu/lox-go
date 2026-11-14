# Ring methods and fields

Ring objects are circular sequences. A ring can be indexed with an integer or bigint, where index `0` is the current element and positive indexes moving forward and negative indexes moving backwards. Indexing wraps around the ring, so there are no out-of-range index values. Using an index that is not an integer or bigint causes a runtime error.

Rings are iterable. Iteration starts at the current element and moves forward for one full cycle. Reverse iteration also starts at the current element, but moves backward for one full cycle.

Ring objects have the following fields associated with them:
- `ring.next`, which is the next ring element
- `ring.prev`, which is the previous ring element
- `ring.value`, which is the value of the current ring element

Rings have the following methods associated with them:
- `ring.extend(length)`, which adds `length` new elements with initial values of `nil` after the current element and returns `nil`
    - `length` must be an integer greater than `0`, or else a runtime error is thrown
- `ring.extendRet(length)`, which adds `length` new elements with initial values of `nil` after the current element and returns the current ring element
    - `length` must be an integer greater than `0`, or else a runtime error is thrown
- `ring.filter(callback)`, which returns a new ring containing the elements for which `callback(element, index)` returns a truthy value, where `index` starts at `0` and increases in forward iteration order
    - If no elements pass the filter, the returned ring contains one `nil` element
- `ring.forEach(callback)`, which calls `callback(element, index)` for each element, starting with the current element and moving forward, where `index` starts at `0` and increases by `1` for each element; returns `nil`
- `ring.forEachPrev(callback)`, which calls `callback(element, index)` for each element, starting with the element before the current element and moving backward through the current element, where `index` starts at `0` and increases by `1` for each element; returns `nil`
- `ring.forEachPrevCallLen(callback)`, which calls `callback(element, index)` for each element, starting with the element before the current element and moving backward through the current element, where `index` starts at `len(ring) - 1` and decreases by `1` for each element; returns `nil`
- `ring.forEachReversed(callback)`, which calls `callback(element, index)` for each element, starting with the current element and moving backward, where `index` starts at `0` and increases by `1` for each element; returns `nil`
- `ring.forEachReversedCallLen(callback)`, which calls `callback(element, index)` for each element, starting with the current element and moving backward, where `index` starts at `len(ring) - 1` and decreases by `1` for each element; returns `nil`
- `ring.forEachReversedCallLen0(callback)`, which calls `callback(element, index)` for each element, starting with the current element and moving backward; the index for the current element is `0`, and indexes for subsequent elements start at `len(ring) - 1` and decrease by `1`; returns `nil`
- `ring.getValue(offset)`, which returns the value at the specified integer `offset` from the current element, where positive offsets move forward and negative offsets move backward
    - If `offset` is not an integer, a runtime error is thrown
- `ring.infIter()`, which returns an iterator that endlessly produces elements starting at the current element and moving forward
- `ring.infIterPrev()`, which returns an iterator that endlessly produces elements starting at the element before the current element and moving backward
- `ring.infIterReversed()`, which returns an iterator that endlessly produces elements starting at the current element and moving backward
- `ring.isLinked(otherRing)`, which returns `true` if `otherRing` is the current ring element or can be reached by repeatedly following `ring.next`, and `false` otherwise
    - `otherRing` must be a ring, or else a runtime error is thrown
- `ring.limitIter(length)`, which returns an iterator that produces at most `length` elements starting at the current element and moving forward; the iterator wraps around the ring if necessary
    - `length` must be an integer, or else a runtime error is thrown
    - If `length` is less than or equal to `0`, the iterator is empty
- `ring.limitIterNoWrap(length)`, which returns an iterator that produces at most `length` elements starting at the current element and moving forward, stopping when it reaches the current element again instead of wrapping around
    - `length` must be an integer, or else a runtime error is thrown
    - If `length` is less than or equal to `0`, the iterator is empty
- `ring.link(otherRing)`, which links `otherRing` into the current ring after the current element and returns the element that was immediately after the current element before linking
    - `otherRing` must be a ring, or else a runtime error is thrown
- `ring.loopIter(count)`, which returns an iterator that produces the ring's elements starting at the current element and moving forward for `count` complete cycles
    - `count` must be an integer, or else a runtime error is thrown
    - If `count` is less than or equal to `0`, the iterator is empty
- `ring.map(callback)`, which returns a new ring containing the results of calling `callback(element, index)` for each element, starting at the current element and moving forward, where `index` starts at `0` and increases by `1` for each element
- `ring.mapInPlace(callback)`, which replaces each element with the result of calling `callback(element, index)`, starting at the current element and moving forward, where `index` starts at `0` and increases by `1` for each element; returns `nil`
- `ring.move(offset)`, which returns the ring element at the specified integer `offset` from the current element, where positive offsets move forward and negative offsets move backward
    - If `offset` is not an integer, a runtime error is thrown
- `ring.prevIter()`, which returns an iterator that starts at the element before the current element and moves backward for one full cycle, ending with the current element
- `ring.printLines()`, which prints each element on its own line and returns `nil`
- `ring.printLinesSep(separator)`, which prints each element on its own line with `separator` printed on its own line between elements, and returns `nil`
    - `separator` must be a string, or else a runtime error is thrown
    - If `separator` is an empty string, no separator lines are printed
- `ring.printList()`, which prints a string representation of the ring in the form `Ring [element1, element2, ..., elementN]` and returns `nil`
- `ring.printListToStr()`, which returns a string representation of the ring in the form `Ring [element1, element2, ..., elementN]`
- `ring.reduce(callback, [initialValue])`, which applies `callback(accumulator, element, index)` to the elements from the current element moving forward and returns the final accumulator
    - If `initialValue` is omitted, the current element's value is used as the initial accumulator and the callback starts with the next element
    - `index` starts at `0` for the current element and increases in forward iteration order
    - Exactly `1` or `2` arguments must be provided, and the first argument must be a function, or else a runtime error is thrown
- `ring.reduceRight(callback, [initialValue])`, which applies `callback(accumulator, element, index)` to the elements starting at the element before the current element and moving backward through the current element, and returns the final accumulator
    - If `initialValue` is omitted, the element before the current element is used as the initial accumulator and the callback starts with the next element in reverse iteration order
    - `index` starts at `0` for the first element in reverse iteration order and increases by `1` for each element
    - Exactly `1` or `2` arguments must be provided, and the first argument must be a function, or else a runtime error is thrown
- `ring.reduceRightCallLen(callback, [initialValue])`, which applies `callback(accumulator, element, index)` to the elements starting at the element before the current element and moving backward through the current element, and returns the final accumulator
    - If `initialValue` is omitted, the element before the current element is used as the initial accumulator and the callback starts with the next element in reverse iteration order
    - `index` starts at `len(ring) - 1` for the element before the current element and decreases by `1` for each element
    - Exactly `1` or `2` arguments must be provided, and the first argument must be a function, or else a runtime error is thrown
- `ring.reduceRightFromCurrent(callback, [initialValue])`, which applies `callback(accumulator, element, index)` to the elements starting at the current element and moving backward, and returns the final accumulator
    - If `initialValue` is omitted, the current element's value is used as the initial accumulator and the callback starts with the element before the current element
    - `index` starts at `0` for the current element and increases by `1` in reverse iteration order
    - Exactly `1` or `2` arguments must be provided, and the first argument must be a function, or else a runtime error is thrown
- `ring.reduceRightFromCurrentCallLen(callback, [initialValue])`, which applies `callback(accumulator, element, index)` to the elements starting at the current element and moving backward, and returns the final accumulator
    - If `initialValue` is omitted, the current element's value is used as the initial accumulator and the callback starts with the element before the current element
    - `index` starts at `len(ring) - 1` for the current element and decreases by `1` for each element
    - Exactly `1` or `2` arguments must be provided, and the first argument must be a function, or else a runtime error is thrown
- `ring.setValue(value)`, which sets the value of the current element to `value` and returns `nil`
- `ring.setValueRet(value)`, which sets the value of the current element to `value` and returns the current ring element
- `ring.toList()`, which returns a list of the ring's values starting at the current element and moving forward
- `ring.unlink(count)`, which unlinks the elements between the current element and the element `count + 1` positions forward, and returns the element that was immediately after the current element before unlinking
    - `count` must be an integer, or else a runtime error is thrown
    - If `count` is less than or equal to `0`, the ring is unchanged and the current ring element is returned
    - Forward movement wraps around the ring; if `count + 1` positions forward lands on the current element, the ring is unchanged and the current ring element is returned
