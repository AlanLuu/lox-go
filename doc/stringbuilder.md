# Stringbuilder methods

Stringbuilder objects have the following methods associated with them:
- `stringbuilder.append(string)`, which appends the specified string argument to the current stringbuilder and returns the current stringbuilder itself
- `stringbuilder.appendBuf(buffer)`, which appends the specified buffer argument to the current stringbuilder and returns the current stringbuilder itself
- `stringbuilder.appendBuilder(stringbuilder2)`, which appends the specified stringbuilder argument's string to the current stringbuilder and returns the current stringbuilder itself
- `stringbuilder.appendByte(byte)`, which appends the specified byte integer argument to the current stringbuilder and returns the current stringbuilder itself
    - The byte integer argument must be between 0 and 255 or else a runtime error is thrown
- `stringbuilder.appendCodePoint(codePoint)`, which appends the specified Unicode code point integer argument to the current stringbuilder and returns the current stringbuilder itself
- `stringbuilder.buffer()`, which returns the current contents of the current stringbuilder as a buffer
- `stringbuilder.cap()`, which is an alias for `stringbuilder.capacity`
- `stringbuilder.capacity()`, which returns the capacity of the current stringbuilder as an integer
- `stringbuilder.clear()`, which clears the contents of the current stringbuilder and returns the current stringbuilder itself
- `stringbuilder.equals(stringbuilder2)`, which returns `true` if the specified stringbuilder argument's string is equal to the current stringbuilder's string and `false` otherwise
- `stringbuilder.grow(factor)`, which increases the capacity of the current stringbuilder by the specified integer factor and returns the current stringbuilder itself
    - The integer argument must not be negative or else a runtime error is thrown
- `stringbuilder.len()`, which is an alias for `stringbuilder.length`
- `stringbuilder.length()`, which returns the length of the current stringbuilder's string as an integer
- `stringbuilder.numBytes()`, which returns the number of bytes of the current stringbuilder's string as an integer
- `stringbuilder.string()`, which returns the current contents of the current stringbuilder as a string
    - If this method was called before and is called again without changing the contents of the current stringbuilder, this method returns the same string that was originally returned from the initial call until the contents of the current stringbuilder are modified
- `stringbuilder.stringNew()`, which returns the current contents of the current stringbuilder as a new string in memory
