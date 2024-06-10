## OS methods and fields

Any method that fails will throw a runtime error with a message describing the error.

The following methods and fields are defined in the built-in `os` class:
- `os.chdir(directory)`, which changes the current working directory to the specified directory string
- `os.chmod(path, mode)`, which changes the mode of the specified path string to `mode`
    - This method works on Windows, but only the read-only flag can be changed. Use mode `0400` to make the file read-only and `0600` to make it readable and writable
- `os.executable()`, which returns the absolute path name of the executable for the current process as a string
- `os.exit([code])`, which exits the program with the specified exit code. If `code` is omitted, the default exit code is 0
    - Calling this method will immediately stop the program without running any other code, e.g., if this method is called inside a try-catch block with a `finally` block, the `finally` block will not be executed
- `os.getcwd()`, which returns the current working directory as a string
- `os.getenv(key, [default])`, which returns the value of the specified environment variable `key`, which is a string, as a string. If the value doesn't exist, the value of `default` is returned if specified, otherwise `nil` is returned
- `os.getenvs()`, which returns a dictionary with all environment variable keys as dictionary keys and all environment variable values as dictionary values
- `os.getgid()`, which returns the group ID of the current process as an integer
    - On Windows, this method always returns `-1`
- `os.getpid()`, which returns the process ID of the current process as an integer
- `os.getppid()`, which returns the process ID of the parent process as an integer
- `os.getuid()`, which returns the user ID of the current process as an integer
    - On Windows, this method always returns `-1`
- `os.hostname()`, which returns the hostname of the computer as a string
- `os.listdir([path])`, which returns a list of names of all directories and files in the specified path as strings. If `path` is omitted, the current working directory is used as the path
- `os.mkdir(name)`, which creates a new directory with the specified name in the current working directory
- `os.name`, which is a string that specifies the operating system that the program is running on
- `os.open(name, mode)`, which opens a file specified by a path name with the mode specified by the mode string. This method returns a file object if successful, which itself is documented [here](./doc/file.md)
    - The following file modes are available:
        - `"r"`, which opens a file for reading and throws a runtime error if the file doesn't exist
        - `"w"`, which opens a file for writing, creating the file if it doesn't exist and truncating the file if it already exists
        - `"a"`, which opens a file for writing, creating the file if it doesn't exist and appending to the file if it already exists
        - Along with the above modes, the letter `"b"` can also be specified to open a file in binary mode, such as `"rb"` for reading a binary file
            - The ordering doesn't matter, so the mode `"br"` is the same as `"rb"`
- `os.remove(path)`, which removes the file or empty directory at the specified path string
    - If the directory is not empty, a runtime error is thrown
- `os.removeAll(path)`, which removes the file or directory at the specified path string
    - If the directory is not empty, all files and directories inside it are removed recursively
- `os.setenv(key, value)`, which sets an environment variable with the specified key and value, which are both strings
- `os.system(command)`, which runs the specified command string in the system shell, which is `sh` on Unix and `cmd` on Windows, and returns the exit code of the command as an integer
- `os.touch(name)`, which creates a new empty file with the specified name in the current working directory
    - If the file already exists, it is truncated
- `os.unsetenv(key)`, which unsets the environment variable `key`, which is a string
- `os.username()`, which returns the username of the user running the current process as a string
