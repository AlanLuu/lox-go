# Cryptographic methods and fields

The following methods are defined in the built-in `crypto` class:
- `crypto.md5([data])`, which returns a hash object that computes the MD5 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
    - Warning: MD5 is cryptographically broken and is unsuitable for security purposes
- `crypto.sha1([data])`, which returns a hash object that computes the SHA-1 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
    - Warning: SHA-1 is cryptographically broken and is unsuitable for security purposes
- `crypto.sha256([data])`, which returns a hash object that computes the SHA-256 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
- `crypto.sha512([data])`, which returns a hash object that computes the SHA-512 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it

Hash objects have the following methods and fields associated with them:
- `hash.blockSize`, which is the block size of the hash object's hash algorithm as an integer
- `hash.digest()`, which returns a buffer of the current hash based on the hash object's hash algorithm and the current data in the hash object
- `hash.hex()`, which is an alias for `hash.hexDigest`
- `hash.hexDigest()`, which returns a string that is the hexadecimal representation of the current hash based on the hash object's hash algorithm and the current data in the hash object
- `hash.reset()`, which clears all the current data from the hash object, resetting it to its initial state
- `hash.size`, which is the number of bytes the final hash will have as an integer
- `hash.type`, which is the type of the hash object's hash algorithm as a string, with the following values:
    - `md5` for MD5
    - `sha1` for SHA-1
    - `sha256` for SHA-256
    - `sha512` for SHA-512
- `hash.update(data)`, which updates the hash object with the specified data, which must be a buffer or string
