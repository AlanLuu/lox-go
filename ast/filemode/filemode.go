package filemode

import (
	"os"

	"github.com/AlanLuu/lox/loxerror"
)

type FileMode int

const (
	READ FileMode = iota
	WRITE
	APPEND
	READ_WRITE
)

var Modes = map[byte]FileMode{
	'r': READ,
	'w': WRITE,
	'a': APPEND,
}

var ModeStrings = map[FileMode]string{
	READ:       "r",
	WRITE:      "w",
	APPEND:     "a",
	READ_WRITE: "rw",
}

func Open(path string, fileMode FileMode, root *os.Root) (*os.File, error) {
	switch fileMode {
	case READ:
		if root != nil {
			return root.Open(path)
		}
		return os.Open(path)
	case WRITE:
		if root != nil {
			return root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		}
		return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	case APPEND:
		if root != nil {
			return root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		}
		return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	case READ_WRITE:
		if root != nil {
			return root.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		}
		return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	default:
		return nil, loxerror.Error("Unknown file mode.")
	}
}
