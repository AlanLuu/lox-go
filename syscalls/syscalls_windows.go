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

func Fchdir(fd int) error {
	return unsupported("fchdir")
}

func Fchmod(fd int, mode uint32) error {
	return unsupported("fchmod")
}

func Fchown(fd int, uid int, gid int) error {
	return unsupported("fchown")
}

func ForkExec(argv0 string, argv []string, attr *syscall.ProcAttr) (int, error) {
	return -1, unsupported("forkExec")
}

func ForkExecvp(argv0 string, argv []string, attr *syscall.ProcAttr) (int, error) {
	return -1, unsupported("forkExecvp")
}

func ForkExecFd(argv0 string, argv []string) (int, error) {
	return ForkExec(argv0, argv, nil)
}

func ForkExecveFd(argv0 string, argv []string, env []string) (int, error) {
	return ForkExec(argv0, argv, nil)
}

func ForkExecvpFd(argv0 string, argv []string) (int, error) {
	return ForkExecvp(argv0, argv, nil)
}

func ForkExecvpeFd(argv0 string, argv []string, env []string) (int, error) {
	return ForkExecvp(argv0, argv, nil)
}

func Fsync(fd int) error {
	return syscall.FlushFileBuffers(syscall.Handle(fd))
}

func Ftruncate(fd int, length int64) error {
	return syscall.Ftruncate(syscall.Handle(fd), length)
}

func Getgroups() ([]int, error) {
	return nil, unsupported("getgroups")
}

func Getpagesize() int {
	return syscall.Getpagesize()
}

func Getsid(pid int) (int, error) {
	return -1, unsupported("getsid")
}

func Mkfifo(path string, mode uint32) error {
	return unsupported("mkfifo")
}

func Pipe(p *[2]int) error {
	var fd [2]syscall.Handle
	if err := syscall.Pipe(fd[:]); err != nil {
		return err
	}
	p[0], p[1] = int(fd[0]), int(fd[1])
	return nil
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

func Setgroups(gids []int) error {
	return unsupported("setgroups")
}

func Setregid(rgid int, egid int) error {
	return unsupported("setregid")
}

func Setreuid(ruid int, euid int) error {
	return unsupported("setreuid")
}

func Setsid() (int, error) {
	return -1, unsupported("setsid")
}

func Setuid(uid int) error {
	return unsupported("setuid")
}

func Sync() {}

func Umask(mask int) int {
	return 0
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

func Wait() (int, WaitStatus, error) {
	return -1, WaitStatus{}, unsupported("wait")
}

func Write(fd int, p []byte) (int, error) {
	return syscall.Write(syscall.Handle(fd), p)
}
