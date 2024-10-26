# Cryptographic methods and fields

Any method that fails will throw a runtime error with a message describing the error.

The following methods are defined in the built-in `crypto` class:
- `crypto.bcrypt(password, [cost])`, which returns a string of the bcrypt hash of the specified password as a string with the specified cost as an integer. If the cost is omitted, the resulting bcrypt hash will have a cost of 10
    - If the cost is less than 4, the cost is set to 10
    - If the cost is greater than 31, a runtime error is thrown
    - If the password is larger than 72 bytes, a runtime error is thrown
- `crypto.bcryptVerify(password, hash)`, which takes in the specified password and bcrypt hash as strings and returns `true` if the password matches the hash and `false` otherwise
- `crypto.ed25519()`, which returns an Ed25519 keypair object with a random private key and the public key corresponding to that private key
- `crypto.ed25519priv(privKey)`, which takes in an Ed25519 private key as a buffer or base64 string and returns an Ed25519 keypair object with the specified private key and the public key corresponding to that private key
- `crypto.ed25519pub(pubKey)` which takes in an Ed25519 public key as a buffer or base64 string and returns an Ed25519 public key object with the specified public key
- `crypto.ed25519seed(seed)`, which takes in an Ed25519 private key seed as a buffer or base64 string and returns an Ed25519 keypair object with the private key that is generated from the specified private key seed and the public key corresponding to that private key
- `crypto.fernet([key])`, which returns a new fernet object from the specified key argument, which must be a buffer of length 32 or a string in base64 or hexadecimal format. If `key` is omitted, a random key is generated and used as the key in the returned fernet object
    - This method throws a runtime error if the specified key is a buffer and its length is not 32
    - This method throws a runtime error if the specified key is a string and is in an invalid format
- `crypto.hmac(function, key)`, which takes in a function that returns a hash object and a buffer or string as the key and returns a hash object that computes the HMAC of data that is passed into it
    - Example: `crypto.hmac(crypto.sha256, os.urandom(32))` returns an HMAC-SHA256 hash object with a random 32-byte key as the HMAC key
    - The specified function argument must return a hash object or else a runtime error is thrown
- `crypto.md5([data])`, which returns a hash object that computes the MD5 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
    - Warning: MD5 is cryptographically broken and is unsuitable for security purposes
- `crypto.md5sum(data)`, which returns a string that is the hexadecimal representation of the MD5 hash of the specified data, which is either a buffer or string
    - Warning: MD5 is cryptographically broken and is unsuitable for security purposes
- `crypto.sha1([data])`, which returns a hash object that computes the SHA-1 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
    - Warning: SHA-1 is cryptographically broken and is unsuitable for security purposes
- `crypto.sha1sum(data)`, which returns a string that is the hexadecimal representation of the SHA-1 hash of the specified data, which is either a buffer or string
    - Warning: SHA-1 is cryptographically broken and is unsuitable for security purposes
- `crypto.sha224([data])`, which returns a hash object that computes the SHA-224 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
- `crypto.sha224sum(data)`, which returns a string that is the hexadecimal representation of the SHA-224 hash of the specified data, which is either a buffer or string
- `crypto.sha256([data])`, which returns a hash object that computes the SHA-256 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
- `crypto.sha256sum(data)`, which returns a string that is the hexadecimal representation of the SHA-256 hash of the specified data, which is either a buffer or string
- `crypto.sha384([data])`, which returns a hash object that computes the SHA-384 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
- `crypto.sha384sum(data)`, which returns a string that is the hexadecimal representation of the SHA-384 hash of the specified data, which is either a buffer or string
- `crypto.sha512([data])`, which returns a hash object that computes the SHA-512 hash of data that is passed into it. If the `data` parameter is specified, which must be a buffer or string, the hash object is initialized with the specified data passed into it
- `crypto.sha512sum(data)`, which returns a string that is the hexadecimal representation of the SHA-512 hash of the specified data, which is either a buffer or string

Ed25519 keypairs and public key objects have the following methods associated with them:
- `ed25519.isKeyPair()`, which returns `true` if the specified Ed25519 object is a keypair and `false` otherwise
- `ed25519.pubKey()`, which returns a buffer of the public key contents associated with the current Ed25519 object
- `ed25519.pubKeyEquals(arg)`, which takes in another Ed25519 keypair or public key object as an argument and returns `true` if both public keys associated with the two Ed25519 objects are the same and `false` otherwise
- `ed25519.pubKeyStr()`, which returns an encoded base64 string of the public key contents associated with the current Ed25519 object
- `ed25519.privKey()`, which returns a buffer of the private key contents associated with the current Ed25519 object
    - This method throws a runtime error if the current Ed25519 object is not a keypair
