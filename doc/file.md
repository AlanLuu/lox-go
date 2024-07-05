## File methods and fields

The following methods and fields are defined on file instances:
- `file.chdir()`, which changes the current working directory to the current file, which must be a directory
- `file.chmod(mode)`, which changes the mode of the file to `mode`
    - This method works on Windows, but only the read-only flag can be changed. Use mode `0400` to make the file read-only and `0600` to make it readable and writable
- `file.close()`, which flushes and closes the file, causing any attempts to read or write from the file to throw a runtime error
    - If the file is already closed, this method can still be called, with no additional effects occurring when doing so
- `file.closed`, which is a boolean that is `true` if the file is closed and `false` otherwise
- `file.flush()`, which flushes the file by writing all buffered output to disk
- `file.isClosed()`, which is a function that returns the value of `file.closed`
- `file.mode`, which is a string representing the mode of the file
- `file.name`, which is a string representing the name of the file
- `file.read([numBytes])`, which reads the specified number of bytes from the file
    - If the number of bytes is omitted or negative, the entire file is read until EOF
    - This method returns a buffer instance with integers representing the raw bytes of the file if the file was opened in binary mode, otherwise it returns a string representing the file contents
        - If the file pointer is at EOF, this method returns an empty buffer instance if the file was opened in binary mode, otherwise it returns an empty string
    - If the file is closed or is not open in read mode, this method throws a runtime error
- `file.readByte()`, which reads a single byte from the file and returns it as an integer
    - If the file pointer is at EOF, this method returns `nil`
    - The file must not be closed and must be in binary read mode, otherwise a runtime error is thrown when this method is called
- `file.readChar()`, which reads a single character from the file and returns it as a string with that character
    - If the file pointer is at EOF, this method returns `nil`
    - The file must not be closed and must be in read mode (not binary read mode), otherwise a runtime error is thrown when this method is called
- `file.readLine()`, which reads a single line from the file, skipping over any blank lines, and returns that line without any trailing newline characters as a string
    - If the file pointer is at EOF, this method returns an empty string
    - The file must not be closed and must be in read mode (not binary read mode), otherwise a runtime error is thrown when this method is called
- `file.readLines([numLines])`, which reads in `numLines` lines from the file, skipping over any blank lines, and returns a list of those lines without any trailing newline characters, where `numLines` is an integer. If `numLines` is omitted or negative, all lines from the file are read into the list
    - If the file pointer is at EOF, this method returns an empty list
    - The file must not be closed and must be in read mode (not binary read mode), otherwise a runtime error is thrown when this method is called
- `file.readNewLine()`, which reads and returns a single line from the file as a string with any trailing newline characters if they exist
    - If the file pointer is at EOF, this method returns an empty string
    - The file must not be closed and must be in read mode (not binary read mode), otherwise a runtime error is thrown when this method is called
- `file.readNewLines([numLines])`, which reads in `numLines` lines from the file and returns a list of those lines, where `numLines` is an integer. If `numLines` is omitted or negative, all lines from the file are read into the list
    - If the file pointer is at EOF, this method returns an empty list
    - The file must not be closed and must be in read mode (not binary read mode), otherwise a runtime error is thrown when this method is called
- `file.seek(offset, whence)`, which sets the position of where to start reading/writing from/to the file to `offset` according to `whence`, which are both integers. Valid integer values for `whence` are the following:
    - `os.SEEK_SET` or `0` – relative to the start of the file
    - `os.SEEK_CUR` or `1` – relative to the file's current offset position
    - `os.SEEK_END` or `2` – relative to the end of the file
    - A runtime error is thrown if `whence` is an invalid value or the file is in append mode when calling this method
- `file.write(buffer/string)`, which writes the contents of the specified buffer or string to the file and returns the number of bytes written as an integer
    - In binary mode, this method's argument must be a buffer, otherwise the argument must be a string
    - If the file is closed or is not open in write or append mode, this method throws a runtime error
- `file.writeByte(byte)`, which writes the specified byte to the file, where `byte` is an integer ranging from 0 to 255
    - If the file is closed, is not open in write or append mode, or is not open in binary mode, this method throws a runtime error
- `file.writeLine(string)`, which writes the contents of the specified string to the file followed by a newline character and returns the number of bytes written as an integer
    - On Windows, CRLF is used as the newline character
    - If the file is closed, is not open in write or append mode, or is open in binary mode, this method throws a runtime error
- `file.writeLines(lines)`, which writes the elements of `lines` to the file, where `lines` is a list of strings, and returns the total number of bytes written as an integer
    - If an element in `lines` is not a string, the string representation of that element is used as the string to write to the file
    - If the file is closed, is not open in write or append mode, or is open in binary mode, this method throws a runtime error
- `file.writeNewLines(lines)`, which writes the elements of `lines` to the file with a newline character following each element, where `lines` is a list of strings, and returns the total number of bytes written as an integer
    - On Windows, CRLF is used as the newline character
    - If an element in `lines` is not a string, the string representation of that element is used as the string to write to the file
    - If the file is closed, is not open in write or append mode, or is open in binary mode, this method throws a runtime error