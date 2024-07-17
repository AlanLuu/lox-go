package syscalls

import "github.com/AlanLuu/lox/loxerror"

func unsupported(name string) error {
	return loxerror.Error("'os." + name + "' is unsupported on Windows.")
}

func Chroot(path string) error {
	return unsupported("chroot")
}

func Mkfifo(path string, mode uint32) error {
	return unsupported("mkfifo")
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
