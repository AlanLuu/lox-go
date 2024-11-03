# Web browser methods

The following methods are defined in the built-in `webbrowser` class:
- `webbrowser.browsers()`, which returns a list of strings that represent the names of browser commands this class uses to attempt to open a url
- `webbrowser.commands()`, which returns a list of lists of strings that represent the commands this class uses to attempt to open a url
- `webbrowser.mustOpen(url)`, which attempts to open the specified url, which is a string, in the default browser of the current system, throwing a runtime error if the browser was not opened successfully
- `webbrowser.open(url)`, which attempts to open the specified url, which is a string, in the default browser of the current system, returning `true` if the browser was opened successfully and `false` otherwise
- `webbrowser.other()`, which returns a list of strings that represent the names of other miscellaneous commands this class uses to attempt to open a url
