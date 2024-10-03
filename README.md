# lox-go
A Golang implementation of the Lox language from the book [Crafting Interpreters](https://craftinginterpreters.com/) with my own features

# Usage
Supports running code from a file:
```
lox code.lox
```
or from stdin as a REPL:
```
lox
```
or as a string from the terminal:
```
lox -c <code>
```

# Installation
First, [install Go](https://go.dev/doc/install) if it's not installed already. Then run the following commands to build this interpreter:
```
git clone https://github.com/AlanLuu/lox-go.git
cd lox-go
go build
```
This will create an executable binary called `lox` on Linux/macOS and `lox.exe` on Windows that can be run directly.

# Differences from Lox
- Concatenating a string with another data type will convert that type into a string and concatenate them together
- Integers and floats are distinct types in this implementation of Lox
    - Integers are signed 64-bit values and floats are double-precision floating-point values
    - A binary operation on an integer and float converts the integer into a float before performing the specified operation
    - The `/` operation on two integers converts both operands to floats before performing the division. If the result of the division can be represented as an integer, the final result is an integer, otherwise it is a float
    - The `**` operation on two integers converts both operands to floats before performing the exponentiation. After the operation, the result of the exponentiation is converted into an integer, which is used as the final result
- The following additional operations are supported in this implementation of Lox:
    - `a % b`, which returns the remainder of two numbers `a` and `b`
    - `a ** b`, which returns `a` raised to the power of `b`, where `a` and `b` are numbers
        - The exponent operator has higher precedence than any unary operators on the left, so `-a ** b` is equivalent to `-(a ** b)`.
    - `a << b` and `a >> b`, which returns a number representing the number `a` shifted by `b` bits to the left and right respectively.
        - If `a` or `b` are floats, they are converted into integers before the shift operation
    - `~a`, which returns the bitwise NOT of the number `a`
        - If `a` is a float, it is converted into an integer before the bitwise operation
    - `a & b`, `a | b`, and `a ^ b`, which returns the bitwise AND, OR, and XOR of two numbers `a` and `b` respectively.
        - If `a` or `b` are floats, they are converted into integers before the bitwise operation
        - Unlike in C, the precedence of the bitwise operators is higher than the precedence of the comparison operators, so `a & b == value` is equivalent to `(a & b) == value`
    - The ternary operator `a ? b : c`, which evaluates to `b` if `a` is a truthy value and `c` otherwise
- Division by 0 results in `Infinity`, which uses Golang's `math.Inf()` under the hood
    - `Infinity` literals are supported using the identifier "Infinity"
- Performing a binary operation that isn't supported between two types results in `NaN`, which stands for "not-a-number", using Golang's `math.NaN()` under the hood
    - `NaN` literals are supported using the identifier "NaN"
- Booleans and `nil` are treated as integers when performing arithmetic operations on them, with `true` and `false` being treated as `1` and `0` respectively, and `nil` being treated as `0`
- Besides `false` and `nil`, the values `0`, `0.0`, `0n`, `0.0n`, `NaN`, `""`, `[]`, `{}`, `Set()`, and `Buffer()` are also falsy values
- The `&&` and `||` operators are supported for the logical AND and OR operations respectively
- Binary, hexadecimal, and octal integer literals are supported in this implementation of Lox
    - Binary literals start with the prefix `0b`
    - Hexadecimal literals start with the prefix `0x`
    - Octal literals start with the prefixes `0o` or `0`
    - All letters in prefixes are case-insensitive
- Scientific notation number literals are supported in this implementation of Lox
    - Examples: `1e5`, `1e+5`, `1e-5`, `1.5e5`, `1.5e+5`, `1.5e-5`
    - All scientific notation number literals are floats
- Number literals support the following features:
    - An underscore character can be used to group digits, such as `1_000_000`, which is equivalent to `1000000`
        - Underscore characters are also allowed in binary, hexadecimal, and octal literals, except for octal literals starting with a `0`
- Arbitrary-precision integers and floats are supported in this implementation of Lox, which are called bigints and bigfloats respectively
    - Examples: `1n`, `1.5n`
    - bigints and bigfloats support the same operations as integers and floats, with the following notes:
        - bigfloats do not support the `**` operator
        - Dividing a bigint by `0` or `0n` throws a runtime error
- A new statement called `put` prints an expression without a newline character at the end
    - Syntax: `put <expression>;`
- `break` and `continue` statements are supported in this implementation of Lox
- For loops are implemented with their own AST node instead of being desugared into while loop nodes
    - This makes it easier to implement the `continue` statement inside for loops
- Variables declared in the initialization part of a for loop are locally scoped to that loop and do not become global variables
- Do while loops are supported in this implementation of Lox
    ```js
    var i = 1;
    do {
        print i;
        i = i + 1;
    } while (i <= 10);
    ```
- Foreach loops are supported in this implementation of Lox
    ```js
    var iterable = [1, 2, 3, 4, 5];
    foreach (var element in iterable) {
        print element;
    }
    ```
    - The target of a foreach loop must be an iterable type. Iterable types are the following:
        - String
            - For each iteration, `element` is each character of the string iterated in order
        - List
            - For each iteration, `element` is each element of the list iterated in order
        - Buffer
            - For each iteration, `element` is each element of the buffer iterated in order
        - Dictionary
            - For each iteration, `element` is a list with two elements, with the first element being a dictionary key and the second element being the dictionary value corresponding to that key
        - Range
            - For each iteration, `element` is each generated integer from the range object
        - Set
            - For each iteration, `element` is each element of the set
    - Note: when iterating over dictionaries or sets using a foreach loop, the iteration order is random since dictionaries and sets are unordered
- Try-catch-finally statements are supported in this implementation of Lox
    ```js
    try {
        print i;
    } catch (e) {
        print "caught error";
    } finally {
        print "done";
    }
    ```
    - A `try` statement must be followed by a `catch` or `finally` statement or both
    - If the exception variable is not needed, it may be omitted from the `catch` statement
        ```js
        try {
            print i;
        } catch {
            print "caught error";
        }
        ```
    - Along with try-catch-finally statements, `throw` statements are supported in this implementation of Lox
        - Syntax: `throw <expression>;`
        - `throw` statements throw a runtime error using the provided expression as the error message. If the provided expression is an error object, the object itself is thrown. Otherwise, if the provided expression is not a string, the string representation of the expression is used as the error message
- Assert statements are supported in this implementation of Lox
    ```java
    assert 1 == 1;
    assert 1 == 2; //Throws a runtime error
    ```
    - If the specified expression is false, a runtime error is thrown, otherwise the statement does nothing
- Anonymous function expressions are supported in this implementation of Lox. There are two forms supported:
    - `fun(param1, paramN) {<statements>}`, which is a traditional anonymous function expression that contains a block with statements
    - `fun(param1, paramN) => <expression>`, which is an arrow function expression that implicitly returns the given expression when called
    - The parser will attempt to parse anonymous function expressions that appear on their own line as function declarations, throwing a parser error as a result. This is expected behavior; to force the parser to parse them as expressions, wrap the function expression inside parentheses, like `(fun() {})()`. In this case, this creates an anonymous function expression that is called immediately
- The spread operator `...` is supported in this implementation of Lox
    - Examples:
        - `function(a, ...iterable, b)`, which passes all elements in the iterable as arguments to the specified function
            ```js
            var arr = [1, 2, 3];
            var arr2 = [1, 2, ...arr, 4];
            print arr2; //[1, 2, 1, 2, 3, 4]
            ```
            ```js
            var dict = {
                1: 2,
                3: 4
            };
            var dict2 = {
                "key": "value",
                ...dict,
                "key2": "value2"
            };
            print dict2; //{"key": "value", 1: 2, 3: 4, "key2": "value2"}
            ```
- Static class fields and methods are supported in this implementation of Lox
    - Classes also support initializing instance fields to an initial value directly in the class body without the need for a constructor
    ```js
    class A {
        static x = 10;
        static y() {
            return 20;
        }
        z = 30;
    }
    print A.x; //Prints "10"
    print A.y(); //Prints "20"
    var a = A();
    print a.z; //Prints "30"
    ```
- Various mathematical methods and constants are defined under a built-in class called `Math`, which is documented [here](./doc/Math.md)
- Various methods to work with JSON strings are defined under a built-in class called `JSON`, which is documented [here](./doc/JSON.md)
- Various methods and fields to work with operating system functionality are defined under a built-in class called `os`, which is documented [here](./doc/os.md)
- Various methods to work with HTTP requests are defined under a built-in class called `http`, which is documented [here](./doc/http.md)
- Various methods to work with cryptographic functionality are defined under a built-in class called `crypto`, which is documented [here](./doc/crypto.md)
- Various methods to work with CSV files are defined under a built-in class called `csv`, which is documented [here](./doc/csv.md)
- Various methods to work with regular expressions are defined under a built-in class called `regex`, which is documented [here](./doc/regex.md)
- Various methods to work with processes are defined under a built-in class called `process`, which is documented [here](./doc/process.md)
- Various methods and fields to work with Windows-specific functionality are defined under a built-in class called `windows`, which is documented [here](./doc/windows.md)
    - This class does not exist on non-Windows systems
- Various methods to work with base64 strings are defined under a built-in class called `base64`, where the following methods are defined:
    - `base64.decode(string)`, which decodes the specified base64-encoded string into a decoded string and returns that string
        - A runtime error is thrown if the specified string is not properly encoded as base64
    - `base64.decodeToBuf(string)`, which decodes the specified base64-encoded string into a buffer and returns that buffer
        - A runtime error is thrown if the specified string is not properly encoded as base64
    - `base64.encode(arg)`, which encodes the specified argument, which is either a string or a buffer, into a base64 string and returns that encoded string
- Various methods to work with base32 strings are defined under a built-in class called `base32`, where the following methods are defined:
    - `base32.decode(string)`, which decodes the specified base32-encoded string into a decoded string and returns that string
        - A runtime error is thrown if the specified string is not properly encoded as base32
    - `base32.decodeToBuf(string)`, which decodes the specified base32-encoded string into a buffer and returns that buffer
        - A runtime error is thrown if the specified string is not properly encoded as base32
    - `base32.encode(arg)`, which encodes the specified argument, which is either a string or a buffer, into a base32 string and returns that encoded string
- Various methods to work with hexadecimal strings are defined under a built-in class called `hexstr`, where the following methods are defined:
    - `hexstr.decode(hexStr)`, which decodes the specified hexadecimal string into a buffer and returns that buffer
    - `hexstr.decodeToStr(hexStr)` which decodes the specified hexadecimal string into a decoded string and returns that string
    - `hexstr.dump(arg)`, which returns a string containing the hex dump of the specified argument, which is either a string or a buffer
    - `hexstr.encode(arg)`, which encodes the specified argument, which is either a string or a buffer, into a hexadecimal string and returns that encoded string
- Various methods and fields to work with integers are defined under a built-in class called `Integer`, where the following methods and fields are defined:
    - `Integer.MAX`, which is the maximum value that an integer can store
    - `Integer.MAX8`, which is the maximum value that an 8-bit integer can store
    - `Integer.MAX16`, which is the maximum value that a 16-bit integer can store
    - `Integer.MAX32`, which is the maximum value that a 32-bit integer can store
    - `Integer.MAXU8`, which is the maximum value that an unsigned 8-bit integer can store
    - `Integer.MAXU16`, which is the maximum value that an unsigned 16-bit integer can store
    - `Integer.MAXU32`, which is the maximum value that an unsigned 32-bit integer can store
    - `Integer.MIN`, which is the minimum value that an integer can store
    - `Integer.MIN8`, which is the minimum value that an 8-bit integer can store
    - `Integer.MIN16`, which is the minimum value that a 16-bit integer can store
    - `Integer.MIN32`, which is the minimum value that a 32-bit integer can store
    - `Integer.parseInt(string)`, which attempts to convert the specified string argument into an integer and returns that integer if successful, otherwise a runtime error is thrown
    - `Integer.toFloat(integer)`, which converts the specified integer argument into a float and returns that float
    - `Integer.toString(integer)`, which returns the string representation of the specified integer argument
- Various methods and fields to work with floats are defined under a built-in class called `Float`, where the following methods and fields are defined:
    - `Float.MAX`, which is the maximum value that a float can store
    - `Float.MAX32`, which is the maximum value that a 32-bit float can store
    - `Float.MIN`, which is the minimum value that a float can store
    - `Float.MIN32`, which is the minimum value that a 32-bit float can store
    - `Float.parseFloat(string)`, which attempts to convert the specified string argument into a float and returns that float if successful, otherwise a runtime error is thrown
    - `Float.toInt(float)`, which converts the specified float argument into an integer and returns that integer
    - `Float.toString(float)`, which returns the string representation of the specified float argument
- Various methods to work with bigints and bigfloats are defined under built-in classes called `bigint` and `bigfloat` respectively, which are documented [here](./doc/bignum.md)
- Various methods and fields that correspond to string constants and utility operations are defined under a built-in class called `String`, where the following methods and fields are defined:
    - `String.digits`, which is the string `"0123456789"`
    - `String.hexDigits`, which is the string `"0123456789abcdefABCDEF"`
    - `String.hexDigitsLower`, which is the string `"0123456789abcdef"`
    - `String.hexDigitsUpper`, which is the string `"0123456789ABCDEF"`
    - `String.letters`, which is the string `"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"`
    - `String.lowercase`, which is the string `"abcdefghijklmnopqrstuvwxyz"`
    - `String.octDigits`, which is the string `"01234567"`
    - `String.punctuation`, which is the string ``"!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"``
    - `String.qwertyLower`, which is the string `"qwertyuiopasdfghjklzxcvbnm"`
    - `String.qwertyUpper`, which is the string `"QWERTYUIOPASDFGHJKLZXCVBNM"`
    - `String.toString(arg)`, which returns the string representation of the specified argument
    - `String.uppercase`, which is the string `"ABCDEFGHIJKLMNOPQRSTUVWXYZ"`
- Methods to work with random number generation are defined under a built-in-class called `Rand`, which is documented [here](./doc/Rand.md)
- Strings have some additional features associated with them:
    - Strings can be represented using single quotes as well
    - Strings can be indexed by an integer, which will return a new string with only the character at the specified index: `string[index]`
    - Get a new string with all characters from indexes `start` to `end` exclusive, where `start` and `end` are integers and `start < end`: `string[start:end]`
        - If `start >= end`, a new empty string is returned
        - `start` or `end` can be omitted, in which case the starting index will have a value of `0` if `start` is omitted and the ending index will have a value of `len(string)` if `end` is omitted
    - Negative integers are supported for string indexes, where a negative index `i` is equivalent to the index `i + len(string)`. For example, `string[-1]` refers to the last character in the string
    - It is a runtime error to use a negative index value whose absolute value is greater than the length of the string or a positive index value greater than or equal to the length of the string to index into that string
    - Get a new string that is the original string repeated `n` times, where `n` is an integer: `string * n`
    - Escape characters in strings are supported:
        - `\'`: single quote
        - `\"`: double quote
        - `\\`: backslash
        - `\a`: bell
        - `\n`: newline
        - `\r`: carriage return
        - `\t`: horizontal tab
        - `\b`: backspace
        - `\f`: form feed
        - `\v`: vertical tab
    - Besides these features, strings also have some methods associated with them:
        - `string.capitalize()`, which returns a new string with the first character from the original string capitalized if possible and the rest of the characters in lowercase if possible
        - `string.compare(string2)`, which lexicographically compares `string` and `string2` and returns `0` if `string == string2`, `-1` if `string < string2`, and `1` if `string > string2`
        - `string.contains(substr)`, which returns `true` if `substr` is contained within `string` and `false` otherwise
        - `string.endsWith(suffix)`, which returns `true` if `string` ends with `suffix` and `false` otherwise
        - `string.index(string2)`, which returns an integer representing the index value of the location of `string2` in `string`, or `-1` if `string2` is not in `string`
        - `string.isEmpty()`, which returns `true` if the length of the string is 0 and `false` otherwise
        - `string.lastIndex(string2)`, which returns an integer representing the index value of the last occurrence of `string2` in `string`, or `-1` if `string2` is not in `string`
        - `string.lower()`, which returns a new string with all lowercase letters
        - `string.lstrip([chars])`, which returns a new string with all leading characters from `chars` removed. If `chars` is omitted, this method returns a new string with all leading whitespace, newlines, and tabs removed
        - `string.padEnd(length, padStr)`, which pads the contents of `padStr` to the end of `string` until the new string is of length `length`
        - `string.padStart(length, padStr)`, which pads the contents of `padStr` to the beginning of `string` until the new string is of length `length`
        - `string.replace(oldStr, newStr)`, which returns a new string where all occurrences of `oldStr` in the original string are replaced with `newStr`
        - `string.rot13()`, which returns a new string that is the ROT13 encoding of the original string
        - `string.rot18()`, which returns a new string that is the ROT18 encoding of the original string
            - ROT18 is a variation of ROT13 that combines ROT13 with ROT5, which shifts numerical digits by a factor of 5
        - `string.rot47()`, which returns a new string that is the ROT47 encoding of the original string
        - `string.rstrip([chars])`, which returns a new string with all trailing characters from `chars` removed. If `chars` is omitted, this method returns a new string with all trailing whitespace, newlines, and tabs removed
        - `string.shuffled()`, which returns a new string that is a shuffled version of the original string
        - `string.split(delimiter)`, which returns a list containing all substrings that are separated by `delimiter`
        - `string.startsWith(prefix)`, which returns `true` if `string` begins with `prefix` and `false` otherwise
        - `string.strip([chars])`, which returns a new string with all leading and trailing characters from `chars` removed. If `chars` is omitted, this method returns a new string with all leading and trailing whitespace, newlines, and tabs removed
        - `string.swapcase()`, which returns a new string with all lowercase characters converted to uppercase and vice-versa
        - `string.title()`, which returns a new string where each word starts with a capital letter if possible and the remaining characters in each word are in lowercase if possible
        - `string.toBuffer()`, which converts `string` into a buffer with the raw UTF-8 byte representation of each character in the string as the buffer elements and returns that buffer
        - `string.toList()`, which converts `string` into a list with each character in the string as the list elements and returns that list
        - `string.toNum([base])`, which attempts to convert `string` into an integer or float and returns that value if successful and `NaN` otherwise. If `base` is specified, then this method will attempt to convert `string` that is represented as the specified base into an integer or float and returns that value if the conversion was successful and `NaN` otherwise
        - `string.toSet()`, which converts `string` into a set with each unique character in the string as the set elements and returns that set
        - `string.upper()`, which returns a new string with all uppercase letters
        - `string.zfill(length)`, which returns a new string where the character `'0'` is padded to the left until the new string is of length `length`. If a leading `'+'` or `'-'` sign is part of the original string, the `'0'` padding is inserted after the leading sign instead of before
- Lists are supported in this implementation of Lox
    - Create a list and assign it to a variable: `var list = [1, 2, 3];`
    - Get an element from a list by index, where `index` is an integer: `list[index]`
    - Get a new list with all elements from indexes `start` to `end` exclusive, where `start` and `end` are integers and `start < end`: `list[start:end]`
        - If `start >= end`, a new empty list is returned
        - `start` or `end` can be omitted, in which case the starting index will have a value of `0` if `start` is omitted and the ending index will have a value of `len(list)` if `end` is omitted
    - Set an element: `list[index] = value;`
    - Negative integers are supported for list indexes, where a negative index `i` is equivalent to the index `i + len(list)`. For example, `list[-1]` refers to the last element in the list
        - Negative indexes are also supported in list methods that accept integer values for list indexes as a parameter
    - It is a runtime error to use a negative index value whose absolute value is greater than the length of the list or a positive index value greater than or equal to the length of the list to get or set
    - Concatenate two lists together into a new list: `list + list2`
    - Get a new list with all elements from the original list repeated `n` times, where `n` is an integer: `list * n`
    - Besides these operations, lists also have some methods associated with them:
        - `list.all(callback)`, which returns `true` if the callback function returns `true` for all elements in the list and `false` otherwise
        - `list.any(callback)` which returns `true` if the callback function returns `true` for any element in the list and `false` otherwise
        - `list.append(element)`, which appends an element to the end of the list
        - `list.clear()`, which removes all elements from the list
        - `list.contains(element)`, which returns `true` if `element` is contained in the list and `false` otherwise
        - `list.copy()`, which returns a shallow copy of the original list as a new list
        - `list.count(element)`, which returns the number of times `element` appears in the list
        - `list.extend(list2)`, which appends every element from `list2` to the end of the list
        - `list.filter(callback)`, which returns a new list containing only the elements from the original list where the callback function returns a truthy value for them
        - `list.find(callback)`, which returns the first element in the list where the callback function returns `true`, or `nil` if the callback returns `false` for every element in the list
        - `list.findIndex(callback)`, which returns the index of the first element in the list where the callback function returns `true`, or `-1` if the callback returns `false` for every element in the list
        - `list.first()`, which returns the first element in the list. If the list is empty, a runtime error is thrown
        - `list.flatten()`, which returns a new list where all elements contained within nested lists are flattened into a list without any nested lists
        - `list.forEach(callback)`, which executes the callback function for each element in the list
        - `list.index(element)`, which returns the index value of the element's position in the list, or `-1` if the element is not in the list
        - `list.insert(index, element)`, which inserts an element into the list at the specified index
        - `list.isEmpty()`, which returns `true` if the list contains no elements and `false` otherwise
        - `list.join(separator)`, which concatenates all elements in the list into a string where each element is separated by a separator string
        - `list.last()`, which returns the last element in the list. If the list is empty, a runtime error is thrown
        - `list.lastIndex(element)`, which returns the index value of the last occurrence of the element in the list, or `-1` if the element is not in the list
        - `list.map(callback)`, which returns a new list with the results of calling a callback function on each element of the original list
        - `list.pop([index])`, which removes and returns the element at the specified index from the list. If `index` is omitted, this method removes and returns the last element from the list
        - `list.reduce(callback, [initialValue])`, which applies a reducer callback function on every element in the list from left to right and returns a single value
        - `list.reduceRight(callback, [initialValue])`, which applies a reducer callback function on every element in the list from right to left and returns a single value
        - `list.remove(element)`, which removes the first occurrence of `element` from the list. Returns `true` if the list contained `element` and `false` otherwise
        - `list.removeAll(element1, element2, ..., elementN)`, which removes all occurrences of each element passed into this method from the list. Returns `true` if an element was removed and `false` otherwise
        - `list.removeAllList(list2)`, which removes all occurrences of each element that are contained in the specified list argument from the list. If `list == list2`, removes all elements from the list. Returns `true` if an element was removed and `false` otherwise
        - `list.reverse()`, which reverses all elements in the list in place
        - `list.reversed()`, which returns a new list with all elements from the original list in reversed order
        - `list.shuffle()`, which shuffles all elements in the list in place
        - `list.shuffled()`, which returns a new list with all elements from the original list in shuffled order
        - `list.sort(callback)`, which sorts all elements in the list in place based on the results of the callback function
            - The callback function is called with two arguments `a` and `b`, which are the first and second elements from the list to compare respectively
            - The callback function should return an integer or float where
                - A negative value means that `a` goes before `b`
                - A positive value means that `a` goes after `b`
                - `0` or `0.0` means that `a` and `b` remain at the same places
            - If the callback function does not return an integer or float, it is equivalent to returning `0` from the function
        - `list.sorted(callback)`, which returns a new list with all elements from the original list in sorted order based on the results of the callback function, which has the same behavior as the function described in `list.sort`
        - `list.toBuffer()`, which attempts to return a new buffer with the elements from the list. If the list contains an element that cannot belong in a buffer, a runtime error is thrown
        - `list.toSet()`, which attempts to return a new set with the elements from the list. If the list contains an element that cannot belong in a set, a runtime error is thrown
        - `list.with(index, element)`, which returns a new list that is a copy of the original list with the original element at the specified index replaced with the new element
    - Two lists are compared based on whether they are the same length and for every index `i`, the element from the first list at index `i` is equal to the element from the second list at index `i`
    - Attempting to use an index value larger than the length of the list will cause a runtime error
- A buffer type is supported in this implementation of Lox, mainly for use in manipulating binary files
    - Buffers are similar to lists, except they can only contain integers between 0 and 255 inclusive, and attempting to set a buffer element to an invalid value will throw a runtime error
    - Buffers share the same syntax in regards to getting and setting elements, concatenating two buffers into a new buffer, and getting a new buffer with all elements from the original buffer repeated a number of times
    - Buffers share the same methods as lists, except that the usual element restrictions are in place in terms of adding and setting elements, and any shared methods that normally return lists return buffers instead
        - Notably, the `map` method on buffers throws a runtime error if its callback function ever returns a value that is not an integer or is an integer less than 0 or greater than 255
    - Besides the methods shared with lists, buffers also have the following methods associated with them:
        - `buffer.memfrob([num])`, which applies the XOR operation to each buffer element with the number 42, changing the original buffer as a result. If an integer `num` is specified, only `num` buffer elements starting with the first element are changed
        - `buffer.memfrobCopy([num])`, which returns a new buffer with the original buffer elements XORed with 42. If an integer `num` is specified, only `num` buffer elements starting with the first element are changed
        - `buffer.memfrobRange(start, [stop])`, which applies the XOR operation to each buffer element with the number 42 starting from index `start` and stopping at index `stop` exclusive, which are both integers, and changing the original buffer as a result. If `stop` is omitted, the length of the buffer is used as the stop value
        - `buffer.memfrobRangeCopy(start, [stop])`, which returns a new buffer with the original buffer elements remaining the same and the elements from integer indexes `start` to `stop` exclusive XORed with 42. If `stop` is omitted, the length of the buffer is used as the stop value
        - `buffer.toList()`, which returns a new list with the elements from the buffer
        - `buffer.toString()`, which attempts to convert the elements from the buffer into a string. If a portion of the buffer cannot be converted into a string, a runtime error is thrown, with the error message specifying the portion of the buffer that cannot be converted into a string
- Dictionaries are supported in this implementation of Lox
    - Create a dictionary and assign it to a variable: `var dict = {"key": "value"};`
    - Get an element from a dictionary by key: `dict[key]`
        - It is a runtime error to attempt to get an element using a key that is not in the dictionary
    - Set an element: `dict[key] = value;`
    - Merge two dictionaries together: `dict | dict2`
        - If a key exists in both `dict` and `dict2`, the key in the merged dictionary becomes associated with the value from `dict2`
    - The following cannot be used as dictionary keys: buffer, dictionary, list, set
    - Besides these operations, dictionaries also have some methods associated with them:
        - `dictionary.clear()`, which removes all keys from the dictionary
        - `dictionary.containsKey(key)`, which returns `true` if the specified key exists in the dictionary and `false` otherwise
        - `dictionary.copy()`, which returns a shallow copy of the original dictionary as a new dictionary
        - `dictionary.get(key, [defaultValue])`, which returns the value associated with the specified key from the dictionary, or `defaultValue` if the key doesn't exist in the dictionary and `defaultValue` is provided, or `nil` otherwise
        - `dictionary.isEmpty()`, which returns `true` if the dictionary contains no keys and `false` otherwise
        - `dictionary.keys()`, which returns a list of all the keys in the dictionary in no particular order
        - `dictionary.removeKey(key)`, which removes the specified key from the dictionary and returns the value originally associated with the key or `nil` if the key doesn't exist in the dictionary. Note that a return value of `nil` can also mean that the specified key had a value of `nil`
        - `dictionary.values()`, which returns a list of all the values in the dictionary in no particular order
- Sets are supported in this implementation of Lox
    - Create a set and assign it to a variable: `var set = Set(element1, element2);`
        - The `Set` function takes in a variable number of arguments and uses them as the set elements: `Set(element1, element2, ..., elementN)`
    - Operations on sets, where `a` and `b` are sets:
        - Union: `a | b`
        - Intersection: `a & b`
        - Difference: `a - b`
        - Symmetric difference: `a ^ b`
        - Proper subset test: `a < b`
        - Subset test: `a <= b`
        - Proper superset test: `a > b`
        - Superset test: `a >= b`
    - The following cannot be used as set elements: buffer, dictionary, list, set
    - Besides these operations, sets also have some methods associated with them:
        - `set.add(element)`, which adds an element to the set if it is not already in the set. This method returns `true` if the element was successfully added, `false` if it was not, and throws a runtime error if the element is an object that cannot be a set element
        - `set.clear()`, which removes all elements from the set
        - `set.contains(element)`, which returns `true` if the specified element is in the set, `false` if it is not, and throws a runtime error if the element is an object that cannot be a set element
        - `set.copy()`, which returns a shallow copy of the original set as a new set
        - `set.isDisjoint(set2)`, which returns `true` if `set` and `set2` are disjoint, meaning they have no elements in common, and `false` otherwise
        - `set.isEmpty()`, which returns `true` if the set contains no elements and `false` otherwise
        - `set.remove(element)`, which removes the specified element from the set. Returns `true` if the set contained `element`, false if it didn't, and throws a runtime error if the element is an object that cannot be a set element
        - `set.toList()`, which returns a list of all the elements in the set in no particular order
- A range type is supported in this implementation of Lox
    - A range is a sequence of integers generated on demand, starting from a start value, stopping at but not including the stop value, and updating the current value using the step value
    - Examples of creating range objects and assigning them to variables:
        - `var x = range(5);`
        - `var x = range(1, 6);`
        - `var x = range(6, 1, -1);`
        - `var x = range(2, 12, 2);`
        - `var x = range(12, 2, -2);`
    - Get an integer from a range object by index, where `index` is an integer: `range[index]`
    - Get a new range object by slicing the original range object from indexes `start` to `end`, where `start` and `end` are integers: `range[start:end]`
    - Ranges are iterables and can be iterated over, yielding the generated integers. For example:
        - `range(5)` yields the integers [0, 1, 2, 3, 4]
        - `range(1, 6)` yields the integers [1, 2, 3, 4, 5]
        - `range(6, 1, -1)` yields the integers [6, 5, 4, 3, 2]
        - `range(2, 12, 2)` yields the integers [2, 4, 6, 8, 10]
        - `range(12, 2, -2)` yields the integers [12, 10, 8, 6, 4]
    - Unlike lists of integers, ranges always take up the same amount of memory no matter what the start, stop, and step values of the range are
    - Ranges have the following fields and methods associated with them:
        - `range.all(callback)`, which returns `true` if the callback function returns `true` for all generated integers in the range and `false` otherwise
        - `range.any(callback)` which returns `true` if the callback function returns `true` for any generated integer in the range and `false` otherwise
        - `range.contains(num)`, which returns `true` if `num` is in the range based on the start, stop, and step values and `false` otherwise
        - `range.filter(callback)`, which returns a list containing only the generated integers from the range where the callback function returns a truthy value for them
        - `range.forEach(callback)`, which executes the callback function for each generated integer in the range
        - `range.index(num)`, which returns the index value of `num` in the range or `-1` if `num` is not in the range
        - `range.map(callback)`, which returns a list with the results of calling a callback function on each generated integer from the range
        - `range.reduce(callback, [initialValue])`, which applies a reducer callback function on every generated integer from the range from start to stop based on the step value and returns a single value
        - `range.start`, which is the range object's start value as an integer
        - `range.step`, which is the range object's step value as an integer
        - `range.stop`, which is the range object's stop value as an integer
        - `range.sum()`, which returns the sum of all generated integers from the range as an integer
        - `range.toBuffer()`, which attempts to return a buffer with all generated integers from the range as buffer elements, throwing an error if an integer generated from the range is an invalid buffer value
        - `range.toList()`, which returns a list with all generated integers from the range as list elements
        - `range.toSet()`, which returns a set with all generated integers from the range as set elements
- A bigrange type is supported in this implementation of Lox
    - Bigranges are like ranges except their start, stop, and step values are bigints instead of integers
    - They are created using the `bigrange` function, which takes the same integer arguments as `range` as well as bigint arguments
    - They share the same methods as ranges
- Enums are supported in this implementation of Lox
    ```js
    enum Token {
        ADD,
        SUBTRACT,
        MULTIPLY,
        DIVIDE
    }
    var a = Token.ADD;
    var b = Token.ADD;
    var c = Token.SUBTRACT;
    print a; //Prints "Token.ADD"
    print type(a); //Prints "Token"
    print a == b; //Prints "true"
    print a == c; //Prints "false"
    ```
- The ability to import other Lox files is supported in this implementation of Lox
    - Syntax:
        ```js
        import "file-name";
        import "file-name" as alias;
        ```
    - The specified import file is executed and all variable, function, and class declarations declared globally in the imported file are brought into the global environment of the current file
    - If the specified import file doesn't exist or if the file exists but an error occurred while it was being executed, a runtime error is thrown
    - `import` statements can also have an optional alias specified, in which case only the alias name is brought into the global environment of the current file and all global variable, function, and class declarations from the imported file become properties of the alias and can be accessed using the following notation: `alias.variable`
- A few other native functions are defined:
    - `bigrange(stop)`, which takes in an integer or bigint and returns a bigrange object with a start value of `0n`, a stop value of `stop`, and a step value of `1n`
    - `bigrange(start, stop, [step])`, which takes in `start`, `stop`, and `step` as integers or bigints and returns a bigrange object with the specified parameters. If `step` is omitted, the resulting bigrange object will have a step value of `1n`
    - `bin(num)`, which converts the specified integer `num` into its binary representation as a string prefixed with "0b"
    - `Buffer(element1, element2, ..., elementN)`, which takes in a variable number of arguments and returns a buffer with the arguments as buffer elements. If an argument is not an integer or is an integer less than 0 or greater than 255, a runtime error is thrown
    - `BufferCap(capacity)`, which returns a new buffer of the specified capacity, which is the number of elements the buffer can store before having to internally resize the underlying array that stores the buffer elements when a new element is added
    - `BufferZero(length)`, which returns a new buffer of the specified length, where each initial element is `0`
    - `cap(item)`, which returns the capacity of a buffer or list, which is the number of elements the buffer or list can store before having to internally resize the underlying array that stores the buffer or list elements when a new element is added
    - `chr(i)`, which returns a string with a single character that is the Unicode character value of the code point `i`, where `i` is an integer
    - `eval(argument)`, which evaluates the string argument as Lox code and returns the result of the final expression in the evaluated code. If the argument is not a string, it is simply returned directly
        - **Warning**: `eval` is a dangerous function to use, as it can execute arbitrary Lox code and must be used with caution
    - `hex(num)`, which converts the specified integer `num` into its hexadecimal representation as a string prefixed with "0x"
    - `input([prompt])`, which writes the value of `prompt` to standard output if it is provided and reads a line from standard input as a string without a trailing newline and returns that string
        - Pressing Ctrl+C will throw a keyboard interrupt runtime error, and pressing Ctrl+D will cause this function to return `nil`
    - `iterator(iterable)`, which returns an iterator object from the specified iterable type and throws a runtime error if the argument is not an iterable type
        - Iterator objects have the following methods associated with them:
            - `iterator.hasNext()`, which returns `true` if there are more elements to be iterated over and `false` otherwise
            - `iterator.next()`, which returns the next element in the iterator
                - If the iterator has no more elements, calling this method will throw a runtime error with the error message `"StopIteration"`
            - `iterator.toList([length])`, which returns a list of elements from the iterator with the specified length as an integer. If `length` is omitted, the resulting list is obtained by repeatedly calling this iterator's `next` method until there are no more elements to be iterated over
        - Various utility iterator methods and fields are defined under a built-in class called `Iterator`, which is documented [here](./doc/Iterator.md)
    - `len(element)`, which returns the length of a buffer, dictionary, list, set, or string
        - Buffers: the length is the number of elements in the buffer
        - Dictionaries: the length is the number of keys in the dictionary
        - Lists: the length is the number of elements in the list
        - Ranges: the length is the number of integers in the range object based on its start, stop, and step values
        - Sets: the length is the number of elements in the set
        - Strings: the length is the number of characters in the string
    - `List(length)`, which returns a new list of the specified length, where each initial element is `nil`
    - `ListCap(capacity)`, which returns a new list of the specified capacity, which is the number of elements the list can store before having to internally resize the underlying array that stores the list elements when a new element is added
    - `ListZero(length)`, which returns a new list of the specified length, where each initial element is `0`
    - `oct(num)`, which converts the specified integer `num` into its octal representation as a string prefixed with "0o"
    - `ord(c)`, which returns an integer that represents the Unicode code point of the character `c`, where `c` is a string that contains a single Unicode character
    - `range(stop)`, which takes in an integer and returns a range object with a start value of `0`, a stop value of `stop`, and a step value of `1`
    - `range(start, stop, [step])`, which takes in `start`, `stop`, and `step` as integers and returns a range object with the specified parameters. If `step` is omitted, the resulting range object will have a step value of `1`
    - `Set(element1, element2, ..., elementN)`, which takes in a variable number of arguments and returns a set with the arguments as set elements with all duplicate elements removed. If an argument cannot be stored in a set, a runtime error is thrown
    - `sleep(duration)`, which pauses the program for the specified duration in seconds
    - `type(element)`, which returns a string representing the type of the element
- This Lox REPL supports typing in block statements with multiple lines
- Expressions such as `1 + 1` that are typed into the REPL are evaluated and their results are displayed, with no need for semicolons at the end
    - Assignment expressions still require semicolons when typed into the REPL as standalone expressions, like `x = 0;`, `object.property = value;`, and `list[index] = value;`

# Known bugs
See [knownbugs.md](./doc/knownbugs.md)

# Progress
- Chapter 4 - Scanning (Complete)
- Chapter 5 - Representing Code (Complete)
- Chapter 6 - Parsing Expressions (Complete)
- Chapter 7 - Evaluating Expressions (Complete)
- Chapter 8 - Statements and State (Complete)
- Chapter 9 - Control Flow (Complete)
- Chapter 10 - Functions (Complete)
- Chapter 11 - Resolving and Binding (Complete)
- Chapter 12 - Classes (Complete)
- Chapter 13 - Inheritance (Complete)

# License
This implementation of Lox is distributed under the terms of the [MIT License](https://github.com/AlanLuu/lox-go/blob/main/LICENSE).
