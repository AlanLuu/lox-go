## JSON methods

The following methods are defined in the built-in `JSON` class:
- `JSON.marshal(arg)`, which attempts to convert the specified argument into a JSON string representation using Go's `json.Marshal`, throwing a runtime error if the argument cannot be converted into one
- `JSON.parse(str)`, which attempts to parse a JSON string representation into a dictionary, throwing any parsing errors encountered during the parsing as runtime errors
- `JSON.stringify(arg)`, which attempts to convert the specified argument into a JSON string representation using a custom method, throwing a runtime error if the argument cannot be converted into one
- `JSON.valid(str)`, which returns `true` if the specified string is a valid JSON encoding and `false` otherwise
