# Stdin methods

The following methods are defined in the `stdin` class:
- `stdin.diff()`, which reads values from standard input separated by a newline character and prints the difference of the numerical values read, ignoring any non-numerical values
- `stdin.diffwin()`, which is like `stdin.diff` except that the separator is CRLF instead of LF
- `stdin.fernetdecrypt(key)`, which reads in data encrypted using fernet as a fernet token from standard input and attempts to decrypt the data using the specified fernet key. If successful, this method writes the decrypted data to standard output, otherwise this method throws a runtime error
- `stdin.fernetencrypt()`, which reads in bytes from standard input and encrypts them using fernet, writing the generated fernet key in base64 format to standard error and the encrypted data as a fernet token to standard output
- `stdin.filter(callback)`, which reads values from standard input separated by a newline character and prints the values where the specified callback function returns a truthy value for them
- `stdin.filterwin(callback)`, which is like `stdin.filter` except that the separator is CRLF instead of LF
- `stdin.lower()`, which prints the lowercase version of the contents from standard input
- `stdin.map(callback)`, which reads values from standard input separated by a newline character and prints the result of calling the specified callback function on each value
- `stdin.mapwin(callback)`, which is like `stdin.map` except that the separator is CRLF instead of LF
- `stdin.max()`, which reads values from standard input separated by a newline character and prints the maximum numerical value out of all the numerical values read, ignoring any non-numerical values
- `stdin.maxwin()`, which is like `stdin.max` except that the separator is CRLF instead of LF
- `stdin.min()`, which reads values from standard input separated by a newline character and prints the minimum numerical value out of all the numerical values read, ignoring any non-numerical values
- `stdin.minwin()`, which is like `stdin.min` except that the separator is CRLF instead of LF
- `stdin.rot13()`, which prints the ROT13 encoding of the contents from standard input
- `stdin.rot18()`, which prints the ROT18 encoding of the contents from standard input
- `stdin.rot47()`, which prints the ROT47 encoding of the contents from standard input
- `stdin.sum()`, which reads values from standard input separated by a newline character and prints the sum of the numerical values read, ignoring any non-numerical values
- `stdin.sumwin()`, which is like `stdin.sum` except that the separator is CRLF instead of LF
- `stdin.upper()`, which prints the uppercase version of the contents from standard input
