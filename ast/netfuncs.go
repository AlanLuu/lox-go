package ast

import (
	"fmt"
	"net"
	"strings"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

func defineNetFields(netClass *LoxClass) {
	stringFields := [...]string{
		"tcp",
		"tcp4",
		"tcp6",
		"udp",
		"udp4",
		"udp6",
		"ip",
		"ip4",
		"ip6",
		"unix",
		"unixgram",
		"unixpacket",
	}
	for _, field := range stringFields {
		netClass.classProperties[field] = NewLoxString(field, '\'')
	}
}

func (i *Interpreter) defineNetFuncs() {
	className := "net"
	netClass := NewLoxClass(className, nil, false)
	netFunc := func(name string, arity int, method func(*Interpreter, list.List[any]) (any, error)) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native net fn %v at %p>", name, &s)
		}
		netClass.classProperties[name] = s
	}
	argMustBeType := func(callToken *token.Token, name string, theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'net.%v' must be a %v.", name, theType)
		return nil, loxerror.RuntimeError(callToken, errStr)
	}

	defineNetFields(netClass)
	netFunc("dial", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dial' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dial' must be a string.")
		}
		network := args[0].(*LoxString).str
		address := args[1].(*LoxString).str
		conn, err := net.Dial(network, address)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialOrNil", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialOrNil' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialOrNil' must be a string.")
		}
		network := args[0].(*LoxString).str
		address := args[1].(*LoxString).str
		conn, err := net.Dial(network, address)
		if err != nil {
			return nil, nil
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialIP", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialIP' must be a string.")
		}
		if _, ok := args[1].(*LoxIPAddress); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialIP' must be an IP address instance.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialIP' must be an integer.")
		}
		address := args[1].(*LoxIPAddress)
		if address.isNil() {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialIP' cannot be a nil IP address instance.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialIP' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address.ip.String() + ":" + fmt.Sprint(portNum)
		conn, err := net.Dial(network, addressStr)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialIPOrNil", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialIPOrNil' must be a string.")
		}
		if _, ok := args[1].(*LoxIPAddress); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialIPOrNil' must be an IP address instance.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialIPOrNil' must be an integer.")
		}
		address := args[1].(*LoxIPAddress)
		if address.isNil() {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialIPOrNil' cannot be a nil IP address instance.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialIPOrNil' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address.ip.String() + ":" + fmt.Sprint(portNum)
		conn, err := net.Dial(network, addressStr)
		if err != nil {
			return nil, nil
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialPort", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialPort' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialPort' must be a string.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialPort' must be an integer.")
		}
		address := args[1].(*LoxString).str
		if strings.Contains(address, ":") {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialPort' cannot contain the character ':'.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialPort' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address + ":" + fmt.Sprint(portNum)
		conn, err := net.Dial(network, addressStr)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialPortOrNil", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialPortOrNil' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialPortOrNil' must be a string.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialPortOrNil' must be an integer.")
		}
		address := args[1].(*LoxString).str
		if strings.Contains(address, ":") {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialPortOrNil' cannot contain the character ':'.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialPortOrNil' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address + ":" + fmt.Sprint(portNum)
		conn, err := net.Dial(network, addressStr)
		if err != nil {
			return nil, nil
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialTimeout", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialTimeout' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeout' must be a string.")
		}
		if _, ok := args[2].(*LoxDuration); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeout' must be a duration.")
		}
		network := args[0].(*LoxString).str
		address := args[1].(*LoxString).str
		timeout := args[2].(*LoxDuration).duration
		conn, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialTimeoutOrNil", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialTimeoutOrNil' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutOrNil' must be a string.")
		}
		if _, ok := args[2].(*LoxDuration); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutOrNil' must be a duration.")
		}
		network := args[0].(*LoxString).str
		address := args[1].(*LoxString).str
		timeout := args[2].(*LoxDuration).duration
		conn, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			return nil, nil
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialTimeoutIP", 4, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialTimeoutIP' must be a string.")
		}
		if _, ok := args[1].(*LoxIPAddress); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutIP' must be an IP address instance.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutIP' must be an integer.")
		}
		if _, ok := args[3].(*LoxDuration); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'net.dialTimeoutIP' must be a duration.")
		}
		address := args[1].(*LoxIPAddress)
		if address.isNil() {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutIP' cannot be a nil IP address instance.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutIP' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address.ip.String() + ":" + fmt.Sprint(portNum)
		timeout := args[3].(*LoxDuration).duration
		conn, err := net.DialTimeout(network, addressStr, timeout)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialTimeoutIPOrNil", 4, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialTimeoutIPOrNil' must be a string.")
		}
		if _, ok := args[1].(*LoxIPAddress); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutIPOrNil' must be an IP address instance.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutIPOrNil' must be an integer.")
		}
		if _, ok := args[3].(*LoxDuration); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'net.dialTimeoutIPOrNil' must be a duration.")
		}
		address := args[1].(*LoxIPAddress)
		if address.isNil() {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutIPOrNil' cannot be a nil IP address instance.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutIPOrNil' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address.ip.String() + ":" + fmt.Sprint(portNum)
		timeout := args[3].(*LoxDuration).duration
		conn, err := net.DialTimeout(network, addressStr, timeout)
		if err != nil {
			return nil, nil
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialTimeoutPort", 4, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialTimeoutPort' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutPort' must be a string.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutPort' must be an integer.")
		}
		if _, ok := args[3].(*LoxDuration); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'net.dialTimeoutPort' must be a duration.")
		}
		address := args[1].(*LoxString).str
		if strings.Contains(address, ":") {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutPort' cannot contain the character ':'.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutPort' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address + ":" + fmt.Sprint(portNum)
		timeout := args[3].(*LoxDuration).duration
		conn, err := net.DialTimeout(network, addressStr, timeout)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("dialTimeoutPortOrNil", 4, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.dialTimeoutPortOrNil' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutPortOrNil' must be a string.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutPortOrNil' must be an integer.")
		}
		if _, ok := args[3].(*LoxDuration); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'net.dialTimeoutPortOrNil' must be a duration.")
		}
		address := args[1].(*LoxString).str
		if strings.Contains(address, ":") {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.dialTimeoutPortOrNil' cannot contain the character ':'.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.dialTimeoutPortOrNil' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address + ":" + fmt.Sprint(portNum)
		timeout := args[3].(*LoxDuration).duration
		conn, err := net.DialTimeout(network, addressStr, timeout)
		if err != nil {
			return nil, nil
		}
		return NewLoxConnection(conn), nil
	})
	netFunc("interfaceAddrs", 0, func(in *Interpreter, _ list.List[any]) (any, error) {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		addrsList := list.NewListCap[any](int64(len(addrs)))
		for _, addr := range addrs {
			innerList := list.NewListCap[any](2)
			innerList.Add(NewLoxStringQuote(addr.Network()))
			innerList.Add(NewLoxStringQuote(addr.String()))
			addrsList.Add(NewLoxList(innerList))
		}
		return NewLoxList(addrsList), nil
	})
	netFunc("ipOther", -1, func(in *Interpreter, args list.List[any]) (any, error) {
		argsLen := len(args)
		if argsLen == 0 {
			return NewLoxIPAddress(nil), nil
		}
		ipArr := make([]byte, 0, argsLen)
		for i, arg := range args {
			switch arg := arg.(type) {
			case int64:
				if arg < 0 {
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"net.ipOther: argument #%v cannot be negative.",
							i+1,
						),
					)
				}
				if arg > 255 {
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"net.ipOther: argument #%v cannot be greater than 255.",
							i+1,
						),
					)
				}
				ipArr = append(ipArr, byte(arg))
			default:
				ipArr = nil
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"net.ipOther: argument #%v must be an integer.",
						i+1,
					),
				)
			}
		}
		return NewLoxIPAddress(net.IP(ipArr)), nil
	})
	netFunc("ipOtherBuf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxBuffer, ok := args[0].(*LoxBuffer); ok {
			elements := loxBuffer.elements
			elementsLen := len(elements)
			if elementsLen == 0 {
				return NewLoxIPAddress(nil), nil
			}
			ipArr := make([]byte, 0, elementsLen)
			for _, element := range elements {
				ipArr = append(ipArr, byte(element.(int64)))
			}
			return NewLoxIPAddress(net.IP(ipArr)), nil
		}
		return argMustBeType(in.callToken, "ipOtherBuf", "buffer")
	})
	netFunc("ipOtherIter", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			var i int64 = 1
			ipArr := []byte{}
			it := iterable.Iterator()
			for it.HasNext() {
				element := it.Next()
				switch element := element.(type) {
				case int64:
					if element < 0 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipOtherIter: iterable element #%v cannot be negative.",
								i,
							),
						)
					}
					if element > 255 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipOtherIter: iterable element #%v cannot be greater than 255.",
								i,
							),
						)
					}
					ipArr = append(ipArr, byte(element))
				default:
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"net.ipOtherIter: iterable element #%v must be an integer.",
							i,
						),
					)
				}
				i++
			}
			if len(ipArr) == 0 {
				return NewLoxIPAddress(nil), nil
			}
			return NewLoxIPAddress(net.IP(ipArr)), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("net.ipOtherIter: type '%v' is not iterable.", getType(args[0])))
	})
	netFunc("ipOtherList", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			elements := loxList.elements
			elementsLen := len(elements)
			if elementsLen == 0 {
				return NewLoxIPAddress(nil), nil
			}
			ipArr := make([]byte, 0, elementsLen)
			for i, element := range elements {
				switch element := element.(type) {
				case int64:
					if element < 0 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipOtherList: list element at index %v cannot be negative.",
								i,
							),
						)
					}
					if element > 255 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipOtherList: list element at index %v cannot be greater than 255.",
								i,
							),
						)
					}
					ipArr = append(ipArr, byte(element))
				default:
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"net.ipOtherList: list element at index %v must be an integer.",
							i,
						),
					)
				}
			}
			return NewLoxIPAddress(net.IP(ipArr)), nil
		}
		return argMustBeType(in.callToken, "ipOtherList", "list")
	})
	netFunc("ipv4", 4, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.ipv4' must be an integer.")
		}
		if _, ok := args[1].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.ipv4' must be an integer.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.ipv4' must be an integer.")
		}
		if _, ok := args[3].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Fourth argument to 'net.ipv4' must be an integer.")
		}
		const ipv4Len = 4
		argNums := [ipv4Len]int64{
			args[0].(int64),
			args[1].(int64),
			args[2].(int64),
			args[3].(int64),
		}
		posWords := [ipv4Len]string{
			"First",
			"Second",
			"Third",
			"Fourth",
		}
		var bytes [ipv4Len]byte
		for i, argNum := range argNums {
			if argNum < 0 {
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"%v argument to 'net.ipv4' cannot be negative.",
						posWords[i],
					),
				)
			}
			if argNum > 255 {
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"%v argument to 'net.ipv4' cannot be greater than 255.",
						posWords[i],
					),
				)
			}
			bytes[i] = byte(argNum)
		}
		return NewLoxIPAddressTo4(
			net.IPv4(
				bytes[0],
				bytes[1],
				bytes[2],
				bytes[3],
			),
		), nil
	})
	netFunc("ipv4Buf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxBuffer, ok := args[0].(*LoxBuffer); ok {
			elements := loxBuffer.elements
			if len(elements) != 4 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'net.ipv4Buf' must be a buffer of length 4.")
			}
			return NewLoxIPAddressTo4(
				net.IPv4(
					byte(elements[0].(int64)),
					byte(elements[1].(int64)),
					byte(elements[2].(int64)),
					byte(elements[3].(int64)),
				),
			), nil
		}
		return argMustBeType(in.callToken, "ipv4Buf", "buffer")
	})
	netFunc("ipv4Iter", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			const ipv4Len = 4
			numElementsErr := func(sign string) (any, error) {
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"net.ipv4Iter: iterable must contain exactly %v elements and not %v than %v.",
						ipv4Len,
						sign,
						ipv4Len,
					),
				)
			}
			posWords := [ipv4Len]string{
				"first",
				"second",
				"third",
				"fourth",
			}
			var bytes [ipv4Len]byte
			var i uint8 = 0
			it := iterable.Iterator()
			for {
				if !it.HasNext() {
					if i <= ipv4Len-1 {
						return numElementsErr("less")
					}
					break
				}
				if i >= ipv4Len {
					return numElementsErr("greater")
				}
				element := it.Next()
				switch element := element.(type) {
				case int64:
					if element < 0 {
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipv4Iter: %v iterable element cannot be negative.",
								posWords[i],
							),
						)
					}
					if element > 255 {
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipv4Iter: %v iterable element cannot be greater than 255.",
								posWords[i],
							),
						)
					}
					bytes[i] = byte(element)
				default:
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"net.ipv4Iter: iterable element #%v must be an integer.",
							i+1,
						),
					)
				}
				i++
			}
			return NewLoxIPAddressTo4(
				net.IPv4(
					bytes[0],
					bytes[1],
					bytes[2],
					bytes[3],
				),
			), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("net.ipv4Iter: type '%v' is not iterable.", getType(args[0])))
	})
	netFunc("ipv4List", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			elements := loxList.elements
			if len(elements) != 4 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'net.ipv4List' must be a list of length 4.")
			}
			var bytes [4]byte
			for i, element := range elements {
				switch element := element.(type) {
				case int64:
					if element < 0 {
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"List element at index %v in 'net.ipv4List' cannot be negative.",
								i,
							),
						)
					}
					if element > 255 {
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"List element at index %v in 'net.ipv4List' cannot be greater than 255.",
								i,
							),
						)
					}
					bytes[i] = byte(element)
				default:
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"List element at index %v in 'net.ipv4List' must be an integer.",
							i,
						),
					)
				}
			}
			return NewLoxIPAddressTo4(
				net.IPv4(
					bytes[0],
					bytes[1],
					bytes[2],
					bytes[3],
				),
			), nil
		}
		return argMustBeType(in.callToken, "ipv4List", "list")
	})
	netFunc("ipv6", 16, func(in *Interpreter, args list.List[any]) (any, error) {
		suffix := func(num int) string {
			switch num {
			case 1:
				return "st"
			case 2:
				return "nd"
			case 3:
				return "rd"
			default:
				return "th"
			}
		}
		ipArr := make([]byte, 16)
		for i, arg := range args {
			pos := i + 1
			switch argNum := arg.(type) {
			case int64:
				if argNum < 0 {
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"%v%v argument to 'net.ipv6' cannot be negative.",
							pos,
							suffix(pos),
						),
					)
				}
				if argNum > 255 {
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"%v%v argument to 'net.ipv6' cannot be greater than 255.",
							pos,
							suffix(pos),
						),
					)
				}
				ipArr[i] = byte(argNum)
			default:
				ipArr = nil
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"%v%v argument to 'net.ipv6' must be an integer.",
						pos,
						suffix(pos),
					),
				)
			}
		}
		return NewLoxIPAddress(net.IP(ipArr)), nil
	})
	netFunc("ipv6Buf", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxBuffer, ok := args[0].(*LoxBuffer); ok {
			elements := loxBuffer.elements
			if len(elements) != 16 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'net.ipv4Buf' must be a buffer of length 16.")
			}
			return NewLoxIPAddress(net.IP([]byte{
				byte(elements[0].(int64)),
				byte(elements[1].(int64)),
				byte(elements[2].(int64)),
				byte(elements[3].(int64)),
				byte(elements[4].(int64)),
				byte(elements[5].(int64)),
				byte(elements[6].(int64)),
				byte(elements[7].(int64)),
				byte(elements[8].(int64)),
				byte(elements[9].(int64)),
				byte(elements[10].(int64)),
				byte(elements[11].(int64)),
				byte(elements[12].(int64)),
				byte(elements[13].(int64)),
				byte(elements[14].(int64)),
				byte(elements[15].(int64)),
			})), nil
		}
		return argMustBeType(in.callToken, "ipv6Buf", "buffer")
	})
	netFunc("ipv6Iter", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if iterable, ok := args[0].(interfaces.Iterable); ok {
			const ipv6Len = 16
			numElementsErr := func(sign string) (any, error) {
				return nil, loxerror.RuntimeError(
					in.callToken,
					fmt.Sprintf(
						"net.ipv6Iter: iterable must contain exactly %v elements and not %v than %v.",
						ipv6Len,
						sign,
						ipv6Len,
					),
				)
			}
			suffix := func(num uint8) string {
				switch num {
				case 1:
					return "st"
				case 2:
					return "nd"
				case 3:
					return "rd"
				default:
					return "th"
				}
			}
			ipArr := make([]byte, ipv6Len)
			var i uint8 = 0
			it := iterable.Iterator()
			for {
				if !it.HasNext() {
					if i <= ipv6Len-1 {
						ipArr = nil
						return numElementsErr("less")
					}
					break
				}
				if i >= ipv6Len {
					ipArr = nil
					return numElementsErr("greater")
				}
				pos := i + 1
				element := it.Next()
				switch element := element.(type) {
				case int64:
					if element < 0 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipv6Iter: %v%v iterable element cannot be negative.",
								pos,
								suffix(pos),
							),
						)
					}
					if element > 255 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"net.ipv6Iter: %v%v iterable element cannot be greater than 255.",
								pos,
								suffix(pos),
							),
						)
					}
					ipArr[i] = byte(element)
				default:
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"net.ipv6Iter: iterable element #%v must be an integer.",
							i+1,
						),
					)
				}
				i++
			}
			return NewLoxIPAddress(net.IP(ipArr)), nil
		}
		return nil, loxerror.RuntimeError(in.callToken,
			fmt.Sprintf("net.ipv6Iter: type '%v' is not iterable.", getType(args[0])))
	})
	netFunc("ipv6List", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxList, ok := args[0].(*LoxList); ok {
			elements := loxList.elements
			if len(elements) != 16 {
				return nil, loxerror.RuntimeError(in.callToken,
					"Argument to 'net.ipv6List' must be a list of length 16.")
			}
			ipArr := make([]byte, 16)
			for i, element := range elements {
				switch element := element.(type) {
				case int64:
					if element < 0 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"List element at index %v in 'net.ipv6List' cannot be negative.",
								i,
							),
						)
					}
					if element > 255 {
						ipArr = nil
						return nil, loxerror.RuntimeError(
							in.callToken,
							fmt.Sprintf(
								"List element at index %v in 'net.ipv6List' cannot be greater than 255.",
								i,
							),
						)
					}
					ipArr[i] = byte(element)
				default:
					ipArr = nil
					return nil, loxerror.RuntimeError(
						in.callToken,
						fmt.Sprintf(
							"List element at index %v in 'net.ipv6List' must be an integer.",
							i,
						),
					)
				}
			}
			return NewLoxIPAddress(net.IP(ipArr)), nil
		}
		return argMustBeType(in.callToken, "ipv6List", "list")
	})
	netFunc("joinHostPort", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.joinHostPort' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.joinHostPort' must be a string.")
		}
		host := args[0].(*LoxString).str
		port := args[1].(*LoxString).str
		return NewLoxStringQuote(net.JoinHostPort(host, port)), nil
	})
	netFunc("listen", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.listen' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.listen' must be a string.")
		}
		network := args[0].(*LoxString).str
		address := args[1].(*LoxString).str
		listener, err := net.Listen(network, address)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxListener(listener), nil
	})
	netFunc("listenIP", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.listenIP' must be a string.")
		}
		if _, ok := args[1].(*LoxIPAddress); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.listenIP' must be an IP address instance.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.listenIP' must be an integer.")
		}
		address := args[1].(*LoxIPAddress)
		if address.isNil() {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.listenIP' cannot be a nil IP address instance.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Port number argument to 'net.listenIP' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address.ip.String() + ":" + fmt.Sprint(portNum)
		listener, err := net.Listen(network, addressStr)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxListener(listener), nil
	})
	netFunc("listenPort", 3, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.listenPort' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.listenPort' must be a string.")
		}
		if _, ok := args[2].(int64); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.listenPort' must be an integer.")
		}
		address := args[1].(*LoxString).str
		if strings.Contains(address, ":") {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.listenPort' cannot contain the character ':'.")
		}
		portNum := args[2].(int64)
		if !isValidPortNum(portNum) {
			return nil, loxerror.RuntimeError(in.callToken,
				"Third argument to 'net.listenPort' must be from 0 to 65535.")
		}
		network := args[0].(*LoxString).str
		addressStr := address + ":" + fmt.Sprint(portNum)
		listener, err := net.Listen(network, addressStr)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return NewLoxListener(listener), nil
	})
	netFunc("lookupAddr", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		var addrStr string
		switch arg := args[0].(type) {
		case *LoxString:
			addrStr = arg.str
		case *LoxIPAddress:
			if arg.isNil() {
				return nil, loxerror.RuntimeError(in.callToken,
					"IP address argument to 'net.lookupAddr' cannot be nil.")
			}
			addrStr = arg.ip.String()
		default:
			return argMustBeType(in.callToken, "lookupAddr", "string or IP address instance")
		}
		addrs, err := net.LookupAddr(addrStr)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		addrsList := list.NewListCap[any](int64(len(addrs)))
		for _, addr := range addrs {
			addrsList.Add(NewLoxStringQuote(addr))
		}
		return NewLoxList(addrsList), nil
	})
	netFunc("lookupCNAME", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			cname, err := net.LookupCNAME(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			return NewLoxStringQuote(cname), nil
		}
		return argMustBeType(in.callToken, "lookupCNAME", "string")
	})
	netFunc("lookupHost", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			addrs, err := net.LookupHost(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			addrsList := list.NewListCap[any](int64(len(addrs)))
			for _, addr := range addrs {
				addrsList.Add(NewLoxStringQuote(addr))
			}
			return NewLoxList(addrsList), nil
		}
		return argMustBeType(in.callToken, "lookupHost", "string")
	})
	netFunc("lookupIP", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			ipAddrs, err := net.LookupIP(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			ipAddrsList := list.NewListCap[any](int64(len(ipAddrs)))
			for _, ipAddr := range ipAddrs {
				ipAddrsList.Add(NewLoxIPAddressTo4(ipAddr))
			}
			return NewLoxList(ipAddrsList), nil
		}
		return argMustBeType(in.callToken, "lookupIP", "string")
	})
	netFunc("lookupMX", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			records, err := net.LookupMX(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			recordsList := list.NewListCap[any](int64(len(records)))
			for _, record := range records {
				innerList := list.NewListCap[any](2)
				innerList.Add(NewLoxStringQuote(record.Host))
				innerList.Add(int64(record.Pref))
				recordsList.Add(NewLoxList(innerList))
			}
			return NewLoxList(recordsList), nil
		}
		return argMustBeType(in.callToken, "lookupMX", "string")
	})
	netFunc("lookupNS", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			records, err := net.LookupNS(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			recordsList := list.NewListCap[any](int64(len(records)))
			for _, record := range records {
				recordsList.Add(NewLoxStringQuote(record.Host))
			}
			return NewLoxList(recordsList), nil
		}
		return argMustBeType(in.callToken, "lookupNS", "string")
	})
	netFunc("lookupPort", 2, func(in *Interpreter, args list.List[any]) (any, error) {
		if _, ok := args[0].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"First argument to 'net.lookupPort' must be a string.")
		}
		if _, ok := args[1].(*LoxString); !ok {
			return nil, loxerror.RuntimeError(in.callToken,
				"Second argument to 'net.lookupPort' must be a string.")
		}
		network := args[0].(*LoxString).str
		service := args[1].(*LoxString).str
		port, err := net.LookupPort(network, service)
		if err != nil {
			return nil, loxerror.RuntimeError(in.callToken, err.Error())
		}
		return int64(port), nil
	})
	netFunc("lookupTXT", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			records, err := net.LookupTXT(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			recordsList := list.NewListCap[any](int64(len(records)))
			for _, record := range records {
				recordsList.Add(NewLoxStringQuote(record))
			}
			return NewLoxList(recordsList), nil
		}
		return argMustBeType(in.callToken, "lookupTXT", "string")
	})
	netFunc("mustParseIP", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			ipAddr := net.ParseIP(loxStr.str)
			if ipAddr == nil {
				return nil, loxerror.RuntimeError(in.callToken,
					"Failed to parse IP address '"+loxStr.str+"'.")
			}
			return NewLoxIPAddressTo4(ipAddr), nil
		}
		return argMustBeType(in.callToken, "mustParseIP", "string")
	})
	netFunc("nilIP", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		return NewLoxIPAddress(nil), nil
	})
	netFunc("parseCIDR", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			ipAddr, ipNet, err := net.ParseCIDR(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultList := list.NewListCap[any](2)
			resultList.Add(NewLoxStringQuote(fmt.Sprint(ipAddr)))
			resultList.Add(NewLoxStringQuote(fmt.Sprint(ipNet)))
			return NewLoxList(resultList), nil
		}
		return argMustBeType(in.callToken, "parseCIDR", "string")
	})
	netFunc("parseIP", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			ipAddr := net.ParseIP(loxStr.str)
			return NewLoxIPAddressTo4(ipAddr), nil
		}
		return argMustBeType(in.callToken, "parseIP", "string")
	})
	netFunc("parseIPOrNil", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			ipAddr := net.ParseIP(loxStr.str)
			if ipAddr == nil {
				return nil, nil
			}
			return NewLoxIPAddressTo4(ipAddr), nil
		}
		return argMustBeType(in.callToken, "parseIPOrNil", "string")
	})
	netFunc("parseMAC", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			hwAddr, err := net.ParseMAC(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			buffer := EmptyLoxBufferCap(int64(len(hwAddr)))
			for _, value := range hwAddr {
				addErr := buffer.add(int64(value))
				if addErr != nil {
					return nil, loxerror.RuntimeError(in.callToken, addErr.Error())
				}
			}
			return buffer, nil
		}
		return argMustBeType(in.callToken, "parseMAC", "string")
	})
	netFunc("pipe", 0, func(_ *Interpreter, _ list.List[any]) (any, error) {
		conn1, conn2 := net.Pipe()
		connList := list.NewListCap[any](2)
		connList.Add(NewLoxConnection(conn1))
		connList.Add(NewLoxConnection(conn2))
		return NewLoxList(connList), nil
	})
	netFunc("splitHostPort", 1, func(in *Interpreter, args list.List[any]) (any, error) {
		if loxStr, ok := args[0].(*LoxString); ok {
			host, port, err := net.SplitHostPort(loxStr.str)
			if err != nil {
				return nil, loxerror.RuntimeError(in.callToken, err.Error())
			}
			resultList := list.NewListCap[any](2)
			resultList.Add(NewLoxStringQuote(host))
			resultList.Add(NewLoxStringQuote(port))
			return NewLoxList(resultList), nil
		}
		return argMustBeType(in.callToken, "splitHostPort", "string")
	})

	i.globals.Define(className, netClass)
}