- `ed25519.privKeyEquals(arg)`, which takes in another Ed25519 keypair as an argument and returns `true` if both private keys associated with the two Ed25519 keypairs are the same and `false` otherwise
    - This method throws a runtime error if the current Ed25519 object or the specified Ed25519 object is not a keypair
- `ed25519.privKeyStr()`, which returns an encoded base64 string of the private key contents associated with the current Ed25519 object
    - This method throws a runtime error if the current Ed25519 object is not a keypair
- `ed25519.seed()`, which returns a buffer of the private key seed contents associated with the current Ed25519 object
    - This method throws a runtime error if the current Ed25519 object is not a keypair
- `ed25519.seedB64()`, which is an alias for `ed25519.seedBase64`
- `ed25519.seedBase64()`, which returns a base64 string of the private key seed contents associated with the current Ed25519 object
    - This method throws a runtime error if the current Ed25519 object is not a keypair
- `ed25519.seedEncoded()`, which returns an encoded string of the private key seed contents associated with the current Ed25519 object, where the encoding used is base64
    - This method throws a runtime error if the current Ed25519 object is not a keypair
- `ed25519.seedStr()`, which is an alias for `ed25519.seedEncoded`
- `ed25519.sign(message)`, which takes in a message as a buffer or string and returns a buffer of the signature generated by signing the specified message with the private key associated with the current Ed25519 object
    - This method throws a runtime error if the current Ed25519 object is not a keypair
- `ed25519.signToStr(message)`, which takes in a message as a buffer or string and returns an encoded string of the signature generated by signing the specified message with the private key associated with the current Ed25519 object, where the encoding used is base64
    - This method throws a runtime error if the current Ed25519 object is not a keypair
- `ed25519.verify(message, signature)`, which takes in a message and a signature, which can both be buffers or base64 strings, and returns `true` if the specified signature is a valid signature of the specified message using the public key associated with the current Ed25519 object and `false` otherwise

Fernet objects have the following methods associated with them:
- `fernet.b64()`, which is an alias for `fernet.base64`
- `fernet.base64()`, which returns a string that is the base64 representation of the fernet key associated with the current fernet object
- `fernet.bytes()`, which returns a buffer of the bytes of the fernet key associated with the current fernet object
- `fernet.decrypt(buffer/file/string)`, which attempts to decrypt the specified buffer, file object, or string representation of the specified fernet token using the key associated with the current fernet object, and returns a buffer of the decrypted bytes if successful
    - If decryption is unsuccessful, this method throws a runtime error
- `fernet.decryptToFile(buffer/file/string, string/file)`, which attempts to decrypt the specified buffer, file object, or string representation of the specified fernet token using the key associated with the current fernet object and writes the decrypted bytes to the specified file, which can be specified as a string or a file object
    - If the destination file is specified as a string, it is created if it doesn't already exist and truncated if it already exists
    - If decryption is unsuccessful, this method throws a runtime error
- `fernet.decryptToStr(buffer/string)`, which attempts to decrypt the specified buffer, file object, or string representation of the specified fernet token using the key associated with the current fernet object, and returns a string of the decrypted bytes if successful
    - If decryption is unsuccessful, this method throws a runtime error
- `fernet.encrypt(buffer/file/string)`, which returns a buffer of the encryption result of the specified buffer, file object, or string, which is known as a fernet token
- `fernet.encryptToFile(buffer/file/string, string/file)`, which writes the encryption result of the specified buffer, file object, or string, which is known as a fernet token, to the specified file, which can be specified as a string or a file object
    - If the destination file is specified as a string, it is created if it doesn't already exist and truncated if it already exists
- `fernet.encryptToStr(buffer/file/string)`, which returns a string of the encryption result of the specified buffer, file object, or string, which is known as a fernet token
- `fernet.hex()`, which returns a string that is the hexadecimal representation of the fernet key associated with the current fernet object

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
    - `sha224` for SHA-224
    - `sha256` for SHA-256
    - `sha384` for SHA-384
    - `sha512` for SHA-512
- `hash.update(data)`, which updates the hash object with the specified data, which must be a buffer or string
