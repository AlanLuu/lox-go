## Wait status object methods

Refer to the Unix manual page for `wait(2)` for further information regarding the behavior of these methods.

The following methods are defined in wait status objects:
- `waitstatus.continued()`, which returns `true` if the child process was resumed by a SIGCONT signal and `false` otherwise
- `waitstatus.exited()`, which returns `true` if the child process exited normally and `false` otherwise
- `waitstatus.exitStatus()`, which returns an integer representing the exit status code of the child process
- `waitstatus.signaled()`, which returns `true` if the child process received a signal and terminated as a result and `false` otherwise
- `waitstatus.stopped()`, which returns `true` if the child process received a signal and stopped as a result and `false` otherwise
- `waitstatus.stopSignal()`, which returns an integer representing the specific signal that terminated the child process
- `waitstatus.waitStatus()`, which returns an integer representing the specific wait status itself, basically `wstatus` from the Unix manual page for `wait(2)`
