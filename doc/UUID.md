# UUID methods and fields

The following special UUID fields are defined in the built-in `UUID` class, which are all UUID instances:
- `UUID.dns`, `UUID.url`, `UUID.oid`, `UUID.null`, `UUID.max`

The following UUID variant fields are defined in the built-in `UUID` class, which are all integers:
- `UUID.invalid = 0`, `UUID.rfc4122 = 1`, `UUID.reserved = 2`, `UUID.microsoft = 3`, `UUID.future = 4`

Any method that fails will throw a runtime error with a message describing the error.

The following methods are defined in the built-in `UUID` class:
- `UUID.disableRandPool()`, which disables the internal random pool for generating UUIDs
- `UUID.enableRandPool()`, which enables the internal random pool for generating UUIDs, which may speed up UUID generation but also be bad for security purposes due to the pool being stored on the heap
- `UUID.fromBytes(buffer)`, which returns a new UUID object from the specified buffer representation of a UUID
- `UUID.mustValidate(string)`, which throws a runtime error if the specified string is not a valid UUID
- `UUID.new()`, which is an alias for `UUID.newV4`
- `UUID.newV1()`, which returns a UUID object with a random version 1 UUID
- `UUID.newV4()`, which returns a UUID object with a random version 4 UUID
- `UUID.newV6()`, which returns a UUID object with a random version 6 UUID
- `UUID.newV7()`, which returns a UUID object with a random version 7 UUID
- `UUID.parse(string)`, which returns a UUID object with the UUID decoded from the specified string
    - This method throws a runtime error if the specified string is not a valid UUID
- `UUID.parseBytes(buffer)`, which returns a UUID object with the UUID decoded from the specified buffer
    - This method throws a runtime error if the specified buffer is not a valid UUID
- `UUID.validate(string)`, which returns `true` if the specified string is a valid UUID and `false` otherwise

UUID objects have the following methods associated with them:
- `uuid.bytes()`, which returns a buffer of the bytes of the UUID associated with the current UUID object
- `uuid.clockSequence()`, which returns the clock sequence that is encoded in the UUID associated with the current UUID object as an integer
    - The UUID associated with the current UUID object must be a version 1 or 2 UUID or else a runtime error is thrown
- `uuid.string()`, which returns a string of the UUID associated with the current UUID object
- `uuid.time()`, which returns an integer that represents the 100s of nanoseconds since October 15, 1582 encoded in the UUID associated with the current UUID object as an integer
    - The UUID associated with the current UUID object must be a version 1, 2, 6, or 7 UUID or else a runtime error is thrown
- `uuid.urn()`, which returns a string that represents the URN form of the UUID associated with the current UUID object
- `uuid.variant()`, which returns an integer that represents the variant of the UUID associated with the current UUID object
- `uuid.variantStr()`, which is an alias for `uuid.variantString`
- `uuid.variantString()`, which returns an string that represents the variant of the UUID associated with the current UUID object
- `uuid.version()`, which returns the version number of the UUID associated with the current UUID object as an integer
- `uuid.versionStr()`, which is an alias for `uuid.versionString`
- `uuid.versionString()`, which returns the version of the UUID associated with the current UUID object as a string
