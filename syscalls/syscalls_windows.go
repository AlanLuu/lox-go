package syscalls

import "github.com/AlanLuu/lox/loxerror"

func unsupported(name string) error {
	return loxerror.Error("'os." + name + "' is unsupported on Windows.")
}

func Mkfifo(path string, mode uint32) error {
	return unsupported("mkfifo")
}