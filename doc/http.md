# HTTP methods and fields

Any method that fails will throw a runtime error with a message describing the error.

The following methods are defined in the built-in `http` class:
- `http.get(url, [headers])`, which sends an HTTP GET request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.handler(callback)`, which returns an HTTP handler object from the specified callback function
- `http.head(url, [headers])`, which sends an HTTP HEAD request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.post(url, [headers])`, which sends an HTTP POST request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.postBin(url, data, [headers])`, which sends an HTTP POST request to the specified URL along with the binary data specified as a buffer and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - The binary data in the buffer is sent with a `Content-Type` of `application/octet-stream` if it is nonempty
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.postForm(url, form, [headers])`, which sends an HTTP POST request to the specified URL along with the form data specified as a dictionary and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - Form data is sent with a `Content-Type` of `application/x-www-form-urlencoded` if it is nonempty
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
- `http.req([url])`, which returns an HTTP request object with the specified URL string or URL object, or an empty URL if `url` is omitted
- `http.request(method, url, body, [headers])`, which sends an HTTP request with the specified method string to the specified URL along with the body parameter. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - The body parameter can be one of the following types:
        - Buffer, which is sent with a `Content-Type` of `application/octet-stream` if it is nonempty
        - Dictionary, which is sent with a `Content-Type` of `application/json`
            - The dictionary's keys must only be strings and its values must be valid JSON values or else a runtime error is thrown
            - `JSON.stringify` is used to convert the dictionary into a JSON string, so a runtime error is thrown if that method is missing
        - String, which is sent with a `Content-Type` of `text/plain` if it is nonempty
        - `nil`, in which case an empty body is sent with the request
        - If the method is equal to `GET` or `HEAD`, the body parameter must be `nil` or else a runtime error is thrown
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.requestForm(method, url, form, [headers])`, which sends an HTTP request with the specified method string to the specified URL along with the form data specified as a dictionary and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - This method does not support `GET` or `HEAD` requests and throws a runtime error if one of those request methods is specified in this method
    - Form data is sent with a `Content-Type` of `application/x-www-form-urlencoded` if it is nonempty
    - The form dictionary's keys must only be strings and its values must either be strings or lists or else a runtime error is thrown
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.serve([path], port)`, which starts an HTTP server that serves all files and directories in the specified directory path on the specified port number. If the path is omitted, the current working directory's path is used as the path to serve
    - On success, this method blocks until it is interrupted using Ctrl+C, in which case the server is shut down and a runtime error is thrown
- `http.srvTest(handler/callback)`, which returns an HTTP test server instance from the specified HTTP handler object or callback function
- `http.srvTestUnstarted(handler/callback)`, which returns an unstarted HTTP test server instance from the specified HTTP handler object or callback function

HTTP request objects have the following methods associated with them:
- `request.body(buffer/dict/string/urlvalues)`
- `request.bodyClear()`
- `request.contentLength(length)`
- `request.contentType(type)`
- `request.cookieClear()`
- `request.cookieJar()`
- `request.cookieKV(key, value)`
- `request.form(buffer/dict/string/urlvalues)`
- `request.formData(buffer/dict/string/urlvalues)`
- `request.headerAdd(key, value)`
- `request.headerClear()`
- `request.headerDel(key)`
- `request.headerSet(key, value)`
- `request.json(buffer/dict/string/urlvalues)`
- `request.method(methodStr)`
- `request.methodCONNECT()`
- `request.methodDELETE()`
- `request.methodGET()`
- `request.methodHEAD()`
- `request.methodOPTIONS()`
- `request.methodPATCH()`
- `request.methodPOST()`
- `request.methodPUT()`
- `request.methodTRACE()`
- `request.redirects(numRedirects)`
- `request.send()`
- `request.sendRet()`
- `request.sendThreads(numThreads)`
- `request.timeout(duration)`
- `request.timeoutClear()`
- `request.url(string/url)`
- `request.userAgent(userAgentStr)`

HTTP response objects have the following methods and fields associated with them:
- `response.close()`, which closes the underlying response content stream, preventing access to `response.raw` and `response.text` if any of them haven't been accessed before from the caller before closing the response
- `response.cookies`, which is a list of the cookies sent from the response in the Set-Cookie header as HTTP cookie objects
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

HTTP cookie objects have the following methods and fields associated with them:
- `cookie.domain`, which is a string representing the `Domain` attribute of the cookie
- `cookie.expires`, which is a date representing the expiration date of the cookie
- `cookie.httpOnly`, which is a boolean that is `true` if the cookie is HTTP only and `false` otherwise
- `cookie.maxAge`, which is an integer representing the `Max-Age` attribute of the cookie
- `cookie.name`, which is a string representing the name of the cookie
- `cookie.path`, which is a string representing the path of the cookie
- `cookie.printSimpleString()`, which prints the cookie's name followed by an `=` character followed by the cookie's value
- `cookie.printString()`, which prints the string representation of the cookie
- `cookie.raw`, which is a string that is the raw representation of the cookie
- `cookie.rawExpires`, which is a string that represents the raw expires value of the cookie
- `cookie.secure`, which is a boolean that is `true` if the cookie is secure and `false` otherwise
- `cookie.simpleString`, which is a string representing the cookie's name followed by an `=` character followed by the cookie's value
- `cookie.string()`, which returns the string representation of the cookie
- `cookie.unparsed`, which is a list of strings of the unparsed cookie attribute pairs
- `cookie.valid()`, which returns `true` if the cookie is valid and `false` otherwise
- `cookie.validErr()`, which returns an error object representing the error if the cookie is not valid
- `cookie.validThrowErr()`, which returns `nil` if the cookie is valid and throws a runtime error otherwise

HTTP handler objects have the following methods associated with them:
- `http handler.delAllIgnoredEndPaths()`
- `http handler.delIgnoredEndPath(path)`
- `http handler.hasIgnoredEndPath(path)`
- `http handler.ignoredEndPaths()`
- `http handler.ignoredEndPathsList()`
- `http handler.ignoreEndPath(path)`
- `http handler.ignoreFaviconPath()`
- `http handler.logging(loggingEnabled)`
- `http handler.wouldIgnorePath(path)`

HTTP test servers have the following methods and fields associated with them:
- `http test server.close()`
- `http test server.isClosed()`
- `http test server.isStarted()`
- `http test server.start()`
- `http test server.url`
- `http test server.wait()`
- `http test server.waitForever()`
