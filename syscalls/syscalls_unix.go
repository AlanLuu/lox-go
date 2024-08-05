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

func Ftruncate(fd int, length int64) error {
	return unix.Ftruncate(fd, length)
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

func Setregid(rgid int, egid int) error {
	return syscall.Setregid(rgid, egid)
}

func Setreuid(ruid int, euid int) error {
	return syscall.Setreuid(ruid, euid)
}

func Setuid(uid int) error {
	return unix.Setuid(uid)
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

func Write(fd int, p []byte) (int, error) {
	return syscall.Write(fd, p)
}
