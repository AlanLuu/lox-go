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
- The following additional operations are supported in this implementation of Lox:
    - `a % b`, which returns the remainder of two numbers `a` and `b`
    - `a ** b`, which returns `a` raised to the power of `b`, where `a` and `b` are numbers
        - The exponent operator has higher precedence than any unary operators on the left, so `-a ** b` is equivalent to `-(a ** b)`.
    - `a << b` and `a >> b`, which returns a number representing the number `a` shifted by `b` bits to the left and right respectively.
        - If `a` or `b` are floats, they are converted into integers before the shift operation
    - `~a`, which returns the bitwise NOT of the number `a`
    - `a & b`, `a | b`, and `a ^ b`, which returns the bitwise AND, OR, and XOR of two numbers `a` and `b` respectively.
        - If `a` or `b` are floats, they are converted into integers before the bitwise operation
        - Unlike in C, the precedence of the bitwise operators is higher than the precedence of the comparison operators, so `a & b == value` is equivalent to `(a & b) == value`
- Division by 0 results in `Infinity`, which uses Golang's `math.Inf()` under the hood
- Performing a binary operation that isn't supported between two types results in `NaN`, which stands for "not-a-number", using Golang's `math.NaN()` under the hood
- Booleans and `nil` are treated as integers when performing arithmetic operations on them, with `true` and `false` being treated as `1` and `0` respectively, and `nil` being treated as `0`
- Besides `false` and `nil`, the values `0`, `0.0`, `NaN`, `""`, `[]`, and `{}` are also falsy values
- Binary, hexadecimal, and octal integer literals are supported in this implementation of Lox
    - Binary literals start with the prefix `0b`
    - Hexadecimal literals start with the prefix `0x`
    - Octal literals start with the prefixes `0o` or `0`
    - All letters in prefixes are case-insensitive
- Number literals support the following features:
    - An underscore character can be used to group digits, such as `1_000_000`, which is equivalent to `1000000`
        - Underscore characters are also allowed in binary, hexadecimal, and octal literals, except for octal literals starting with a `0`
- `break` and `continue` statements are supported in this implementation of Lox
- For loops are implemented with their own AST node instead of being desugared into while loop nodes
    - This makes it easier to implement the `continue` statement inside for loops
- Variables declared in the initialization part of a for loop are locally scoped to that loop and do not become global variables
- Anonymous function expressions are supported in this implementation of Lox. There are two forms supported:
    - `fun(param1, paramN) {<statements>}`, which is a traditional anonymous function expression that contains a block with statements
    - `fun(param1, paramN) => <expression>`, which is an arrow function expression that implicitly returns the given expression when called
    - The parser will attempt to parse anonymous function expressions that appear on their own line as function declarations, throwing a parser error as a result. This is expected behavior; to force the parser to parse them as expressions, wrap the function expression inside parentheses, like `(fun() {})()`. In this case, this creates an anonymous function expression that is called immediately
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
- Various mathematical functions and constants are defined under a built-in class called `Math`, which is documented [here](./doc/Math.md)
- Strings have some additional features associated with them:
    - Strings can be represented using single quotes as well
    - Strings can be indexed by an integer, which will return a new string with only the character at the specified index: `string[index]`
    - Get a new string with all characters from indexes `start` to `end` exclusive, where `start < end`: `string[start:end]`
        - If `start >= end`, a new empty string is returned
        - `start` or `end` can be omitted, in which case the starting index will have a value of `0` if `start` is omitted and the ending index will have a value of `len(string)` if `end` is omitted
    - It is a runtime error to use an index value that is less than 0 or greater than or equal to the length of the string to index into that string
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
        - `string.compare(string2)`, which lexicographically compares `string` and `string2` and returns `0` if `string == string2`, `-1` if `string < string2`, and `1` if `string > string2`
        - `string.contains(substr)`, which returns `true` if `substr` is contained within `string` and `false` otherwise
        - `string.endsWith(suffix)`, which returns `true` if `string` ends with `suffix` and `false` otherwise
        - `string.index(string2)`, which returns an integer representing the index value of the location of `string2` in `string`, or `-1` if `string2` is not in `string`
        - `string.lastIndex(string2)`, which returns an integer representing the index value of the last occurrence of `string2` in `string`, or `-1` if `string2` is not in `string`
        - `string.lower()`, which returns a new string with all lowercase letters
        - `string.padEnd(length, padStr)`, which pads the contents of `padStr` to the end of `string` until the new string is of length `length`
        - `string.padStart(length, padStr)`, which pads the contents of `padStr` to the beginning of `string` until the new string is of length `length`
        - `string.replace(oldStr, newStr)`, which returns a new string where all occurrences of `oldStr` in the original string are replaced with `newStr`
        - `string.split(delimiter)`, which returns a list containing all substrings that are separated by `delimiter`
        - `string.startsWith(prefix)`, which returns `true` if `string` begins with `prefix` and `false` otherwise
        - `string.strip([chars])`, which returns a new string with all leading and trailing characters from `chars` removed. If `chars` is omitted, this method returns a new string with all leading and trailing whitespace removed
        - `string.toNum([base])`, which attempts to convert `string` into an integer or float and returns that value if successful and `NaN` otherwise. If `base` is specified, then this method will attempt to convert `string` that is represented as the specified base into an integer or float and returns that value if the conversion was successful and `NaN` otherwise
        - `string.upper()`, which returns a new string with all uppercase letters
        - `string.zfill(length)`, which returns a new string where the character `'0'` is padded to the left until the new string is of length `length`. If a leading `'+'` or `'-'` sign is part of the original string, the `'0'` padding is inserted after the leading sign instead of before
