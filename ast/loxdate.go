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
	case "isoWeek":
		return dateFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			year, week := l.date.ISOWeek()
			elements := list.NewListCap[any](2)
			elements.Add(int64(year))
			elements.Add(int64(week))
			return NewLoxList(elements), nil
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
