# Secrets methods

The following methods are defined in the `secrets` class:
- `secrets.base32(numBytes)`, which returns a string that is the base32 representation of `numBytes` random bytes that are generated in a cryptographically secure manner
    - If `numBytes` is negative, a runtime error is thrown
- `secrets.base32def()`, which returns the value of calling `secrets.base32` with an argument of `32`
- `secrets.base64(numBytes)`, which returns a string that is the base64 representation of `numBytes` random bytes that are generated in a cryptographically secure manner
    - If `numBytes` is negative, a runtime error is thrown
- `secrets.base64def()`, which returns the value of calling `secrets.base64` with an argument of `32`
- `secrets.hex(numBytes)`, which returns a string that is the hexadecimal representation of `numBytes` random bytes that are generated in a cryptographically secure manner
    - If `numBytes` is negative, a runtime error is thrown
- `secrets.hexdef()`, which returns the value of calling `secrets.hex` with an argument of `32`
- `secrets.urlsafe(numBytes)`, which returns a string that is the URL-safe base64 representation of `numBytes` random bytes that are generated in a cryptographically secure manner
    - If `numBytes` is negative, a runtime error is thrown
- `secrets.urlsafedef()`, which returns the value of calling `secrets.urlsafe` with an argument of `32`
