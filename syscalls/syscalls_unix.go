//go:build !windows

package syscalls

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/AlanLuu/lox/loxerror"
	"golang.org/x/sys/unix"
)

func execCommandNotFound(funcName string, path string) error {
	return loxerror.Error(fmt.Sprintf("os.%v: %v: command not found", funcName, path))
}

func forkExecOptions(env []string) *syscall.ProcAttr {
	return &syscall.ProcAttr{
		Files: []uintptr{0, 1, 2}, //stdin, stdout, stderr
		Env:   env,
	}
}

func getWaitStatus(waitStatus unix.WaitStatus) WaitStatus {
	return WaitStatus{
		Continued:  waitStatus.Continued,
		ExitStatus: waitStatus.ExitStatus,
		Exited:     waitStatus.Exited,
		Signaled:   waitStatus.Signaled,
		StopSignal: waitStatus.StopSignal,
		Stopped:    waitStatus.Stopped,
		WaitStatus: int64(waitStatus),
	}
}

func Close(fd int) error {
	return syscall.Close(fd)
}

func Chroot(path string) error {
	return syscall.Chroot(path)
}

func Dup(oldfd int) (int, error) {
	return unix.Dup(oldfd)
}

func Dup2(oldfd int, newfd int) error {
	//if oldfd == newfd and oldfd is a valid file descriptor, do nothing
	if oldfd == newfd {
		err := unix.FcntlFlock(uintptr(oldfd), unix.F_GETFD, nil)
		if err == nil {
			return nil
		}
		if errno, ok := err.(syscall.Errno); ok {
			if errno != unix.EBADF {
				return nil
			}
		}
	}
	return unix.Dup2(oldfd, newfd)
}

func Execv(path string, argv []string) error {
	return syscall.Exec(path, argv, os.Environ())
}

func Execve(path string, argv []string, envp []string) error {
	return syscall.Exec(path, argv, envp)
}

func Execvp(file string, argv []string) error {
	fullPath, err := exec.LookPath(file)
	if err != nil {
		return execCommandNotFound("execvp", file)
	}
	return syscall.Exec(fullPath, argv, os.Environ())
}

func Execvpe(file string, argv []string, envp []string) error {
	fullPath, err := exec.LookPath(file)
	if err != nil {
		return execCommandNotFound("execvpe", file)
	}
	return syscall.Exec(fullPath, argv, envp)
}

func Fchdir(fd int) error {
	return unix.Fchdir(fd)
}

func Fchmod(fd int, mode uint32) error {
	return unix.Fchmod(fd, mode)
}

func Fchown(fd int, uid int, gid int) error {
	return unix.Fchown(fd, uid, gid)
}

func Fork() (int, error) {
	pid, _, errno := unix.Syscall(unix.SYS_FORK, 0, 0, 0)
	var err error
	if errno != 0 {
		err = errno
	}
	return int(pid), err
}

func ForkExec(argv0 string, argv []string, attr *syscall.ProcAttr) (int, error) {
	return syscall.ForkExec(argv0, argv, attr)
}

func ForkExecvp(argv0 string, argv []string, attr *syscall.ProcAttr) (int, error) {
	fullargv0, err := exec.LookPath(argv0)
	if err != nil {
		return -1, execCommandNotFound("forkExecvp", argv0)
	}
	return syscall.ForkExec(fullargv0, argv, attr)
}

func ForkExecFd(argv0 string, argv []string) (int, error) {
	return ForkExec(argv0, argv, forkExecOptions(nil))
}

func ForkExecvpFd(argv0 string, argv []string) (int, error) {
	return ForkExecvp(argv0, argv, forkExecOptions(nil))
}

func ForkExecvpeFd(argv0 string, argv []string, env []string) (int, error) {
	return ForkExecvp(argv0, argv, forkExecOptions(env))
}

func Fsync(fd int) error {
	return unix.Fsync(fd)
}

func Ftruncate(fd int, length int64) error {
	return unix.Ftruncate(fd, length)
}

func Getgroups() ([]int, error) {
	return unix.Getgroups()
}

func Getsid(pid int) (int, error) {
	return unix.Getsid(pid)
}

func Mkfifo(path string, mode uint32) error {
	return unix.Mkfifo(path, mode)
}

func Read(fd int, p []byte) (int, error) {
	return syscall.Read(fd, p)
}

func Setegid(egid int) error {
	return syscall.Setegid(egid)
}

func Seteuid(euid int) error {
	return syscall.Seteuid(euid)
}

func Setgid(gid int) error {
	return syscall.Setgid(gid)
}

func Setgroups(gids []int) error {
	return unix.Setgroups(gids)
}

func Setregid(rgid int, egid int) error {
	return syscall.Setregid(rgid, egid)
}

func Setreuid(ruid int, euid int) error {
	return syscall.Setreuid(ruid, euid)
}

func Setsid() (int, error) {
	return unix.Setsid()
}

func Setuid(uid int) error {
	return unix.Setuid(uid)
}

func Sync() {
	unix.Sync()
}

func Umask(mask int) int {
	return unix.Umask(mask)
}

type UnameResult struct {
	Sysname  string
	Nodename string
	Release  string
	Version  string
	Machine  string
}

func Uname() (UnameResult, error) {
	buf := unix.Utsname{}
	err := unix.Uname(&buf)
	if err != nil {
		return UnameResult{}, err
	}
	toString := func(arr []byte) string {
		index := bytes.IndexByte(arr, 0)
		if index == 0 {
			return ""
		} else if index > 0 {
			return string(arr[:index])
		}
		return string(arr)
	}
	return UnameResult{
		toString(buf.Sysname[:]),
		toString(buf.Nodename[:]),
		toString(buf.Release[:]),
		toString(buf.Version[:]),
		toString(buf.Machine[:]),
	}, nil
}

func Wait() (int, WaitStatus, error) {
	var waitStatus unix.WaitStatus
	pid, err := unix.Wait4(-1, &waitStatus, 0, nil)
	if err != nil {
		return -1, WaitStatus{}, err
	}
	return pid, getWaitStatus(waitStatus), nil
}

func Write(fd int, p []byte) (int, error) {
	return syscall.Write(fd, p)
}
