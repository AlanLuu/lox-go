package ast

import "github.com/AlanLuu/lox/ast/filemode"

type LoxLazyFile struct {
	*LoxFile
	path     string
	mode     filemode.FileMode
	isBinary bool
}

func NewLoxLazyFile(path string, mode filemode.FileMode, isBinary bool) *LoxLazyFile {
	return &LoxLazyFile{
		LoxFile:  nil,
		path:     path,
		mode:     mode,
		isBinary: isBinary,
	}
}

func (l *LoxLazyFile) LazyTypeEval() error {
	if l.LoxFile == nil {
		file, err := filemode.Open(l.path, l.mode)
		if err != nil {
			return err
		}
		l.LoxFile = &LoxFile{
			file:       file,
			name:       file.Name(),
			mode:       l.mode,
			isBinary:   l.isBinary,
			stat:       nil,
			properties: make(map[string]any),
		}
	}
	return nil
}
