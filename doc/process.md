# Process methods

The following methods are defined in the built-in `process` class:
- `process class.new(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments
- `process class.newReusable(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments that can be reused to run the command multiple times
- `process class.newReusableSetStd(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments that can be reused to run the command multiple times with the process' stdin, stdout, and stderr already set
- `process class.newSetStd(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments with the process' stdin, stdout, and stderr already set
- `process class.newShell(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments passed into the system shell
- `process class.newShellReusable(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments passed into the system shell that can be reused to run the command multiple times
- `process class.newShellReusableSetStd(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments passed into the system shell that can be reused to run the command multiple times with the process' stdin, stdout, and stderr already set
- `process class.newShellSetStd(list/args)`, which takes in a list of strings or various string arguments and returns a process object with the specified command and arguments passed into the system shell with the process' stdin, stdout, and stderr already set
- `process class.run(list/args)`, which takes in a list of strings or various string arguments, creates a process object with the specified command and arguments, executes the process and waits for it to complete, and returns a process result object once the process completes successfully
- `process class.runSetStd(list/args)`, which takes in a list of strings or various string arguments, creates a process object with the specified command and arguments with the process' stdin, stdout, and stderr already set, executes the process and waits for it to complete, and returns a process result object once the process completes successfully
- `process class.runShell(list/args)`, which takes in a list of strings or various string arguments, creates a process object with the specified command and arguments passed into the system shell, executes the process and waits for it to complete, and returns a process result object once the process completes successfully
- `process class.runShellSetStd(list/args)`, which takes in a list of strings or various string arguments, creates a process object with the specified command and arguments passed into the system shell with the process' stdin, stdout, and stderr already set, executes the process and waits for it to complete, and returns a process result object once the process completes successfully

Process objects have the following methods associated with them:
- `process.args()`, which returns a list of the command and argument strings associated with this process object
- `process.dir()`, which returns a string of the current working directory of the command associated with this process object
- `process.combinedOutput()`, which executes the process and returns a string of the standard output and standard error contents combined
- `process.combinedOutputBuf()`, which executes the process and returns a buffer of the standard output and standard error contents combined
- `process.isReusable()`, which returns `true` if this process object can be reused to run the associated command multiple times and `false` otherwise
- `process.output()`, which executes the process and returns a string of the standard output contents
- `process.outputBuf()`, which executes the process and returns a buffer of the standard output contents
- `process.path()`, which returns a string of the path of the command associated with this process object
- `process.run()`, which executes the process, waits for it to complete, and returns a process result object once the process completes successfully
- `process.running()`, which returns `true` if the process is currently running, meaning it has been started but hasn't been waited on yet, and `false` otherwise
- `process.setArgs(argsList)`, which takes in a list of strings, sets the internal command and argument strings list of this process object to that list, and returns the process object itself
- `process.setDir(dir)`, which takes in a string, sets the internal current working directory string of this process object to that string, and returns the process object itself
- `process.setPath(path)`, which takes in a string, sets the internal path string of this process object to that string, and returns the process object itself
- `process.setReusable(bool)`, which takes in a boolean, sets the internal property of whether this process object can be reused to run the associated command multiple times to that boolean, and returns the process object itself
- `process.setStderr(file)`, which takes in a file object, sets the internal stderr property of this process object to that file, and returns the process object itself
- `process.setStdin(file)`, which takes in a file object, sets the internal stdin property of this process object to that file, and returns the process object itself
- `process.setStdout(file)`, which takes in a file object, sets the internal stdout property of this process object to that file, and returns the process object itself
- `process.start()`, which executes the process but does not wait for it to complete
- `process.started()`, which returns `true` if the process has been started and `false` otherwise
- `process.wait()`, which waits for the process, which must have been started already, to complete and returns a process result object once the process completes successfully
- `process.waited()`, which returns `true` if the process has been successfully waited on and `false` otherwise

Process result objects have the following methods associated with them:
- `process result.exitCode()`, which returns the exit code of the completed process as an integer
- `process result.exited()`, which returns `true` if the completed process exited and `false` otherwise
- `process result.pid()`, which returns the process ID of the completed process as an integer
- `process result.success()`, which returns `true` if the completed process exited successfully and `false` otherwise
