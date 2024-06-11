## File methods and fields

The following methods and fields are defined on file instances:
- `file.close()`, which flushes and closes the file, causing any attempts to read or write from the file to throw a runtime error
    - If the file is already closed, this method can still be called, with no additional effects occurring when doing so
- `file.closed`, which is a boolean that is `true` if the file is closed and `false` otherwise
- `file.flush()`, which flushes the file by writing all buffered output to disk
- `file.isClosed()`, which is a function that returns the value of `file.closed`
- `file.name`, which is a string representing the name of the file
- `file.read([numBytes])`, which reads the specified number of bytes from the file
    - If the number of bytes is omitted or negative, the entire file is read until EOF
    - This method returns a buffer instance with integers representing the raw bytes of the file if the file was opened in binary mode, otherwise it returns a string representing the file contents
    - If the file is closed or is not open in read mode, this method throws a runtime error
- `file.seek(offset, whence)`, which sets the position of where to start reading/writing from/to the file to `offset` according to `whence`, which are both integers. Valid integer values for `whence` are the following:
    - `os.SEEK_SET` or `0` – relative to the start of the file
    - `os.SEEK_CUR` or `1` – relative to the file's current offset position
    - `os.SEEK_END` or `2` – relative to the end of the file
    - A runtime error is thrown if `whence` is an invalid value or the file is in append mode when calling this method
- `file.write(buffer/string)`, which writes the contents of the specified buffer or string to the file and returns the number of bytes written as an integer
    - In binary mode, this method's argument must be a buffer, otherwise the argument must be a string
    - If the file is closed or is not open in write mode, this method throws a runtime error
- `file.writeLine(string)`, which writes the contents of the specified string to the file followed by a newline character and returns the number of bytes written as an integer
    - On Windows, CRLF is used as the newline character
    - If the file is closed, is not open in write mode, or is open in binary mode, this method throws a runtime error
