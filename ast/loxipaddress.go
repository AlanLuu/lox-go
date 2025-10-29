package ast

import (
	"fmt"
	"net"
	"reflect"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxIPAddressIterator struct {
	ip    net.IP
	index int
}

func (l *LoxIPAddressIterator) HasNext() bool {
	return l.index < len(l.ip)
}

func (l *LoxIPAddressIterator) Next() any {
	element := l.ip[l.index]
	l.index++
	return int64(element)
}

type LoxIPAddress struct {
	ip      net.IP
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxIPAddress(ip net.IP) *LoxIPAddress {
	return &LoxIPAddress{
		ip:      ip,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxIPAddressTo4(ip net.IP) *LoxIPAddress {
	loxIP := NewLoxIPAddress(ip)
	if ip != nil {
		if ipTo4 := ip.To4(); ipTo4 != nil {
			loxIP.ip = ipTo4
		}
	}
	return loxIP
}

func (l *LoxIPAddress) isIpOther() bool {
	return !l.isIpv4() && !l.isIpv6()
}

func (l *LoxIPAddress) isIpv4() bool {
	return len(l.ip) == 4
}

func (l *LoxIPAddress) isIpv4Len16() bool {
	if len(l.ip) != 16 {
		return false
	}
	var i uint8 = 0
	for ; i < 10; i++ {
		if l.ip[i] != 0 {
			return false
		}
	}
	return l.ip[i] == 255 && l.ip[i+1] == 255
}

func (l *LoxIPAddress) isIpv6() bool {
	return len(l.ip) == 16
}

func (l *LoxIPAddress) isNil() bool {
	return len(l.ip) == 0
}

func (l *LoxIPAddress) str() string {
	if l.isNil() {
		return "nil"
	}
	return l.ip.String()
}

func (l *LoxIPAddress) Equals(obj any) bool {
	switch obj := obj.(type) {
	case *LoxIPAddress:
		if l == obj {
			return true
		}
		return l.ip.Equal(obj.ip)
	default:
		return false
	}
}

func (l *LoxIPAddress) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	ipFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native ip address fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'ip address.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'ip address.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeIP := func() (any, error) {
		return argMustBeTypeAn("IP address instance")
	}
	switch methodName {
	case "defaultMask":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			defaultMask := l.ip.DefaultMask()
			return NewLoxIPAddressTo4(net.IP(defaultMask)), nil
		})
	case "defaultMaskOrNil":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			defaultMask := l.ip.DefaultMask()
			if defaultMask == nil {
				return nil, nil
			}
			return NewLoxIPAddressTo4(net.IP(defaultMask)), nil
		})
	case "getByte", "get":
		return ipFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				if arg < 1 {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"Argument to 'ip address.%v' cannot be less than 1.",
							methodName,
						),
					)
				}
				if arg > int64(len(l.ip)) {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"ip address.%v: integer %v out of range.",
							methodName,
							arg,
						),
					)
				}
				return int64(l.ip[arg-1]), nil
			}
			return argMustBeType("integer")
		})
	case "getIndex":
		return ipFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if arg, ok := args[0].(int64); ok {
				if arg < 0 {
					return nil, loxerror.RuntimeError(name,
						"Argument to 'ip address.getIndex' cannot be negative.")
				}
				if arg >= int64(len(l.ip)) {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"ip address.getIndex: integer %v out of range.",
							arg,
						),
					)
				}
				return int64(l.ip[arg]), nil
			}
			return argMustBeType("integer")
		})
	case "getSliceBuf":
		return ipFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ipLen := int64(len(l.ip))
			argsLen := len(args)
			switch argsLen {
			case 1:
				if arg, ok := args[0].(int64); ok {
					if arg < 1 || arg > ipLen {
						return EmptyLoxBuffer(), nil
					}
					buffer := EmptyLoxBufferCap(ipLen - arg + 1)
					for i := arg - 1; i < ipLen; i++ {
						addErr := buffer.add(int64(l.ip[i]))
						if addErr != nil {
							return nil, loxerror.RuntimeError(name, addErr.Error())
						}
					}
					return buffer, nil
				}
				return argMustBeTypeAn("integer")
			case 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'ip address.getSliceBuf' must be an integer.")
				}
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'ip address.getSliceBuf' must be an integer.")
				}
				start := args[0].(int64)
				if start < 1 || start > ipLen {
					return EmptyLoxBuffer(), nil
				}
				stop := args[1].(int64)
				if stop < 1 {
					return EmptyLoxBuffer(), nil
				}
				capacity := stop - start + 1
				if capacity < 0 {
					return EmptyLoxBuffer(), nil
				}
				buffer := EmptyLoxBufferCap(capacity)
				for i := start; i <= stop; i++ {
					if i-1 >= ipLen {
						break
					}
					addErr := buffer.add(int64(l.ip[i-1]))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "getSliceIndexBuf":
		return ipFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ipLen := int64(len(l.ip))
			argsLen := len(args)
			switch argsLen {
			case 1:
				if arg, ok := args[0].(int64); ok {
					if arg < 0 || arg >= ipLen {
						return EmptyLoxBuffer(), nil
					}
					buffer := EmptyLoxBufferCap(ipLen - arg)
					for i := arg; i < ipLen; i++ {
						addErr := buffer.add(int64(l.ip[i]))
						if addErr != nil {
							return nil, loxerror.RuntimeError(name, addErr.Error())
						}
					}
					return buffer, nil
				}
				return argMustBeTypeAn("integer")
			case 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'ip address.getSliceIndexBuf' must be an integer.")
				}
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'ip address.getSliceIndexBuf' must be an integer.")
				}
				start := args[0].(int64)
				if start < 0 || start >= ipLen {
					return EmptyLoxBuffer(), nil
				}
				stop := args[1].(int64)
				if stop < 0 {
					return EmptyLoxBuffer(), nil
				}
				capacity := stop - start
				if capacity < 0 {
					return EmptyLoxList(), nil
				}
				buffer := EmptyLoxBufferCap(capacity)
				for i := start; i <= stop; i++ {
					if i >= ipLen {
						break
					}
					addErr := buffer.add(int64(l.ip[i]))
					if addErr != nil {
						return nil, loxerror.RuntimeError(name, addErr.Error())
					}
				}
				return buffer, nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "getSliceIndexIter":
		return ipFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ipLen := int64(len(l.ip))
			argsLen := len(args)
			switch argsLen {
			case 1:
				if arg, ok := args[0].(int64); ok {
					if arg < 0 || arg >= ipLen {
						return EmptyLoxIterator(), nil
					}
					i := arg
					iterator := ProtoIterator{}
					iterator.hasNextMethod = func() bool {
						return i < ipLen
					}
					iterator.nextMethod = func() any {
						element := l.ip[i]
						i++
						return int64(element)
					}
					return NewLoxIterator(iterator), nil
				}
				return argMustBeTypeAn("integer")
			case 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'ip address.getSliceIndexIter' must be an integer.")
				}
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'ip address.getSliceIndexIter' must be an integer.")
				}
				start := args[0].(int64)
				if start < 0 || start >= ipLen {
					return EmptyLoxIterator(), nil
				}
				stop := args[1].(int64)
				if stop < 0 {
					return EmptyLoxIterator(), nil
				}
				capacity := stop - start
				if capacity < 0 {
					return EmptyLoxIterator(), nil
				}
				i := start
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if i >= ipLen {
						return false
					}
					return i <= stop
				}
				iterator.nextMethod = func() any {
					element := l.ip[i]
					i++
					return int64(element)
				}
				return NewLoxIterator(iterator), nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "getSliceIndexList":
		return ipFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ipLen := int64(len(l.ip))
			argsLen := len(args)
			switch argsLen {
			case 1:
				if arg, ok := args[0].(int64); ok {
					if arg < 0 || arg >= ipLen {
						return EmptyLoxList(), nil
					}
					elements := list.NewListCap[any](ipLen - arg)
					for i := arg; i < ipLen; i++ {
						elements.Add(int64(l.ip[i]))
					}
					return NewLoxList(elements), nil
				}
				return argMustBeTypeAn("integer")
			case 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'ip address.getSliceIndexList' must be an integer.")
				}
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'ip address.getSliceIndexList' must be an integer.")
				}
				start := args[0].(int64)
				if start < 0 || start >= ipLen {
					return EmptyLoxList(), nil
				}
				stop := args[1].(int64)
				if stop < 0 {
					return EmptyLoxList(), nil
				}
				capacity := stop - start
				if capacity < 0 {
					return EmptyLoxList(), nil
				}
				elements := list.NewListCap[any](capacity)
				for i := start; i <= stop; i++ {
					if i >= ipLen {
						break
					}
					elements.Add(int64(l.ip[i]))
				}
				return NewLoxList(elements), nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "getSliceIter":
		return ipFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ipLen := int64(len(l.ip))
			argsLen := len(args)
			switch argsLen {
			case 1:
				if arg, ok := args[0].(int64); ok {
					if arg < 1 || arg > ipLen {
						return EmptyLoxIterator(), nil
					}
					i := arg - 1
					iterator := ProtoIterator{}
					iterator.hasNextMethod = func() bool {
						return i < ipLen
					}
					iterator.nextMethod = func() any {
						element := l.ip[i]
						i++
						return int64(element)
					}
					return NewLoxIterator(iterator), nil
				}
				return argMustBeTypeAn("integer")
			case 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'ip address.getSliceIter' must be an integer.")
				}
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'ip address.getSliceIter' must be an integer.")
				}
				start := args[0].(int64)
				if start < 1 || start > ipLen {
					return EmptyLoxIterator(), nil
				}
				stop := args[1].(int64)
				if stop < 1 {
					return EmptyLoxIterator(), nil
				}
				capacity := stop - start + 1
				if capacity < 0 {
					return EmptyLoxIterator(), nil
				}
				i := start
				iterator := ProtoIterator{}
				iterator.hasNextMethod = func() bool {
					if i-1 >= ipLen {
						return false
					}
					return i <= stop
				}
				iterator.nextMethod = func() any {
					element := l.ip[i-1]
					i++
					return int64(element)
				}
				return NewLoxIterator(iterator), nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "getSliceList":
		return ipFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			ipLen := int64(len(l.ip))
			argsLen := len(args)
			switch argsLen {
			case 1:
				if arg, ok := args[0].(int64); ok {
					if arg < 1 || arg > ipLen {
						return EmptyLoxList(), nil
					}
					elements := list.NewListCap[any](ipLen - arg + 1)
					for i := arg - 1; i < ipLen; i++ {
						elements.Add(int64(l.ip[i]))
					}
					return NewLoxList(elements), nil
				}
				return argMustBeTypeAn("integer")
			case 2:
				if _, ok := args[0].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"First argument to 'ip address.getSliceList' must be an integer.")
				}
				if _, ok := args[1].(int64); !ok {
					return nil, loxerror.RuntimeError(name,
						"Second argument to 'ip address.getSliceList' must be an integer.")
				}
				start := args[0].(int64)
				if start < 1 || start > ipLen {
					return EmptyLoxList(), nil
				}
				stop := args[1].(int64)
				if stop < 1 {
					return EmptyLoxList(), nil
				}
				capacity := stop - start + 1
				if capacity < 0 {
					return EmptyLoxList(), nil
				}
				elements := list.NewListCap[any](capacity)
				for i := start; i <= stop; i++ {
					if i-1 >= ipLen {
						break
					}
					elements.Add(int64(l.ip[i-1]))
				}
				return NewLoxList(elements), nil
			default:
				return nil, loxerror.RuntimeError(name,
					fmt.Sprintf("Expected 1 or 2 arguments but got %v.", argsLen))
			}
		})
	case "httpurl":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if l.isNil() {
				return EmptyLoxString(), nil
			}
			if l.isIpv4() {
				return NewLoxStringQuote("http://" + l.str()), nil
			}
			return NewLoxStringQuote("[" + l.str() + "]"), nil
		})
	case "ipv4Len16Iter":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			if !l.isIpv4Len16() {
				return EmptyLoxIterator(), nil
			}
			var index uint8 = 12
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				return index < 16
			}
			iterator.nextMethod = func() any {
				element := l.ip[index]
				index++
				return int64(element)
			}
			return NewLoxIterator(iterator), nil
		})
	case "isGlobalUnicast":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsGlobalUnicast(), nil
		})
	case "isInterfaceLocalMulticast":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsInterfaceLocalMulticast(), nil
		})
	case "isIpOther":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isIpOther(), nil
		})
	case "isIpv4":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isIpv4(), nil
		})
	case "isIpv4Len16":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isIpv4Len16(), nil
		})
	case "isIpv6":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isIpv6(), nil
		})
	case "isLinkLocalMulticast":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsLinkLocalMulticast(), nil
		})
	case "isLinkLocalUnicast":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsLinkLocalUnicast(), nil
		})
	case "isLoopback":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsLoopback(), nil
		})
	case "isMulticast":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsMulticast(), nil
		})
	case "isNil":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isNil(), nil
		})
	case "isPrivate":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsPrivate(), nil
		})
	case "isUnspecified":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.ip.IsUnspecified(), nil
		})
	case "mask":
		return ipFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxIPAddr, ok := args[0].(*LoxIPAddress); ok {
				mask := l.ip.Mask(net.IPMask(loxIPAddr.ip))
				return NewLoxIPAddressTo4(mask), nil
			}
			return argMustBeIP()
		})
	case "maskOrNil":
		return ipFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxIPAddr, ok := args[0].(*LoxIPAddress); ok {
				mask := l.ip.Mask(net.IPMask(loxIPAddr.ip))
				if mask == nil {
					return nil, nil
				}
				return NewLoxIPAddressTo4(mask), nil
			}
			return argMustBeIP()
		})
	case "setByte", "set":
		return ipFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"First argument to 'ip address.%v' must be an integer.",
						methodName,
					),
				)
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Second argument to 'ip address.%v' must be an integer.",
						methodName,
					),
				)
			}
			pos := args[0].(int64)
			if pos < 1 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"First argument to 'ip address.%v' cannot be less than 1.",
						methodName,
					),
				)
			}
			if pos > int64(len(l.ip)) {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"ip address.%v: first argument integer %v out of range.",
						methodName,
						pos,
					),
				)
			}
			value := args[1].(int64)
			if value < 0 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Second argument to 'ip address.%v' cannot be negative.",
						methodName,
					),
				)
			}
			if value > 255 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Second argument to 'ip address.%v' cannot be greater than 255.",
						methodName,
					),
				)
			}
			l.ip[pos-1] = byte(value)
			return nil, nil
		})
	case "setIndex":
		return ipFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'ip address.setIndex' must be an integer.")
			}
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'ip address.setIndex' must be an integer.")
			}
			pos := args[0].(int64)
			if pos < 0 {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'ip address.setIndex' cannot be negative.")
			}
			if pos >= int64(len(l.ip)) {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"ip address.setIndex: first argument integer %v out of range.",
						pos,
					),
				)
			}
			value := args[1].(int64)
			if value < 0 {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'ip address.setIndex' cannot be negative.")
			}
			if value > 255 {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'ip address.setIndex' cannot be greater than 255.")
			}
			l.ip[pos] = byte(value)
			return nil, nil
		})
	case "strictEquals":
		return ipFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxIPAddr, ok := args[0].(*LoxIPAddress); ok {
				return reflect.DeepEqual(l.ip, loxIPAddr.ip), nil
			}
			return argMustBeIP()
		})
	case "string":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.str()), nil
		})
	case "to16":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxIPAddress(l.ip.To16()), nil
		})
	case "to16OrNil":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			to16 := l.ip.To16()
			if to16 == nil {
				return nil, nil
			}
			return NewLoxIPAddress(to16), nil
		})
	case "to4":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxIPAddress(l.ip.To4()), nil
		})
	case "to4OrNil":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			to4 := l.ip.To4()
			if to4 == nil {
				return nil, nil
			}
			return NewLoxIPAddress(to4), nil
		})
	case "toBuffer":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			buffer := EmptyLoxBufferCap(int64(len(l.ip)))
			for _, element := range l.ip {
				addErr := buffer.add(int64(element))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			return buffer, nil
		})
	case "toList":
		return ipFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			newList := list.NewListCap[any](int64(len(l.ip)))
			for _, element := range l.ip {
				newList.Add(int64(element))
			}
			return NewLoxList(newList), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "IP address instances have no property called '"+methodName+"'.")
}

func (l *LoxIPAddress) Iterator() interfaces.Iterator {
	return &LoxIPAddressIterator{l.ip, 0}
}

func (l *LoxIPAddress) Length() int64 {
	return int64(len(l.ip))
}

func (l *LoxIPAddress) String() string {
	if l.isIpv4() {
		return fmt.Sprintf("<IPv4 address: %v>", l.str())
	}
	if l.isIpv6() {
		return fmt.Sprintf("<IPv6 address: %v>", l.str())
	}
	return fmt.Sprintf("<IP address: %v>", l.str())
}

func (l *LoxIPAddress) Type() string {
	return "ip address"
}
