# CSV methods and fields

The following methods are defined in the built-in `csv` class:
- `csv.reader(file/string, [delimiter])`, which returns a CSV reader object that reads from the specified file object or string with the specified delimiter that must be a single-character string. If the delimiter is omitted, the character `,` is used as the delimiter
- `csv.writer(file, [delimiter])`, which returns a CSV writer object that writes to the specified file object with the specified delimiter that must be a single-character string. If the delimiter is omitted, the character `,` is used as the delimiter

CSV reader objects have the following methods associated with them:
- `csv reader.read()`, which reads in a line and returns a list of strings of the values that are separated by the delimiter
- `csv reader.readAll()`, which reads in all lines and returns a list of lists, where the inner lists contain strings of the values that are separated by the delimiter

CSV writer objects have the following methods associated with them:
- `csv writer.bufferedWrite(list)`, which buffers a write of the list of values to the file object, separated by the delimiter
    - To flush the buffer and write all buffered data to the file object, call the method `csv writer.flush`
    - If an element in `list` is not a string, the string representation of that element is used as the value to write to the file
- `csv writer.flush()`, which writes all buffered data to the file object
- `csv writer.write(list)`, which writes the list of values to the file object, separated by the delimiter
    - If an element in `list` is not a string, the string representation of that element is used as the value to write to the file
- `csv writer.writeAll(list)`, which takes in a list of iterables and for each iterable in the list, write all their iterated elements on a new row separated by the delimiter, incrementing the row number for each iterable in the list
    - If the specified list contains an element that is not an iterable, a runtime error is thrown
    - If an element in an iterable is not a string, the string representation of that element is used as the value to write to the file
