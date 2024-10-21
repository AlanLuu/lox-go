# Date class methods and fields

The following layout fields are defined in the built-in `Date` class, which are all strings corresponding to date layouts:
- `Date.ansic`, `Date.dateOnly`, `Date.dateTime`, `Date.kitchen`, `Date.layout`, `Date.rfc822`, `Date.rfc822z`, `Date.rfc850`, `Date.rfc1123`, `Date.rfc1123z`, `Date.rfc3339`, `Date.rfc3339nano`, `Date.rubyDate`, `Date.stamp`, `Date.stampMicro`, `Date.stampMilli`, `Date.stampNano`, `Date.timeOnly`, `Date.unixDate`

The following month fields are defined in the built-in `Date` class, which are all integers corresponding to months:
- `Date.january`, `Date.february`, `Date.march`, `Date.april`, `Date.may`, `Date.june`, `Date.july`, `Date.august`, `Date.september`, `Date.october`, `Date.november`, `Date.december`

The following day of week fields are defined in the built-in `Date` class, which are all integers corresponding to days of weeks:
- `Date.sunday`, `Date.monday`, `Date.tuesday`, `Date.wednesday`, `Date.thursday`, `Date.friday`, `Date.saturday`

The following methods are defined in the built-in `Date` class:
- `Date.date(year, month, day, hour, minute, second)`, which returns a date object in UTC with the specified arguments, which are all integers
- `Date.dateLocal(year, month, day, hour, minute, second)`, which returns a date object in local time with the specified arguments, which are all integers
- `Date.dateNow()`, which returns a date object that represents the current date, with the year, month, day, hour, minute, and second being the values at the moment this method is called
- `Date.loopUntil(date, callback)`, which takes in a date object and a callback function and repeatedly invokes the callback as long as the current date is less than the specified date object
- `Date.monthStr(monthInt)`, which returns a string that corresponds to the month given by the specified integer, with `1` corresponding to `"January"`, `2` corresponding to `"February"`, and so on up to `12` corresponding to `"December"`
    - If `monthInt < 1` or `monthInt > 12`, the string `"Unknown"` is returned
- `Date.now()`, which returns the number of milliseconds since the Unix epoch as an integer
- `Date.parse(layout, str)`, which takes in a layout string and the string to parse and returns a date object that corresponds to the parsed string according to the layout string. If parsing is unsuccessful, a runtime error is thrown
- `Date.parseDefault(str)`, which takes in the string to parse and returns a date object that corresponds to the parsed string, where the layout is RFC 3339. If parsing is unsuccessful, a runtime error is thrown
- `Date.sleepUntil(date)`, which pauses the program until the current date is greater than or equal to the specified date argument
- `Date.time(hour, minute, second)`, which returns a date object in UTC with the year, month, and day being today's values with the specified hour, minute, and second arguments, which are all integers
- `Date.timeLocal(hour, minute, second)`, which returns a date object in local time with the year, month, and day being today's values with the specified hour, minute, and second arguments, which are all integers
- `Date.unix(seconds)`, which returns a date object corresponding to the date that is the specified number of seconds since the Unix epoch, where `seconds` is an integer
- `Date.unixMicro(microseconds)`, which returns a date object corresponding to the date that is the specified number of microseconds since the Unix epoch, where `microseconds` is an integer
- `Date.unixMilli(milliseconds)`, which returns a date object corresponding to the date that is the specified number of milliseconds since the Unix epoch, where `milliseconds` is an integer
- `Date.weekdayStr(weekdayInt)`, which returns a string that corresponds to the day of week given by the specified integer, with `1` corresponding to `"Sunday"`, `2` corresponding to `"Monday"`, and so on up to `7` corresponding to `"Saturday"`
    - If `weekdayInt < 1` or `weekdayInt > 7`, the string `"Unknown"` is returned

