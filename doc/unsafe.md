# Unsafe methods

This class and its methods are not defined in non-unsafe mode.

The following methods are defined in the built-in `unsafe` class:
- `<unsafe> unsafe.threadFunc(numThreads, callback)`, which takes in an integer `numThreads` and a callback function and spins up `numThreads` threads that execute the callback function concurrently
    - If `numThreads` is negative, it is the same as specifying `0` for that argument
    - The call to `unsafe.threadFunc` blocks until all threads have finishing running
    - If a runtime error is thrown in a thread, the error is printed to standard error but this doesn't affect the remaining threads
- `<unsafe> unsafe.threadFuncs(num, callback1, [callback2, ..., callbackN])`, which takes in an integer `num` and at least one callback function and spins up `num * callbackCount` threads that execute all provided callback functions concurrently, where `callbackCount` is the number of callback functions provided as arguments to `unsafe.threadFuncs`
    - If `num` is negative, it is the same as specifying `0` for that argument
    - The call to `unsafe.threadFuncs` blocks until all threads have finishing running
    - If a runtime error is thrown in a thread, the error is printed to standard error but this doesn't affect the remaining threads
