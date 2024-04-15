# lox-go
A Golang implementation of the Lox language from the book [Crafting Interpreters](https://craftinginterpreters.com/) with my own features

# Usage
Supports running code from a file:
```
lox code.txt
```
or from stdin:
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
- Booleans are treated as numbers when performing arithmetic operations on them
- Strings can be multiplied by a whole number `n`, which returns a copy of the string repeated `n` times

# Progress
- Chapter 4 - Scanning (Complete)
- Chapter 5 - Representing Code (Complete)
- Chapter 6 - Parsing Expressions (Complete)
- Chapter 7 - Evaluating Expressions (Complete)
- Chapter 8 - Statements and State (TODO)
- Chapter 9 - Control Flow (TODO)
- Chapter 10 - Functions (TODO)
- Chapter 11 - Resolving and Binding (TODO)
- Chapter 12 - Classes (TODO)
- Chapter 13 - Inheritance (TODO)

# License
This implementation of Lox is distributed under the terms of the [MIT License](https://github.com/AlanLuu/lox-go/blob/main/LICENSE).
