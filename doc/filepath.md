# File path methods and fields

The following fields are defined in the built-in `filepath` class:
- `filepath.listSep`, which is a string that is the path separator character for the current OS
- `filepath.sep`, which is a string that is the path list separator character for the current OS
- `filepath.skipAll`, which is equal to `2` and can be used as a return value in the callback function argument of `filepath.walk`
- `filepath.skipDir`, which is equal to `1` and can be used as a return value in the callback function argument of `filepath.walk`

The following methods are defined in the built-in `filepath` class:
- `filepath.abs(path)`
- `filepath.base(path)`
- `filepath.clean(path)`
- `filepath.dir(path)`
- `filepath.evalSymlinks(path)`
- `filepath.ext(path)`
- `filepath.exts(path)`
- `filepath.fileInfo(path)`
- `filepath.fileInfoNil(path)`
- `filepath.fromSlash(path)`
- `filepath.glob(path)`
- `filepath.isAbs(path)`
- `filepath.isLocal(path)`
- `filepath.join(elements...)`
- `filepath.match(pattern, name)`
- `filepath.split(path)`
- `filepath.splitList(path)`
- `filepath.stem(path)`
- `filepath.toSlash(path)`
- `filepath.volumeName(path)`
- `filepath.walk(path, callback)`
- `filepath.walkFileInfo(path, callback)`
- `filepath.walkIter(path)`
- `filepath.walkStrIter(path)`
