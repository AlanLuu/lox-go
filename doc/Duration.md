# Duration class methods and fields

The following fields are defined in the built-in `Duration` class, which are all duration objects:
- `Duration.zero`, `Duration.nanosecond`, `Duration.microsecond`, `Duration.millisecond`, `Duration.second`, `Duration.minute`, `Duration.hour`, `Duration.day`, `Duration.year`

The following methods are defined in the built-in `Duration` class:
- `Duration.days(days)`, which returns a duration object of the specified number of days as an integer
- `Duration.fromInt(integer)`, which returns a duration object from the specified integer, where the integer is in nanoseconds
- `Duration.hours(hours)`, which returns a duration object of the specified number of hours as an integer
- `Duration.loop(duration, callback)`, which takes in a duration object and a callback function and repeatedly invokes the callback for the specified duration
- `Duration.microseconds(microseconds)`, which returns a duration object of the specified number of microseconds as an integer
- `Duration.milliseconds(milliseconds)`, which returns a duration object of the specified number of milliseconds as an integer
- `Duration.minutes(minutes)`, which returns a duration object of the specified number of minutes as an integer
- `Duration.parse(str)`, which parses the specified duration string and returns a duration object. If parsing is unsuccessful, a runtime error is thrown
- `Duration.since(date)`, which returns a duration object that represents the time elapsed since the specified date object
- `Duration.seconds(seconds)`, which returns a duration object of the specified number of seconds as an integer
- `Duration.sleep(duration)`, which takes in a duration object and pauses the program for the specified duration
- `Duration.stopwatch()`, which returns a new stopwatch instance
- `Duration.until(date)`, which returns a duration object that represents the time until the specified date object
- `Duration.years(years)`, which returns a duration object of the specified number of years as an integer

Duration objects have the following methods associated with them:
- `duration.abs()`, which returns a new duration object that is the absolute value of the current duration object
- `duration.add(duration2)`, which returns a new duration object that is the specified duration object added to the current duration object
- `duration.div(duration2)`, which returns a new duration object that is the current duration object divided by the specified duration object
    - If the specified duration object is zero, a runtime error is thrown
- `duration.hours()`, which returns the number of hours associated with the current duration object as a float
- `duration.int()`, which returns the integer value of the current duration object
- `duration.loop(callback)`, which takes in a callback function and repeatedly invokes the callback for the duration of the current duration object
- `duration.microseconds()`, which returns the number of microseconds associated with the current duration object as an integer
- `duration.milliseconds()`, which returns the number of milliseconds associated with the current duration object as an integer
- `duration.minutes()`, which returns the number of minutes associated with the current duration object as a float
- `duration.mod(duration2)`, which returns a new duration object that is the remainder of the current duration object divided by the specified duration object
    - If the specified duration object is zero, a runtime error is thrown
- `duration.mul(duration2)`, which returns a new duration object that is the specified duration object multiplied by to the current duration object
- `duration.nanoseconds()`, which returns the number of nanoseconds associated with the current duration object as an integer
- `duration.round(duration2)`, which returns a new duration object that is the current duration object rounded to the nearest multiple of the specified duration object
- `duration.scale(factor)`, which takes in an integer and returns a new duration object that is the current duration object multiplied by the specified integer
- `duration.seconds()`, which returns the number of seconds associated with the current duration object as a float
- `duration.sleep()`, which pauses the program for the duration of the current duration object
- `duration.string()`, which returns a string representation of the current duration object
- `duration.sub(duration2)`, which returns a new duration object that is the current duration object subtracted by the specified duration object
- `duration.times(factor)`, which is an alias for `duration.scale`
- `duration.truncate()`, which returns a new duration object that is the current duration object truncated towards zero to the nearest multiple of the specified duration object

Stopwatch instances have the following methods associated with them:
- `stopwatch.duration()`, which returns a duration object that represents the current time of the current stopwatch instance
- `stopwatch.hours()`, which returns the number of hours in the current time of the current stopwatch instance as a float
- `stopwatch.isReset()`, which returns `true` if the current stopwatch instance is newly created or has been reset and `false` otherwise
- `stopwatch.microseconds()`, which returns the number of microseconds in the current time of the current stopwatch instance as an integer
- `stopwatch.milliseconds()`, which returns the number of milliseconds in the current time of the current stopwatch instance as an integer
- `stopwatch.minutes()`, which returns the number of minutes in the current time of the current stopwatch instance as a float
- `stopwatch.reset()`, which resets the current stopwatch instance to zero
- `stopwatch.seconds()`, which returns the number of seconds in the current time of the current stopwatch instance as a float
- `stopwatch.start()`, which starts the current stopwatch instance
    - If the current stopwatch instance is already started, this method does nothing
- `stopwatch.stop()`, which stops the current stopwatch instance
    - If the current stopwatch instance is already stopped, this method does nothing
