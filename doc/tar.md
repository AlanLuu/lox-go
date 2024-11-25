# Tar methods and fields

The following fields are defined in the built-in `tar` class, which are all integers:
- `tar.USE_BUFFER`

The following methods are defined in the built-in `tar` class:
- `tar.writer(file/tar.USE_BUFFER)`, which returns a tar writer object that writes to the specified file object. If `tar.USE_BUFFER` is specified instead of a file object, the returned tar writer object writes to an internal buffer instead

Tar writer objects have the following methods associated with them:
- `tar writer.addFile(file)`, which creates a file with its name and content being the name and content of the specified file object and returns the current tar writer object itself
    - If a file or directory with the same file name already exists, a runtime error is thrown
    - This method throws a runtime error if the tar writer object is closed
    - This method preserves the Unix permissions of the original file on Unix systems
- `tar writer.addFile(path, content)`, which creates a file with the specified path name string in the current tar object with its content being a buffer, file object, or string and returns the current tar writer object itself
    - If a file or directory with the specified path name already exists, a runtime error is thrown
    - This method throws a runtime error if the tar writer object is closed
    - This method utilizes the umask value for file permissions on Unix systems
- `tar writer.buffer()`, which returns a buffer of the raw bytes of the final tar file
    - The current tar writer object must have been created using `tar.writer(tar.USE_BUFFER)` or else this method throws a runtime error
    - The current tar writer object must be flushed or closed before calling this method
- `tar writer.close()`, which closes the current tar writer object and writes the tar file bytes to the specified file or buffer, depending on how the tar writer object was created
- `tar writer.fileNames()`, which returns a list of file names in the current tar writer object as strings in alphabetical order
- `tar writer.flush()`, which writes the tar file bytes to the specified file or buffer, depending on how the tar writer object was created, without closing the current tar writer object and returns the current tar writer object itself
- `tar writer.isBuffer()`, which returns `true` if the current tar writer object was created using `tar.writer(tar.USE_BUFFER)` and `false` otherwise
- `tar writer.isClosed()`, which returns `true` if the current tar writer object is closed and `false` otherwise
- `tar writer.mkdir(path)`, which creates a directory with the specified path name string in the current tar object and returns the current tar writer object itself
    - If a file or directory with the specified path name already exists, a runtime error is thrown
    - This method throws a runtime error if the tar writer object is closed
    - This method utilizes the umask value for directory permissions on Unix systems
- `tar writer.printFileNames()`, which prints the file names in the current tar writer object as strings in alphabetical order
