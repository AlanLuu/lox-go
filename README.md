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
```
git clone https://github.com/AlanLuu/lox-go.git
cd lox-go
go build
```

# Differences from Lox
- Concatenating a string with another data type will convert that type into a string and concatenate them together
- The following additional binary operations are supported in this implementation of Lox:
    - `a % b`, which returns the remainder of two numbers `a` and `b`
- Division by 0 results in `Infinity`, which uses Golang's `math.Inf()` under the hood
- Performing a binary operation that isn't supported between two types results in `NaN`, which stands for "not-a-number", using Golang's `math.NaN()` under the hood
- Booleans and `nil` are treated as numbers when performing arithmetic operations on them, with `true` and `false` being treated as `1` and `0` respectively, and `nil` being treated as `0`
- Strings can be multiplied by a whole number `n`, which returns a copy of the string repeated `n` times
- Besides `false` and `nil`, the values `0`, `0.0`, `NaN`, and `""` are also falsy values
- `break` and `continue` statements are supported in this implementation of Lox
- For loops are implemented with their own AST node instead of being desugared into while loop nodes
    - This makes it easier to implement the `continue` statement inside for loops
- Variables declared in the initialization part of a for loop are locally scoped to that loop and do not become global variables
- Anonymous function expressions are supported in this implementation of Lox. There are two forms supported:
    - `fun(param1, paramN) {<statements>}`, which is a traditional anonymous function expression that contains a block with statements
    - `fun(param1, paramN) => <expression>`, which is an arrow function expression that implicitly returns the given expression when called
    - The parser will attempt to parse anonymous function expressions that appear on their own line as function declarations, throwing a parser error as a result. This is expected behavior; to force the parser to parse them as expressions, wrap the function expression inside parentheses, like `(fun() {})()`. In this case, this creates an anonymous function expression that is called immediately
- Lists are supported in this implementation of Lox
    - Create a list and assign it to a variable: `var list = [1, 2, 3];`
    - Get an element from a list by index: `list[index]`
    - Set an element: `list[index] = value;`
    - Besides these operations, lists also have some methods associated with them:
        - `list.append(element)`, which appends an element to the end of the list
        - `list.clear()`, which removes all elements from the list
        - `list.contains(element)`, which returns `true` if `element` is contained in the list, and `false` otherwise
        - `list.count(element)`, which returns the number of times `element` appears in the list
        - `list.extend(list2)`, which appends every element from `list2` to the end of the list
        - `list.index(element)`, which returns the index value of the element's position in the list, or `-1` if the element is not in the list
        - `list.insert(index, element)`, which inserts an element into the list at the specified index
        - `list.pop([index])`, which removes and returns the element at the specified index from the list. If `index` is omitted, this method removes and returns the last element from the list
    - Two lists are compared based on whether they are the same length and for every index `i`, the element from the first list at index `i` is equal to the element from the second list at index `i`
    - Attempting to use an index value larger than the length of the list will cause a runtime error
- A few other native functions are defined:
    - `len(element)`, which returns the length of a string or list element
    - `List(length)`, which returns a new list of the specified length, where each initial element is `nil`
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
