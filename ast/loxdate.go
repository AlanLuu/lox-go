package ast

import (
	"fmt"
	"time"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

const LoxDateDefaultFormat = time.RFC3339

type LoxDate struct {
	date    time.Time
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxDate(date time.Time) *LoxDate {
	return &LoxDate{
		date:    date,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxDate) defaultFormatStr() string {
	return l.date.Format(LoxDateDefaultFormat)
}

func (l *LoxDate) setDay(day int) {
	l.date = time.Date(
		l.date.Year(),
		l.date.Month(),
		day,
		l.date.Hour(),
		l.date.Minute(),
		l.date.Second(),
		l.date.Nanosecond(),
		l.date.Location(),
	)
}

func (l *LoxDate) setHour(hour int) {
	l.date = time.Date(
		l.date.Year(),
		l.date.Month(),
		l.date.Day(),
		hour,
		l.date.Minute(),
		l.date.Second(),
		l.date.Nanosecond(),
		l.date.Location(),
	)
}

func (l *LoxDate) setMinute(minute int) {
	l.date = time.Date(
		l.date.Year(),
		l.date.Month(),
		l.date.Day(),
		l.date.Hour(),
		minute,
		l.date.Second(),
		l.date.Nanosecond(),
		l.date.Location(),
	)
}

func (l *LoxDate) setMonth(month time.Month) {
	l.date = time.Date(
		l.date.Year(),
		month,
		l.date.Day(),
		l.date.Hour(),
		l.date.Minute(),
		l.date.Second(),
		l.date.Nanosecond(),
		l.date.Location(),
	)
}

func (l *LoxDate) setSecond(second int) {
	l.date = time.Date(
		l.date.Year(),
		l.date.Month(),
		l.date.Day(),
		l.date.Hour(),
		l.date.Minute(),
		second,
		l.date.Nanosecond(),
		l.date.Location(),
	)
}

func (l *LoxDate) setYear(year int) {
	l.date = time.Date(
		year,
		l.date.Month(),
		l.date.Day(),
		l.date.Hour(),
		l.date.Minute(),
		l.date.Second(),
		l.date.Nanosecond(),
		l.date.Location(),
	)
}

func (l *LoxDate) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxDate:
		return l.date.Equal(obj.date)
	default:
		return false
	}
}

func (l *LoxDate) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	dateFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native date fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'date.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'date.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "add":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDuration, ok := args[0].(*LoxDuration); ok {
				return NewLoxDate(l.date.Add(loxDuration.duration)), nil
			}
			return argMustBeType("duration")
		})
	case "addDate":
		return dateFunc(3, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'date.addDate' must be an integer.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'date.addDate' must be an integer.")
			}
			if _, ok := args[2].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Third argument to 'date.addDate' must be an integer.")
			}
			months := int(args[0].(int64))
			days := int(args[1].(int64))
			years := int(args[2].(int64))
			return NewLoxDate(l.date.AddDate(years, months, days)), nil
		})
	case "compare":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDate, ok := args[0].(*LoxDate); ok {
				return int64(l.date.Compare(loxDate.date)), nil
			}
			return argMustBeType("date")
		})
	case "day":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Day()), nil
		})
	case "format":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if format, ok := args[0].(*LoxString); ok {
				return NewLoxStringQuote(l.date.Format(format.str)), nil
			}
			return argMustBeType("string")
		})
	case "hour":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Hour()), nil
		})
	case "inLocal":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxDate(l.date.In(time.Local)), nil
		})
	case "inUTC":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxDate(l.date.In(time.UTC)), nil
		})
	case "isAfter":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDate, ok := args[0].(*LoxDate); ok {
				return l.date.After(loxDate.date), nil
			}
			return argMustBeType("date")
		})
	case "isBefore":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDate, ok := args[0].(*LoxDate); ok {
				return l.date.Before(loxDate.date), nil
			}
			return argMustBeType("date")
		})
	case "isDST":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.IsDST(), nil
		})
	case "isLocal":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.Location().String() == "Local", nil
		})
	case "isoWeek":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			year, week := l.date.ISOWeek()
			elements := list.NewListCap[any](2)
			elements.Add(int64(year))
			elements.Add(int64(week))
			return NewLoxList(elements), nil
		})
	case "isUTC":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.Location().String() == "UTC", nil
		})
	case "isZero":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.IsZero(), nil
		})
	case "local":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxDate(l.date.Local()), nil
		})
	case "location":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.date.Location().String()), nil
		})
	case "loopUntil":
		return dateFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(*LoxFunction); ok {
				stopCallback := false
				callbackChan := make(chan struct{}, 1)
				errorChan := make(chan error, 1)
				argList := getArgList(callback, 0)
				go func() {
					for !stopCallback {
						result, resultErr := callback.call(i, argList)
						if resultErr != nil && result == nil {
							errorChan <- resultErr
							break
						}
					}
					callbackChan <- struct{}{}
				}()
				select {
				case err := <-errorChan:
					return nil, err
				case <-time.After(time.Until(l.date)):
					stopCallback = true
					<-callbackChan
				}
				return nil, nil
			}
			return argMustBeType("function")
		})
	case "minute":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Minute()), nil
		})
	case "month":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Month()), nil
		})
	case "monthStr":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(l.date.Month().String(), '\''), nil
		})
	case "nanosecond":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Nanosecond()), nil
		})
	case "second":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Second()), nil
		})
	case "setDay":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if day, ok := args[0].(int64); ok {
				l.setDay(int(day))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "setHour":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if hour, ok := args[0].(int64); ok {
				l.setHour(int(hour))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "setMinute":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if minute, ok := args[0].(int64); ok {
				l.setMinute(int(minute))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "setMonth":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if month, ok := args[0].(int64); ok {
				l.setMonth(time.Month(month))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "setSecond":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if second, ok := args[0].(int64); ok {
				l.setSecond(int(second))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "setYear":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if year, ok := args[0].(int64); ok {
				l.setYear(int(year))
				return l, nil
			}
			return argMustBeTypeAn("integer")
		})
	case "sleepUntil":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			time.Sleep(time.Until(l.date))
			return nil, nil
		})
	case "string":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(l.defaultFormatStr(), '\''), nil
		})
	case "sub":
		return dateFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDate, ok := args[0].(*LoxDate); ok {
				return NewLoxDuration(l.date.Sub(loxDate.date)), nil
			}
			return argMustBeType("date")
		})
	case "unix":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.Unix(), nil
		})
	case "unixMicro":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.UnixMicro(), nil
		})
	case "unixMilli", "now":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.UnixMilli(), nil
		})
	case "unixNano":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.date.UnixNano(), nil
		})
	case "utc":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxDate(l.date.UTC()), nil
		})
	case "weekday":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Weekday()) + 1, nil
		})
	case "weekdayStr":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxString(l.date.Weekday().String(), '\''), nil
		})
	case "year":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.Year()), nil
		})
	case "yearDay":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return int64(l.date.YearDay()), nil
		})
	case "zone":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			name, offset := l.date.Zone()
			zoneList := list.NewListCap[any](2)
			zoneList.Add(NewLoxStringQuote(name))
			zoneList.Add(int64(offset))
			return NewLoxList(zoneList), nil
		})
	case "zoneBounds":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			zoneBoundsList := list.NewListCap[any](2)
			start, end := l.date.ZoneBounds()
			zoneBoundsList.Add(NewLoxDate(start))
			zoneBoundsList.Add(NewLoxDate(end))
			return NewLoxList(zoneBoundsList), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Dates have no property called '"+methodName+"'.")
}

func (l *LoxDate) String() string {
	return fmt.Sprintf("<date: %v>", l.defaultFormatStr())
}

func (l *LoxDate) Type() string {
	return "date"
}
