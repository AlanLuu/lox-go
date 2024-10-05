# Windows methods and fields

Note: this class does not exist on non-Windows systems.

Any method that fails will throw a runtime error with a message describing the error.

The following methods and fields are defined in the built-in `windows` class:
- `windows.computerName()`, which returns a string that is the NetBIOS name of the current system
- `windows.getFileType(handle)`, which returns an integer corresponding to the type of file with the specified file handle integer, which is further documented [here](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getfiletype)
- `windows.getFileTypeStr(handle)`, which returns a string corresponding to the type of file with the specified file handle integer, according to the following mapping:
    ```go
    map[uint32]string{
        0x0002: "CHAR",
        0x0001: "DISK",
        0x0003: "PIPE",
        0x8000: "REMOTE",
        0x0000: "UNKNOWN",
    }
    ```
- `windows.getLastError()`, which returns an error object corresponding to the last error returned by a `windows` class method, or `nil` if there is no such error
- `windows.getLogicalDrives()`, which returns an integer that is a bitmask of the available disk drives on the current system, further documented [here](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getlogicaldrives)
- `windows.getMaximumProcessorCount()`, which returns an integer representing the number of logical processors on the current system
- `windows.getSystemDirectory()`, which returns a string that is the path of the system directory, which is usually `C:\Windows\System32`
- `windows.getSystemWindowsDirectory()`, which returns a string that is the path of the Windows directory, which is usually `C:\Windows`
- `windows.listDrives()`, which returns a list that contains the Windows drive names on the current system as strings, which typically looks like `C:\\`
- `windows.stderr`, which is an integer that refers to the file handle for the standard error stream on Windows
- `windows.stdin`, which is an integer that refers to the file handle for the standard input stream on Windows
- `windows.stdout`, which is an integer that refers to the file handle for the standard output stream on Windows
