package syscalls

import "syscall"

type WaitStatus struct {
	Continued  func() bool
	ExitStatus func() int
	Exited     func() bool
	Signaled   func() bool
	StopSignal func() syscall.Signal
	Stopped    func() bool
	WaitStatus int64
}
