package ast

import (
	"fmt"
	"time"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func defineDateFields(dateClass *LoxClass) {
	layouts := map[string]*LoxString{
		"ansic":       NewLoxString(time.ANSIC, '\''),
		"dateOnly":    NewLoxString(time.DateOnly, '\''),
		"dateTime":    NewLoxString(time.DateTime, '\''),
		"kitchen":     NewLoxString(time.Kitchen, '\''),
		"layout":      NewLoxString(time.Layout, '"'),
		"rfc822":      NewLoxString(time.RFC822, '\''),
		"rfc822z":     NewLoxString(time.RFC822Z, '\''),
		"rfc850":      NewLoxString(time.RFC850, '\''),
		"rfc1123":     NewLoxString(time.RFC1123, '\''),
		"rfc1123z":    NewLoxString(time.RFC1123Z, '\''),
		"rfc3339":     NewLoxString(time.RFC3339, '\''),
		"rfc3339nano": NewLoxString(time.RFC3339Nano, '\''),
		"rubyDate":    NewLoxString(time.RubyDate, '\''),
		"stamp":       NewLoxString(time.Stamp, '\''),
		"stampMicro":  NewLoxString(time.StampMicro, '\''),
		"stampMilli":  NewLoxString(time.StampMilli, '\''),
		"stampNano":   NewLoxString(time.StampNano, '\''),
		"timeOnly":    NewLoxString(time.TimeOnly, '\''),
		"unixDate":    NewLoxString(time.UnixDate, '\''),
	}
	for key, value := range layouts {
		dateClass.classProperties[key] = value
	}

	months := map[string]time.Month{
		"january":   time.January,
		"february":  time.February,
		"march":     time.March,
		"april":     time.April,
		"may":       time.May,
		"june":      time.June,
		"july":      time.July,
		"august":    time.August,
		"september": time.September,
		"october":   time.October,
		"november":  time.November,
		"december":  time.December,
	}
	for key, value := range months {
		dateClass.classProperties[key] = int64(value)
	}

	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	for key, value := range weekdays {
		dateClass.classProperties[key] = int64(value) + 1
	}
}

func (i *Interpreter) defineDateFuncs() {
	className := "Date"
	dateClass := NewLoxClass(className, nil, false)
	dateFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native Date class fn %v at %p>", name, &s)
		}
		dateClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Date.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}
	argMustBeTypeAn := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'Date.%v' must be an %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineDateFields(dateClass)
	dateFunc("date", 6, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Date.date' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Date.date' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'Date.date' must be an integer.")
		}
		if _, ok := args[3].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'Date.date' must be an integer.")
		}
		if _, ok := args[4].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fifth argument to 'Date.date' must be an integer.")
		}
		if _, ok := args[5].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Sixth argument to 'Date.date' must be an integer.")
		}
		year := int(args[0].(int64))
		month := time.Month(args[1].(int64))
		day := int(args[2].(int64))
		hour := int(args[3].(int64))
		minute := int(args[4].(int64))
		second := int(args[5].(int64))
		date := time.Date(year, month, day, hour, minute, second, 0, time.UTC)
		return NewLoxDate(date), nil
	})
	dateFunc("dateLocal", 6, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Date.dateLocal' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Date.dateLocal' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'Date.dateLocal' must be an integer.")
		}
		if _, ok := args[3].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'Date.dateLocal' must be an integer.")
		}
		if _, ok := args[4].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fifth argument to 'Date.dateLocal' must be an integer.")
		}
		if _, ok := args[5].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Sixth argument to 'Date.dateLocal' must be an integer.")
		}
		year := int(args[0].(int64))
		month := time.Month(args[1].(int64))
		day := int(args[2].(int64))
		hour := int(args[3].(int64))
		minute := int(args[4].(int64))
		second := int(args[5].(int64))
		date := time.Date(year, month, day, hour, minute, second, 0, time.Local)
		return NewLoxDate(date), nil
	})
	dateFunc("dateNow", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxDate(time.Now()), nil
	})
	dateFunc("monthStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if monthNum, ok := args[0].(int64); ok {
			if monthNum < 1 || monthNum > 12 {
				return NewLoxString("Unknown", '\''), nil
			}
			return NewLoxString(time.Month(monthNum).String(), '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "monthStr", "integer")
	})
	dateFunc("now", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return time.Now().UnixMilli(), nil
	})
	dateFunc("parse", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'Date.parse' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'Date.parse' must be a string.")
		}
		layout := args[0].(*LoxString).str
		dateStr := args[1].(*LoxString).str
		date, err := time.Parse(layout, dateStr)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxDate(date), nil
	})
	dateFunc("parseDefault", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if dateStr, ok := args[0].(*LoxString); ok {
			date, err := time.Parse(LoxDateDefaultFormat, dateStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxDate(date), nil
		}
		return argMustBeType(in.callToken, "parseDefault", "string")
	})
	dateFunc("sleepUntil", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxDate, ok := args[0].(*LoxDate); ok {
			time.Sleep(time.Until(loxDate.date))
			return nil, nil
		}
		return argMustBeType(in.callToken, "sleepUntil", "date")
	})
	dateFunc("weekdayStr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if weekdayNum, ok := args[0].(int64); ok {
			if weekdayNum < 1 || weekdayNum > 7 {
				return NewLoxString("Unknown", '\''), nil
			}
			return NewLoxString(time.Weekday(weekdayNum-1).String(), '\''), nil
		}
		return argMustBeTypeAn(in.callToken, "weekdayStr", "integer")
	})

	i.globals.Define(className, dateClass)
}
