# Regex methods and fields

The following methods are defined in the built-in `regex` class:
- `regex class.compile(pattern)`, which returns a compiled regex instance with the specified regex pattern string, throwing a runtime error if the pattern is not a valid regular expression
- `regex class.escape(string)`, which returns a new string with all regular expression characters in the original string escaped
- `regex class.test(pattern, string)`, which returns `true` if the specified string matches the regex pattern string and `false` otherwise
    - If the regex pattern if not a valid regular expression, a runtime error is thrown

Compiled regexes have the following methods and fields associated with them:
- `regex.findAll(string)`, which returns a list of all matches in the specified string according to the regex pattern of the compiled regex instance
    - If there are no matches, the returned list is empty
- `regex.findAllGroups(string)`, which returns a list of lists that contain a matched string and any subexpressions matched along with it according to the regex pattern of the compiled regex instance
    - If there are no matches, the returned list is empty
- `regex.numSubexp`, which is the number of parenthesized subexpressions in the regex pattern of the compiled regex instance as an integer
- `regex.pattern`, which is the regex pattern of the compiled regex instance as a string
- `regex.replace(string, replacement)`, which returns a new string with all matches from the regex pattern of the compiled regex instance in the specified string replaced with the replacement string
- `regex.replacen(string, replacement)`, which returns a list with two elements: the first being a new string with all matches from the regex pattern of the compiled regex instance in the specified string replaced with the replacement string, and the second being the number of replacements made as an integer
- `regex.split(string)`, which returns a list containing all substrings that are separated by the regex pattern of the compiled regex instance
- `regex.test(string)`, which returns `true` if the specified string matches the regex pattern of the compiled regex instance and `false` otherwise