Date objects have the following methods associated with them:
- `date.add(duration)`, which returns a new date object that is the current date object with the specified duration object added to it
- `date.addDate(months, days, years)`, which returns a new date object that is the current date object with the specified months, days, and years added to it, which are all integers
- `date.compare(date2)`, which compares both `date` and `date2` and returns `0` if `date == date2`, `-1` if `date < date2`, and `1` if `date > date2`
- `date.day()`, which returns the day associated with the current date object as an integer
- `date.format(layout)`, which formats the date object according to the specified layout string into a string and returns that string
- `date.hour()`, which returns the hour associated with the current date object as an integer
- `date.inLocal()`, which returns a new date object that is the current date object in local time for display purposes
- `date.inUTC()` which returns a new date object that is the current date object in UTC time for display purposes
- `date.isAfter(date2)`, which returns `true` if `date` is greater than `date2` and `false` otherwise
- `date.isBefore(date2)`, which returns `true` if `date` is less than `date2` and `false` otherwise
- `date.isDST()`, which returns `true` if the current date object is in Daylight Savings Time and `false` otherwise
- `date.isLocal()`, which returns `true` if the current date object is in local time and `false` otherwise
- `date.isoWeek()`, which returns a list with two elements, with the first being the ISO 8601 year value of the current date object as an integer and the second being the ISO 8601 week value of the current date object as an integer
- `date.isUTC()`, which returns `true` if the current date object is in UTC time and `false` otherwise
- `date.isZero()`, which returns `true` if the current date object corresponds to the date January 1, 0001, with a time of 00:00:00 UTC and `false` otherwise
- `date.local()`, which returns a new date object that is the current date object in local time
- `date.location()`, which returns a string that represents the current time zone information associated with the current date object, which is usually `"UTC"` or `"Local"`
- `date.loopUntil(callback)`, which takes in a callback function and repeatedly invokes the callback as long as the current date is less than `date`
- `date.minute()`, which returns the minute associated with the current date object as an integer
- `date.month()`, which returns the month associated with the current date object as an integer
- `date.monthStr()`, which returns the month associated with the current date object as an string
- `date.nanosecond()`, which returns the nanosecond associated with the current date object as an integer
- `date.now()`, which is an alias for `date.unixMilli`
- `date.second()`, which returns the second associated with the current date object as an integer
- `date.setDay(day)`, which sets the day of the current date object to the specified day integer and returns the current date object itself
- `date.setHour(hour)`, which sets the hour of the current date object to the specified hour integer and returns the current date object itself
- `date.setMinute(minute)`, which sets the minute of the current date object to the specified minute integer and returns the current date object itself
- `date.setMonth(month)`, which sets the month of the current date object to the specified month integer and returns the current date object itself
- `date.setSecond(second)`, which sets the second of the current date object to the specified second integer and returns the current date object itself
- `date.setYear(year)`, which sets the year of the current date object to the specified year integer and returns the current date object itself
- `date.sleepUntil()`, which pauses the program until the current date is greater than or equal to the current date object
- `date.string()`, which formats the date object according to the RFC 3339 layout into a string and returns that string
- `date.sub(date2)`, which returns a duration object that is the difference between the current date object and the specified date object
- `date.unix()`, which returns the number of seconds since the Unix epoch of the current date object as an integer
- `date.unixMicro()`, which returns the number of microseconds since the Unix epoch of the current date object as an integer
- `date.unixMilli()`, which returns the number of milliseconds since the Unix epoch of the current date object as an integer
- `date.unixNano()`, which returns the number of nanoseconds since the Unix epoch of the current date object as an integer
- `date.utc()`, which returns a new date object that is the current date object in UTC time
- `date.weekday()`, which returns the day of week associated with the current date object as an integer
- `date.weekdayStr()`, which returns the day of week associated with the current date object as an string
- `date.year()`, which returns the year associated with the current date object as an integer
- `date.yearDay()`, which returns the day of year associated with the current date object as an integer
- `date.zone()`, which returns a list with two elements, with the first being a string that is the abbreviated version of the current date object's time zone, and the second being an integer that is the offset in seconds east of UTC
- `date.zoneBounds()` which returns a list with two elements, with the first being a date object that is the lower bound of the current date object's time zone, and the second being a date object that is the upper bound of the current date object's time zone
