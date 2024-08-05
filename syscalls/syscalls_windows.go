package syscalls

import (
	"syscall"

	"github.com/AlanLuu/lox/loxerror"
)

func unsupported(name string) error {
	return loxerror.Error("'os." + name + "' is unsupported on Windows.")
}

func Close(fd int) error {
	return syscall.Close(syscall.Handle(fd))
}

func Chroot(path string) error {
	return unsupported("chroot")
}

func Dup(oldfd int) (int, error) {
	return -1, unsupported("dup")
}

func Dup2(oldfd int, newfd int) error {
	return unsupported("dup2")
}

func Execv(path string, argv []string) error {
	return unsupported("execv")
}

func Execve(path string, argv []string, envp []string) error {
	return unsupported("execve")
}

func Execvp(file string, argv []string) error {
	return unsupported("execvp")
}

func Execvpe(file string, argv []string, envp []string) error {
	return unsupported("execvpe")
}

func Ftruncate(fd int, length int64) error {
	return syscall.Ftruncate(syscall.Handle(fd), length)
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
