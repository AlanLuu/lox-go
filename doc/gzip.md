# gzip methods and fields

The following compression level fields are defined in the built-in `gzip` class, which are all integers:
- `gzip.bestCompression`, which is equal to `9`
- `gzip.bestSpeed`, which is equal to `1`
- `gzip.defaultCompression`, which is equal to `-1`
- `gzip.huffmanOnly`, which is equal to `-2`
- `gzip.noCompression`, which is equal to `0`

The following fields are defined in the built-in `gzip` class, which are all integers:
- `gzip.USE_BUFFER`

The following methods are defined in the built-in `gzip` class:
- `gzip.buffer(buffer/file/string, [compressionLevel])`, which returns a buffer of the raw bytes of the data from the specified buffer, file object, or string compressed in the gzip format with the specified gzip compression level as an integer. If `compressionLevel` is omitted, `gzip.defaultCompression` is used as the compression level
- `gzip.write(file, buffer/file/string, [compressionLevel])`, which writes to the specified file the raw bytes of the data from the specified buffer, file object, or string compressed in the gzip format with the specified gzip compression level as an integer. If `compressionLevel` is omitted, `gzip.defaultCompression` is used as the compression level
- `gzip.writer(file/gzip.USE_BUFFER)`, which returns a gzip writer object that writes to the specified file object. If `gzip.USE_BUFFER` is specified instead of a file object, the returned gzip writer object writes to an internal buffer instead
- `gzip.writerLevel(file/gzip.USE_BUFFER, compressionLevel)`, which returns a gzip writer object with the specified compression level integer that writes to the specified file object. If `gzip.USE_BUFFER` is specified instead of a file object, the returned gzip writer object writes to an internal buffer instead

gzip writer objects have the following methods associated with them:
- `gzip writer.buffer()`, which returns a buffer of the raw bytes of the final gzip-compressed data
    - The current gzip writer object must have been created using `gzip.writer(gzip.USE_BUFFER)` or else this method throws a runtime error
    - The current gzip writer object must be flushed or closed before calling this method
- `gzip writer.close()`, which closes the current gzip writer object and writes the gzip-compressed bytes to the specified file or buffer, depending on how the gzip writer object was created
- `gzip writer.flush()`, which writes the gzip-compressed bytes to the specified file or buffer, depending on how the gzip writer object was created, without closing the current gzip writer object and returns the current gzip writer object itself
- `gzip writer.isBuffer()`, which returns `true` if the current gzip writer object was created using `gzip.writer(gzip.USE_BUFFER)` and `false` otherwise
- `gzip writer.isClosed()`, which returns `true` if the current gzip writer object is closed and `false` otherwise
- `gzip writer.reset(file/gzip.USE_BUFFER)`, which resets the state of the current gzip writer object to the specified file object, clearing all internal buffers as well. If `gzip.USE_BUFFER` is specified instead of a file object, the gzip writer object is reset to a new internal buffer
    - This method throws a runtime error if the gzip writer object is closed
- `gzip writer.write(content)`, which writes the specified content, which is a buffer, file object, or string, into the current gzip object, which will include the content in the final gzip-compressed data when the current gzip object is flushed or closed
    - This method throws a runtime error if the gzip writer object is closed
