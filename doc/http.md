# HTTP methods and fields

Any method that fails will throw a runtime error with a message describing the error.

The following methods are defined in the built-in `http` class:
- `http.get(url, [headers])`, which sends an HTTP GET request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.head(url, [headers])`, which sends an HTTP HEAD request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.post(url, [headers])`, which sends an HTTP POST request to the specified URL along with any HTTP headers in the headers dictionary if specified and returns an HTTP response object
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown
- `http.postForm(url, form, [headers])`, which sends an HTTP POST request to the specified URL along with the form data specified as a dictionary and returns an HTTP response object. If the headers dictionary is specified, all headers in the dictionary are sent with the request
    - Form data is sent with a `Content-Type` of `application/x-www-form-urlencoded`
    - The form dictionary's keys must only be strings and its values must either be strings or lists or else a runtime error is thrown
    - The headers dictionary must be empty or only contain strings or else a runtime error is thrown

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
