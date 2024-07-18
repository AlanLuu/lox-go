package syscalls

import (
	"syscall"

	"github.com/AlanLuu/lox/loxerror"
)

func unsupported(name string) error {
	return loxerror.Error("'os." + name + "' is unsupported on Windows.")
}

func Chroot(path string) error {
	return unsupported("chroot")
}

func Execvp(file string, argv []string) error {
	return unsupported("execvp")
}

func Mkfifo(path string, mode uint32) error {
	return unsupported("mkfifo")
}

func Read(fd int, p []byte) (int, error) {
	return syscall.Read(syscall.Handle(fd), p)
}

func Setegid(egid int) error {
	return unsupported("setegid")
}

func Seteuid(euid int) error {
	return unsupported("seteuid")
}

func Setgid(gid int) error {
	return unsupported("setgid")
}

func Setregid(rgid int, egid int) error {
	return unsupported("setregid")
}

func Setreuid(ruid int, euid int) error {
	return unsupported("setreuid")
}

func Setuid(uid int) error {
	return unsupported("setuid")
}

type UnameResult struct {
	Sysname  string
	Nodename string
	Release  string
	Version  string
	Machine  string
}

func Uname() (UnameResult, error) {
	return UnameResult{}, unsupported("uname")
}

func Write(fd int, p []byte) (int, error) {
	return syscall.Write(syscall.Handle(fd), p)
}
