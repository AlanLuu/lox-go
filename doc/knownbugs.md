# Known bugs

- In REPL mode, there is an issue where printed expressions without a newline at the end are not printed at all
- The `eval` function is buggy when used inside block statements
- Runtime errors that get thrown in functions defined in files from the `loxcode` directory show a line number that is the internal line number of where the error was thrown in the file from the `loxcode` directory instead of the line number of where the function was actually called
