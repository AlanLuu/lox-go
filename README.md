# lox-go
A Golang implementation of the Lox language from the book [Crafting Interpreters](https://craftinginterpreters.com/) with my own features

# Usage
Supports running code from a file:
```
lox code.txt
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
- Division by 0 results in `Infinity`, which uses Golang's `math.Inf()` under the hood
- Performing a binary operation that isn't supported between two types results in `NaN`, which stands for "not-a-number", using Golang's `math.NaN()` under the hood
- Booleans are treated as numbers when performing arithmetic operations on them
- Strings can be multiplied by a whole number `n`, which returns a copy of the string repeated `n` times
- Besides `false` and `nil`, the values `0`, `0.0`, `NaN`, and `""` are also falsy values
- `break` and `continue` statements are supported in this implementation of Lox
- For loops are implemented with their own AST node instead of being desugared into while loop nodes
    - This makes it easier to implement the `continue` statement inside for loops
- This Lox REPL supports typing in block statements with multiple lines
- Expressions such as `1 + 1` that are typed into the REPL are evaluated and their results are displayed, with no need for semicolons at the end
    - Assignment expressions still require semicolons when typed into the REPL as standalone expressions, like `x = 0;`

# Progress
- Chapter 4 - Scanning (Complete)
- Chapter 5 - Representing Code (Complete)
- Chapter 6 - Parsing Expressions (Complete)
- Chapter 7 - Evaluating Expressions (Complete)
- Chapter 8 - Statements and State (Complete)
- Chapter 9 - Control Flow (Complete)
- Chapter 10 - Functions (TODO)
- Chapter 11 - Resolving and Binding (TODO)
- Chapter 12 - Classes (TODO)
- Chapter 13 - Inheritance (TODO)

# License
This implementation of Lox is distributed under the terms of the [MIT License](https://github.com/AlanLuu/lox-go/blob/main/LICENSE).
