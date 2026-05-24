# bufio class methods

The following fields are defined in the built-in `bufio` class:
- `bufio.scanLines` (funcNum)
- `bufio.scanBytes` (funcNum)
- `bufio.scanChars` (funcNum)
- `bufio.scanWords` (funcNum)
- `bufio.maxScanTokenSize`

The following methods are defined in the built-in `bufio` class:
- `bufio.reader(file)`
- `bufio.readerConn(conn)`
- `bufio.readerConnSize(conn, size)`
- `bufio.readerHTTPResponse(response)`
- `bufio.readerHTTPResponseSize(response, size)`
- `bufio.readerSize(file, size)`
- `bufio.readerStdin()`
- `bufio.readerStdinSize(size)`
- `bufio.readerString(str)`
- `bufio.readerStringSize(str, size)`
- `bufio.scanner(file)`
- `bufio.scannerConn(conn)`
- `bufio.scannerConnFunc(conn, funcNum)`
- `bufio.scannerFunc(file, funcNum)`
- `bufio.scannerHTTPResponse(response)`
- `bufio.scannerHTTPResponseFunc(response, funcNum)`
- `bufio.scannerStdin()`
- `bufio.scannerStdinFunc(funcNum)`
- `bufio.scannerString(str)`
- `bufio.scannerStringFunc(str, funcNum)`
- `bufio.writer(file)`
- `bufio.writerConn(conn)`
- `bufio.writerConnSize(conn, size)`
- `bufio.writerSize(file, size)`
- `bufio.writerStderr()`
- `bufio.writerStderrSize(size)`
- `bufio.writerStdout()`
- `bufio.writerStdoutSize(size)`
- `bufio.writerStringBuilder(builder)`
- `bufio.writerStringBuilderSize(builder, size)`

Buffered reader objects have the following methods associated with them:
- `buffered reader.buffered()`
- `buffered reader.copyToBufWriter(bufWriter)`
- `buffered reader.discard(num)`
- `buffered reader.peek(num)`
- `buffered reader.read(buffer)`
- `buffered reader.readBuf(delim)`, which is an alias for `buffered reader.readBuffer`
- `buffered reader.readBuffer(delim)`
- `buffered reader.readByte()`
- `buffered reader.readBytes(num)`
- `buffered reader.readBytesIter(num)`
- `buffered reader.readChar()`
- `buffered reader.readChars(num)`
- `buffered reader.readCharsIter(num)`
- `buffered reader.readStr(delim)`, which is an alias for `buffered reader.readString`
- `buffered reader.readStrs(delim, num)`, which is an alias for `buffered reader.readStrings`
- `buffered reader.readStrsIter(delim, num)`, which is an alias for `buffered reader.readStringsIter`
- `buffered reader.readString(delim)`
- `buffered reader.readStrings(delim, num)`
- `buffered reader.readStringsIter(delim, num)`
- `buffered reader.readToFile(file)`
- `buffered reader.reset(file)`
- `buffered reader.size()`
- `buffered reader.unreadByte()`
- `buffered reader.unreadByteBool()`
- `buffered reader.unreadChar()`
- `buffered reader.unreadCharBool()`

Buffered writer objects have the following methods associated with them:
- `buffered writer.available()`
- `buffered writer.availableBuf()`
- `buffered writer.buffered()`
- `buffered writer.copyBufReader(bufReader)`
- `buffered writer.flush()`
- `buffered writer.flushBool()`
- `buffered writer.flushForce()`
- `buffered writer.isFile(file)`
- `buffered writer.isStderr()`
- `buffered writer.isStdout()`
- `buffered writer.reset(file)`
- `buffered writer.size()`
- `buffered writer.write(byte/buffer/file/string)`
- `buffered writer.writeArgs(args...)`
- `buffered writer.writeList(list)`

Buffered scanner objects have the following methods associated with them:
- `buffered scanner.buffer(size)`
- `buffered scanner.bytes()`
- `buffered scanner.err()`
- `buffered scanner.errThrow()`
- `buffered scanner.scan()`
- `buffered scanner.scanBytes()`
- `buffered scanner.scanBytesErr()`
- `buffered scanner.scanBytesErrThrow()`
- `buffered scanner.scanBytesIter()`
- `buffered scanner.scanErr()`
- `buffered scanner.scanErrThrow()`
- `buffered scanner.scanText()`
- `buffered scanner.scanTextErr()`
- `buffered scanner.scanTextErrThrow()`
- `buffered scanner.scanTextIter()`
- `buffered scanner.splitFunc(funcNum)`
- `buffered scanner.startedScanning()`
- `buffered scanner.text()`
