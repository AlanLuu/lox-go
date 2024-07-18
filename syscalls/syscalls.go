package syscalls

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/AlanLuu/lox/loxerror"
)

func execCommandNotFound(funcName string, path string) error {
	return loxerror.Error(fmt.Sprintf("os.%v: %v: command not found", funcName, path))
}

func Execvp(file string, argv []string) error {
	fullPath, err := exec.LookPath(file)
	if err != nil {
		return execCommandNotFound("execvp", file)
	}
	return syscall.Exec(fullPath, argv, os.Environ())
}
