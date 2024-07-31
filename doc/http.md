# HTTP methods and fields

Any method that fails will throw a runtime error with a message describing the error.

The following methods are defined in the built-in `http` class:
- `http.get(url, [headers])`, which sends an HTTP GET request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.head(url, [headers])`, which sends an HTTP HEAD request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.post(url, [headers])`, which sends an HTTP POST request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.postBin(url, data, [headers])`, which sends an HTTP POST request to the specified URL along with the binary data specified as a buffer and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - The binary data in the buffer is sent with a `Content-Type` of `application/octet-stream` if it is nonempty
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.postForm(url, form, [headers])`, which sends an HTTP POST request to the specified URL along with the form data specified as a dictionary and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - Form data is sent with a `Content-Type` of `application/x-www-form-urlencoded`
    - The form dictionary's keys must only be strings and its values must either be strings or lists or else a runtime error is thrown
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.postJSON(url, json, [headers])`, which sends an HTTP POST request to the specified URL along with the JSON data specified as a string or dictionary and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - JSON data is sent with a `Content-Type` of `application/json`
    - If `json` is a dictionary, the JSON dictionary's keys must only be strings and its values must be valid JSON values or else a runtime error is thrown
        - This method utilizes `JSON.stringify` to convert the JSON dictionary into a string, so a runtime error is thrown if that method is missing
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.postText(url, text, [headers])`, which sends an HTTP POST request to the specified URL along with the body text specified as a string and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - The body text is sent with a `Content-Type` of `text/plain` if it is nonempty
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.request(method, url, body, [headers])`, which sends an HTTP request with the specified method as a string to the specified URL along with the body parameter. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - The body parameter can be one of the following types:
        - Buffer, which is sent with a `Content-Type` of `application/octet-stream` if it is nonempty
        - Dictionary, which is sent with a `Content-Type` of `application/json`
            - The dictionary's keys must only be strings and its values must be valid JSON values or else a runtime error is thrown
            - `JSON.stringify` is used to convert the dictionary into a JSON string, so a runtime error is thrown if that method is missing
        - String, which is sent with a `Content-Type` of `text/plain` if it is nonempty
        - If the method is equal to `GET` or `HEAD`, the body parameter must be `nil` or else a runtime error is thrown
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.serve([path], port)`, which starts an HTTP server that serves all files and directories in the specified directory path on the specified port number. If the path is omitted, the current working directory's path is used as the path to serve
    - On success, this method blocks until it is interrupted using Ctrl+C, in which case the server is shut down and a runtime error is thrown

HTTP response objects have the following methods and fields associated with them:
- `response.close()`, which closes the underlying response content stream, preventing access to `response.raw` and `response.text` if any of them haven't been accessed before from the caller before closing the response
- `response.elapsed`, which is a float that represents the amount of time the HTTP request took in seconds
- `response.headers`, which is a dictionary of all the HTTP headers sent from the server
- `response.raw`, which is a buffer containing the raw bytes of the response content
    - This field isn't stored in memory until it is accessed from the caller
    - When this field is accessed, if the field `response.text` isn't already in memory, it becomes stored in memory along with this field
- `response.status`, which is the HTTP status code as an integer
- `response.text`, which is the response content as a string
    - This field isn't stored in memory until it is accessed from the caller
    - When this field is accessed, if the field `response.raw` isn't already in memory, it becomes stored in memory along with this field
- `response.url`, which is the URL of the HTTP request as a string