- Lists are supported in this implementation of Lox
    - Create a list and assign it to a variable: `var list = [1, 2, 3];`
    - Get an element from a list by index: `list[index]`
    - Get a new list with all elements from indexes `start` to `end` exclusive, where `start < end`: `list[start:end]`
        - If `start >= end`, a new empty list is returned
        - `start` or `end` can be omitted, in which case the starting index will have a value of `0` if `start` is omitted and the ending index will have a value of `len(list)` if `end` is omitted
    - Set an element: `list[index] = value;`
    - It is a runtime error to use an index value that is less than 0 or greater than or equal to the length of the list to get or set
    - Concatenate two lists together into a new list: `list + list2`
    - Get a new list with all elements from the original list repeated `n` times, where `n` is an integer: `list * n`
    - Besides these operations, lists also have some methods associated with them:
        - `list.all(callback)`, which returns `true` if the callback function returns `true` for all elements in the list and `false` otherwise
        - `list.any(callback)` which returns `true` if the callback function returns `true` for any element in the list and `false` otherwise
        - `list.append(element)`, which appends an element to the end of the list
        - `list.clear()`, which removes all elements from the list
        - `list.contains(element)`, which returns `true` if `element` is contained in the list and `false` otherwise
        - `list.count(element)`, which returns the number of times `element` appears in the list
        - `list.extend(list2)`, which appends every element from `list2` to the end of the list
        - `list.filter(callback)`, which returns a new list containing only the elements from the original list where the callback function returns a truthy value for them
        - `list.find(callback)`, which returns the first element in the list where the callback function returns `true`, or `nil` if the callback returns `false` for every element in the list
        - `list.findIndex(callback)`, which returns the index of the first element in the list where the callback function returns `true`, or `-1` if the callback returns `false` for every element in the list
        - `list.flatten()`, which returns a new list where all elements contained within nested lists are flattened into a list without any nested lists
        - `list.forEach(callback)`, which executes the callback function for each element in the list
        - `list.index(element)`, which returns the index value of the element's position in the list, or `-1` if the element is not in the list
        - `list.lastIndex(element)`, which returns the index value of the last occurrence of the element in the list, or `-1` if the element is not in the list
        - `list.insert(index, element)`, which inserts an element into the list at the specified index
        - `list.join(separator)`, which concatenates all elements in the list into a string where each element is separated by a separator string
        - `list.map(callback)`, which returns a new list with the results of calling a callback function on each element of the original list
        - `list.pop([index])`, which removes and returns the element at the specified index from the list. If `index` is omitted, this method removes and returns the last element from the list
        - `list.reduce(callback, [initialValue])`, which applies a reducer callback function on every element in the list from left to right and returns a single value
        - `list.remove(element)`, which removes the first occurrence of `element` from the list. Returns `true` if the list contained `element` and `false` otherwise
        - `list.shuffle()`, which shuffles all elements in the list in place
    - Two lists are compared based on whether they are the same length and for every index `i`, the element from the first list at index `i` is equal to the element from the second list at index `i`
    - Attempting to use an index value larger than the length of the list will cause a runtime error
- Dictionaries are supported in this implementation of Lox
    - Create a dictionary and assign it to a variable: `var dict = {"key": "value"};`
    - Get an element from a dictionary by key: `dict[key]`
        - It is a runtime error to attempt to get an element using a key that is not in the dictionary
    - Set an element: `dict[key] = value;`
    - Merge two dictionaries together: `dict | dict2`
        - If a key exists in both `dict` and `dict2`, the key in the merged dictionary becomes associated with the value from `dict2`
    - The following cannot be used as dictionary keys: dictionary, list
    - Besides these operations, dictionaries also have some methods associated with them:
        - `dictionary.clear()`, which removes all keys from the dictionary
        - `dictionary.containsKey(key)`, which returns `true` if the specified key exists in the dictionary and `false` otherwise
        - `dictionary.copy()`, which returns a shallow copy of the original dictionary as a new dictionary
        - `dictionary.get(key, [defaultValue])`, which returns the value associated with the specified key from the dictionary, or `defaultValue` if the key doesn't exist in the dictionary and `defaultValue` is provided, or `nil` otherwise
        - `dictionary.keys()`, which returns a list of all the keys in the dictionary in no particular order
        - `dictionary.removeKey(key)`, which removes the specified key from the dictionary and returns the value originally associated with the key or `nil` if the key doesn't exist in the dictionary. Note that a return value of `nil` can also mean that the specified key had a value of `nil`
        - `dictionary.values()`, which returns a list of all the values in the dictionary in no particular order
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
- A few other native functions are defined:
    - `chr(i)`, which returns a string with a single character that is the Unicode character value of the code point `i`, where `i` is an integer
    - `input([prompt])`, which writes the value of `prompt` to standard output if it is provided and reads a line from standard input as a string without a trailing newline and returns that string
    - `len(element)`, which returns the length of a dictionary, list, or string
        - Dictionaries: the length is the number of keys in the dictionary
        - Lists: the length is the number of elements in the list
        - Strings: the length is the number of characters in the string
    - `List(length)`, which returns a new list of the specified length, where each initial element is `nil`
    - `ord(c)`, which returns an integer that represents the Unicode code point of the character `c`, where `c` is a string that contains a single Unicode character
    - `sleep(duration)`, which pauses the program for the specified duration in seconds
    - `type(element)`, which returns a string representing the type of the element
- This Lox REPL supports typing in block statements with multiple lines
- Expressions such as `1 + 1` that are typed into the REPL are evaluated and their results are displayed, with no need for semicolons at the end
    - Assignment expressions still require semicolons when typed into the REPL as standalone expressions, like `x = 0;`, `object.property = value;`, and `list[index] = value;`

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
