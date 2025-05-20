# Log class methods and fields

By default, the built-in `log` class logs to standard error.

The following log format fields are defined in the built-in `log` class, which are all integers corresponding to flags that define log prefixes:
- `log.Ldate`, `log.Ltime`, `log.Lmicroseconds`, `log.Llongfile`, `log.Lshortfile`, `log.LUTC`, `log.Lmsgprefix`, `log.LstdFlags`

An optional prefix string can be set for the `log` class and logger objects.

The following methods are defined in the built-in `log` class:
- `log.fatal(...args)`, which logs the specified arguments to the output file associated with the `log` class and exits the program with a status code of 1
- `log.flags()`, which returns an integer that represents the flags of the `log` class
- `log.logger()`, which returns the default logger object associated with the `log` class
- `log.logger(prefix, flag)`, which returns a new logger object with the specified prefix string and flag integer
- `log.logger(file, prefix, flag)`, which returns a new logger object with the specified prefix string and flag integer that logs to the specified file object
- `log.loggerFromDefault()`, which returns a new logger object with its output file, prefix, and flag copied from the default logger object associated with the `log` class
- `log.outputIs(file)`, which returns `true` if the file that the `log` class logs to is the same as the specified file argument and `false` otherwise
- `log.prefix()`, which returns the optional prefix string associated with the `log` class as a string
- `log.print(...args)`, which logs the specified arguments to the output file associated with the `log` class
- `log.println(...args)`, which logs the specified arguments to the output file associated with the `log` class
- `log.setFlags(flag)`, which sets the flag integer of the `log` class to the specified flag integer
- `log.setOutput(file)`, which sets the output file associated with the `log` class to the specified file object
- `log.setPrefix(prefix)`, which sets the optional prefix string of the `log` class to the specified prefix string
- `log.sprint(...args)`, which returns the log line generated from the specified arguments as a string rather than logging it to the output file associated with the `log` class
- `log.sprintln(...args)`, which returns the log line generated from the specified arguments plus a newline character as a string rather than logging it to the output file associated with the `log` class

Logger objects have the following methods associated with them:
- `logger.clearSavedLogs()`, which removes all saved log lines in the current logger object
- `logger.fatal(...args)`, which logs the specified arguments to the output file associated with the current logger object and exits the program with a status code of 1
- `logger.flags()`, which returns an integer that represents the flags of the current logger object
- `logger.getSaved(index)`, which returns the saved log line from the underlying list in the current logger object at the specified index integer
    - If `index` is out of range, this method throws a runtime error
- `logger.outputIs(file)`, which returns `true` if the file that the current logger object logs to is the same as the specified file argument and `false` otherwise
- `logger.prefix()`, which returns the optional prefix string associated with the current logger object as a string
- `logger.print(...args)`, which logs the specified arguments to the output file associated with the current logger object
- `logger.println(...args)`, which logs the specified arguments to the output file associated with the current logger object
- `logger.printSave(...args)`, which saves the log line generated from the specified arguments as a string in a list in the current logger object rather than logging it to the output file associated with the current logger object
- `logger.printSaved()`, which logs all saved log lines in the current logger object to the output file associated with the current logger object
- `logger.savedGet(index)`, which is an alias for `logger.getSaved`
- `logger.savedLen()`, which returns the number of saved log lines in the current logger object as an integer
- `logger.savedLogs()`, which returns a list of strings of the saved log lines in the current logger object
- `logger.setFlags(flag)`, which sets the flag integer of the current logger object to the specified flag integer
- `logger.setOutput(file)`, which sets the output file associated with the current logger object to the specified file object
- `logger.setPrefix(prefix)`, which sets the optional prefix string of the current logger object to the specified prefix string
- `logger.sprint(...args)`, which returns the log line generated from the specified arguments as a string rather than logging it to the output file associated with the current logger object
- `logger.sprintln(...args)`, which returns the log line generated from the specified arguments plus a newline character as a string rather than logging it to the output file associated with the current logger object
