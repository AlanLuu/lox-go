//https://pkg.go.dev/github.com/emirpasic/gods@v1.18.1/queues/priorityqueue

package ast

import (
	"fmt"
	"reflect"

	"github.com/AlanLuu/lox/interfaces"
	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
	pq "github.com/emirpasic/gods/queues/priorityqueue"
	godsutils "github.com/emirpasic/gods/utils"
)

type LoxPriorityQueueElement struct {
	element  any
	priority int64
}

func (l LoxPriorityQueueElement) String() string {
	e := l.element
	p := l.priority
	return fmt.Sprintf(
		"{priority: %v, value: %v}",
		p,
		getResult(e, e, false),
	)
}

func loxPriorityQueue_cmp(a, b any) int {
	p1 := a.(LoxPriorityQueueElement).priority
	p2 := b.(LoxPriorityQueueElement).priority
	return -godsutils.Int64Comparator(p1, p2)
}

func loxPriorityQueue_cmpReversed(a, b any) int {
	return -loxPriorityQueue_cmp(a, b)
}

type LoxPriorityQueueIterator struct {
	iter      *pq.Iterator
	firstIter bool
}

func (l *LoxPriorityQueueIterator) HasNext() bool {
	if l.firstIter {
		l.firstIter = false
		return l.iter.First()
	}
	return l.iter.Next()
}

func (l *LoxPriorityQueueIterator) Next() any {
	pair := list.NewListCap[any](2)
	element := l.iter.Value().(LoxPriorityQueueElement)
	pair.Add(element.element)
	pair.Add(element.priority)
	return NewLoxList(pair)
}

type LoxPriorityQueueDup struct {
	list *list.List[any]
}

type LoxPriorityQueueOpts struct {
	allowDupPriorities bool
	cmp                func(a, b any) int
	isReversed         bool
}

var (
	loxPriorityQueueOpts_def = LoxPriorityQueueOpts{
		cmp: loxPriorityQueue_cmp,
	}
	loxPriorityQueueOpts_reversed = LoxPriorityQueueOpts{
		cmp:        loxPriorityQueue_cmpReversed,
		isReversed: true,
	}
	loxPriorityQueueOpts_allowDup = LoxPriorityQueueOpts{
		cmp:                loxPriorityQueue_cmp,
		allowDupPriorities: true,
	}
	loxPriorityQueueOpts_allowDupReversed = LoxPriorityQueueOpts{
		cmp:                loxPriorityQueue_cmpReversed,
		allowDupPriorities: true,
		isReversed:         true,
	}
)

type LoxPriorityQueue struct {
	queue              *pq.Queue
	allowDupPriorities bool
	isReversed         bool
	priorities         map[int64]any
	methods            map[string]*struct{ ProtoLoxCallable }
}

func NewLoxPriorityQueue(opts LoxPriorityQueueOpts) *LoxPriorityQueue {
	if opts.cmp == nil {
		opts.cmp = loxPriorityQueue_cmp
	}
	return &LoxPriorityQueue{
		queue:              pq.NewWith(opts.cmp),
		allowDupPriorities: opts.allowDupPriorities,
		isReversed:         opts.isReversed,
		priorities:         make(map[int64]any),
		methods:            make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func NewLoxPriorityQueueMap(
	m map[int64]any,
	opts LoxPriorityQueueOpts,
) *LoxPriorityQueue {
	loxPriQueue := NewLoxPriorityQueue(opts)
	for priority, element := range m {
		loxPriQueue.enqueue(element, priority)
	}
	return loxPriQueue
}

func (l *LoxPriorityQueue) clear() {
	l.queue.Clear()
	clear(l.priorities)
}

func (l *LoxPriorityQueue) contains(value any, priority int64) bool {
	result := false
	l.forEachBool(func(v any, p int64, _ int64) bool {
		if value == v {
			result = true
		} else if first, ok := value.(interfaces.Equatable); ok {
			result = first.Equals(v)
		} else if second, ok := v.(interfaces.Equatable); ok {
			result = second.Equals(value)
		} else {
			result = reflect.DeepEqual(v, value)
		}
		result = result && priority == p
		return !result
	})
	return result
}

func (l *LoxPriorityQueue) containsPriority(priority int64) bool {
	_, ok := l.priorities[priority]
	return ok
}

func (l *LoxPriorityQueue) containsValue(value any) bool {
	result := false
	l.forEachBool(func(v any, _ int64, _ int64) bool {
		if value == v {
			result = true
		} else if first, ok := value.(interfaces.Equatable); ok {
			result = first.Equals(v)
		} else if second, ok := v.(interfaces.Equatable); ok {
			result = second.Equals(value)
		} else {
			result = reflect.DeepEqual(v, value)
		}
		return !result
	})
	return result
}

func (l *LoxPriorityQueue) dequeue() (any, int64, bool) {
	value, ok := l.queue.Dequeue()
	if !ok {
		return nil, 0, false
	}
	element := value.(LoxPriorityQueueElement)
	l.dequeuePriorityMap(element.element, element.priority)
	return element.element, element.priority, true
}

func (l *LoxPriorityQueue) dequeuePriorityMap(element any, priority int64) {
	if l.allowDupPriorities {
		mapElement, ok := l.priorities[priority]
		if !ok {
			return
		}
		switch mapElement := mapElement.(type) {
		case LoxPriorityQueueDup:
			lst := mapElement.list
			var i int64 = 0
			for _, e := range *lst {
				if i < 0 {
					return
				}
				if element == e {
					lst.RemoveIndex(i)
					if len(*lst) == 0 {
						delete(l.priorities, priority)
					}
					return
				}
				i++
			}
			if i == 0 {
				delete(l.priorities, priority)
			}
		default:
			delete(l.priorities, priority)
		}
	} else {
		delete(l.priorities, priority)
	}
}

func (l *LoxPriorityQueue) enqueue(element any, priority int64) error {
	if !l.allowDupPriorities {
		if _, ok := l.priorities[priority]; ok {
			return loxerror.Error(
				fmt.Sprintf(
					"priority value '%v' already exists.",
					priority,
				),
			)
		}
	}
	for {
		asserted, ok := element.(LoxPriorityQueueElement)
		if !ok {
			break
		}
		element = asserted.element
	}
	l.queue.Enqueue(LoxPriorityQueueElement{element, priority})
	l.enqueuePriorityMap(element, priority)
	return nil
}

func (l *LoxPriorityQueue) enqueuePriorityMap(element any, priority int64) {
	if l.allowDupPriorities {
		if item, ok := l.priorities[priority]; ok {
			switch item := item.(type) {
			case LoxPriorityQueueDup:
				item.list.Add(element)
			default:
				lst := list.NewListCap[any](2)
				lst.Add(item)
				lst.Add(element)
				l.priorities[priority] = LoxPriorityQueueDup{&lst}
			}
		} else {
			lst := list.NewList[any]()
			lst.Add(element)
			l.priorities[priority] = LoxPriorityQueueDup{&lst}
		}
	} else {
		l.priorities[priority] = element
	}
}

func (l *LoxPriorityQueue) equals(
	other *LoxPriorityQueue,
	checkPriority bool,
) bool {
	if l == other {
		return true
	}
	i1 := l.queue.Iterator()
	i2 := other.queue.Iterator()
	i1First := i1.First()
	i2First := i2.First()
	if i1First != i2First {
		return false
	}
	if !i1First && !i2First {
		return true
	}
	firstIter := true
	for {
		if !firstIter {
			i1Next := i1.Next()
			i2Next := i2.Next()
			if i1Next != i2Next {
				return false
			}
			if !i1Next && !i2Next {
				return true
			}
		}
		e1 := i1.Value().(LoxPriorityQueueElement)
		e2 := i2.Value().(LoxPriorityQueueElement)
		var result bool
		if e1.element == e2.element {
			result = true
		} else if first, ok := e1.element.(interfaces.Equatable); ok {
			result = first.Equals(e2.element)
		} else if second, ok := e2.element.(interfaces.Equatable); ok {
			result = second.Equals(e1.element)
		} else {
			result = reflect.DeepEqual(e1.element, e2.element)
		}
		if checkPriority {
			result = result && e1.priority == e2.priority
		}
		if !result {
			return false
		}
		firstIter = false
	}
}

func (l *LoxPriorityQueue) equalsPriorities(other *LoxPriorityQueue) bool {
	if l == other {
		return true
	}
	if len(l.priorities) != len(other.priorities) {
		return false
	}
	for priority := range l.priorities {
		if _, ok := other.priorities[priority]; !ok {
			return false
		}
	}
	return true
}

func (l *LoxPriorityQueue) equalsStrict(other *LoxPriorityQueue) bool {
	return l.equals(other, true)
}

func (l *LoxPriorityQueue) equalsValues(other *LoxPriorityQueue) bool {
	return l.equals(other, false)
}

func (l *LoxPriorityQueue) forEach(
	f func(element any, priority int64, index int64),
) {
	l.forEachErr(func(element any, priority int64, index int64) error {
		f(element, priority, index)
		return nil
	})
}

func (l *LoxPriorityQueue) forEachBool(
	f func(element any, priority int64, index int64) bool,
) {
	l.forEachErr(func(element any, priority int64, index int64) error {
		if !f(element, priority, index) {
			return loxerror.Error("")
		}
		return nil
	})
}

func (l *LoxPriorityQueue) forEachErr(
	f func(element any, priority int64, index int64) error,
) error {
	it := l.queue.Iterator()
	if !it.First() {
		return nil
	}
	firstIter := true
	for i := int64(0); firstIter || it.Next(); i++ {
		firstIter = false
		element := it.Value().(LoxPriorityQueueElement)
		if err := f(element.element, element.priority, i); err != nil {
			return err
		}
	}
	return nil
}

func (l *LoxPriorityQueue) forEachPrev(
	f func(element any, priority int64, index int64),
) {
	l.forEachPrevErr(func(element any, priority int64, index int64) error {
		f(element, priority, index)
		return nil
	})
}

func (l *LoxPriorityQueue) forEachPrevErr(
	f func(element any, priority int64, index int64) error,
) error {
	it := l.queue.Iterator()
	if !it.Last() {
		return nil
	}
	firstIter := true
	for i := int64(0); firstIter || it.Prev(); i++ {
		firstIter = false
		element := it.Value().(LoxPriorityQueueElement)
		if err := f(element.element, element.priority, i); err != nil {
			return err
		}
	}
	return nil
}

func (l *LoxPriorityQueue) getPriorityByValue(value any) (int64, error) {
	var priority int64
	found := false
	l.forEachBool(func(v any, p int64, _ int64) bool {
		if value == v {
			found = true
		} else if first, ok := value.(interfaces.Equatable); ok {
			found = first.Equals(v)
		} else if second, ok := v.(interfaces.Equatable); ok {
			found = second.Equals(value)
		} else {
			found = reflect.DeepEqual(v, value)
		}
		priority = p
		return !found
	})
	if !found {
		return 0, loxerror.Error(
			fmt.Sprintf(
				"failed to get priority from value '%v'.",
				getResult(value, value, true),
			),
		)
	}
	return priority, nil
}

func (l *LoxPriorityQueue) getValueByPriority(priority int64) (any, error) {
	element, ok := l.priorities[priority]
	if !ok {
		return nil, loxerror.Error(
			fmt.Sprintf(
				"failed to find value with priority '%v'.",
				priority,
			),
		)
	}
	return element, nil
}

func (l *LoxPriorityQueue) isEmpty() bool {
	return l.queue.Empty()
}

func (l *LoxPriorityQueue) peek() (any, int64, bool) {
	value, ok := l.queue.Peek()
	if !ok {
		return nil, 0, false
	}
	element := value.(LoxPriorityQueueElement)
	return element.element, element.priority, true
}

func (l *LoxPriorityQueue) prioritiesListAny() list.List[any] {
	var seenPriorities map[int64]struct{} = nil
	if l.allowDupPriorities {
		seenPriorities = map[int64]struct{}{}
	}
	result := list.NewListCap[any](l.Length())
	if l.isReversed {
		l.forEach(func(_ any, priority int64, _ int64) {
			if l.allowDupPriorities {
				if _, ok := seenPriorities[priority]; !ok {
					result.Add(priority)
					seenPriorities[priority] = struct{}{}
				}
			} else {
				result.Add(priority)
			}
		})
	} else {
		l.forEachPrev(func(_ any, priority int64, _ int64) {
			if l.allowDupPriorities {
				if _, ok := seenPriorities[priority]; !ok {
					result.Add(priority)
					seenPriorities[priority] = struct{}{}
				}
			} else {
				result.Add(priority)
			}
		})
	}
	return result
}

func (l *LoxPriorityQueue) reset() {
	l.setIsReversedReset(false)
}

func (l *LoxPriorityQueue) setAllowDupPriorities(allow bool) {
	l.allowDupPriorities = allow
}

func (l *LoxPriorityQueue) setIsReversedReset(isReversed bool) {
	l.clear()
	l.setAllowDupPriorities(false)
	if isReversed {
		l.queue = pq.NewWith(loxPriorityQueue_cmpReversed)
	} else {
		l.queue = pq.NewWith(loxPriorityQueue_cmp)
	}
	l.isReversed = isReversed
}

func (l *LoxPriorityQueue) values() list.List[any] {
	values := l.queue.Values()
	for i := 0; i < len(values); i++ {
		for {
			asserted, ok := values[i].(LoxPriorityQueueElement)
			if !ok {
				break
			}
			values[i] = asserted.element
		}
	}
	values = values[:len(values):len(values)]
	return values
}

func (l *LoxPriorityQueue) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	priQueueFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native priority queue fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'priority queue.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	argMustBeTypeAn := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'priority queue.%v' must be an %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "add", "enqueue":
		return priQueueFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"First argument to 'priority queue.%v' must be an integer.",
						methodName,
					),
				)
			}
			if err := l.enqueue(args[1], args[0].(int64)); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"priority queue.%v: %v",
						methodName,
						err.Error(),
					),
				)
			}
			return nil, nil
		})
	case "addArgSwitched", "enqueueArgSwitched":
		return priQueueFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"Second argument to 'priority queue.%v' must be an integer.",
						methodName,
					),
				)
			}
			if err := l.enqueue(args[0], args[1].(int64)); err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"priority queue.%v: %v",
						methodName,
						err.Error(),
					),
				)
			}
			return nil, nil
		})
	case "allowDupPriorities":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if allow, ok := args[0].(bool); ok {
				l.setAllowDupPriorities(allow)
				return nil, nil
			}
			return argMustBeType("boolean")
		})
	case "allowsDupPriorities":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.allowDupPriorities, nil
		})
	case "clear":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.clear()
			return nil, nil
		})
	case "contains":
		return priQueueFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[0].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"First argument to 'priority queue.contains' must be an integer.")
			}
			return l.contains(args[1], args[0].(int64)), nil
		})
	case "containsArgSwitched":
		return priQueueFunc(2, func(_ *Interpreter, args list.List[any]) (any, error) {
			if _, ok := args[1].(int64); !ok {
				return nil, loxerror.RuntimeError(name,
					"Second argument to 'priority queue.containsArgSwitched' must be an integer.")
			}
			return l.contains(args[0], args[1].(int64)), nil
		})
	case "containsPriority":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if priority, ok := args[0].(int64); ok {
				return l.containsPriority(priority), nil
			}
			return argMustBeTypeAn("integer")
		})
	case "containsValue":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			return l.containsValue(args[0]), nil
		})
	case "dequeue", "remove":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, priority, ok := l.dequeue()
			if !ok {
				return nil, nil
			}
			pair := list.NewListCap[any](2)
			pair.Add(element)
			pair.Add(priority)
			return NewLoxList(pair), nil
		})
	case "dequeueErr", "removeErr":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, priority, ok := l.dequeue()
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"Cannot remove from empty priority queue.")
			}
			pair := list.NewListCap[any](2)
			pair.Add(element)
			pair.Add(priority)
			return NewLoxList(pair), nil
		})
	case "dequeueIter", "removeIter":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var element any
			var priority int64
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				e, p, ok := l.dequeue()
				if !ok {
					return false
				}
				element = e
				priority = p
				return true
			}
			iterator.nextMethod = func() any {
				pair := list.NewListCap[any](2)
				pair.Add(element)
				pair.Add(priority)
				return NewLoxList(pair)
			}
			return NewLoxIterator(iterator), nil
		})
	case "dequeuePriority", "removePriority":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			_, priority, ok := l.dequeue()
			if !ok {
				return nil, nil
			}
			return priority, nil
		})
	case "dequeuePriorityErr", "removePriorityErr":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			_, priority, ok := l.dequeue()
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"Cannot remove from empty priority queue.")
			}
			return priority, nil
		})
	case "dequeuePriorityIter", "removePriorityIter":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var priority int64
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				_, p, ok := l.dequeue()
				if !ok {
					return false
				}
				priority = p
				return true
			}
			iterator.nextMethod = func() any {
				return priority
			}
			return NewLoxIterator(iterator), nil
		})
	case "dequeueValue", "removeValue":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, _, ok := l.dequeue()
			if !ok {
				return nil, nil
			}
			return element, nil
		})
	case "dequeueValueErr", "removeValueErr":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, _, ok := l.dequeue()
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"Cannot remove from empty priority queue.")
			}
			return element, nil
		})
	case "dequeueValueIter", "removeValueIter":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			var element any
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				e, _, ok := l.dequeue()
				if !ok {
					return false
				}
				element = e
				return true
			}
			iterator.nextMethod = func() any {
				return element
			}
			return NewLoxIterator(iterator), nil
		})
	case "equals":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxPriQueue, ok := args[0].(*LoxPriorityQueue); ok {
				return l.equalsStrict(loxPriQueue), nil
			}
			return argMustBeType("priority queue")
		})
	case "equalsPriorities":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxPriQueue, ok := args[0].(*LoxPriorityQueue); ok {
				return l.equalsPriorities(loxPriQueue), nil
			}
			return argMustBeType("priority queue")
		})
	case "equalsValues":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxPriQueue, ok := args[0].(*LoxPriorityQueue); ok {
				return l.equalsValues(loxPriQueue), nil
			}
			return argMustBeType("priority queue")
		})
	case "forEach":
		return priQueueFunc(1, func(i *Interpreter, args list.List[any]) (any, error) {
			if callback, ok := args[0].(LoxCallable); ok {
				argList := getArgList(callback, 3)
				defer argList.Clear()
				return nil, l.forEachErr(func(element any, priority, index int64) error {
					argList[0] = element
					argList[1] = priority
					argList[2] = index
					result, resultErr := callback.call(i, argList)
					if resultErr != nil && result == nil {
						return resultErr
					}
					return nil
				})
			}
			return argMustBeType("function")
		})
	case "getPriorityByValue":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			priority, err := l.getPriorityByValue(args[0])
			if err != nil {
				return nil, nil
			}
			return priority, nil
		})
	case "getPriorityByValueErr":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			priority, err := l.getPriorityByValue(args[0])
			if err != nil {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"priority queue.getPriorityByValueErr: %v",
						err.Error(),
					),
				)
			}
			return priority, nil
		})
	case "getValueByPriority":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if priority, ok := args[0].(int64); ok {
				element, err := l.getValueByPriority(priority)
				if err != nil {
					return nil, nil
				}
				switch element := element.(type) {
				case LoxPriorityQueueDup:
					lst := *element.list
					if len(lst) == 1 {
						return lst[0], nil
					}
					newList := list.NewListCap[any](int64(len(lst)))
					for _, e := range lst {
						newList.Add(e)
					}
					return NewLoxList(newList), nil
				default:
					return element, nil
				}
			}
			return argMustBeTypeAn("integer")
		})
	case "getValueByPriorityErr":
		return priQueueFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if priority, ok := args[0].(int64); ok {
				element, err := l.getValueByPriority(priority)
				if err != nil {
					return nil, loxerror.RuntimeError(
						name,
						fmt.Sprintf(
							"priority queue.getValueByPriorityErr: %v",
							err.Error(),
						),
					)
				}
				switch element := element.(type) {
				case LoxPriorityQueueDup:
					lst := *element.list
					if len(lst) == 1 {
						return lst[0], nil
					}
					newList := list.NewListCap[any](int64(len(lst)))
					for _, e := range lst {
						newList.Add(e)
					}
					return NewLoxList(newList), nil
				default:
					return element, nil
				}
			}
			return argMustBeTypeAn("integer")
		})
	case "isEmpty":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isEmpty(), nil
		})
	case "isReversed":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return l.isReversed, nil
		})
	case "peek":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, priority, ok := l.peek()
			if !ok {
				return nil, nil
			}
			pair := list.NewListCap[any](2)
			pair.Add(element)
			pair.Add(priority)
			return NewLoxList(pair), nil
		})
	case "peekErr":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, priority, ok := l.peek()
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"Cannot peek from empty priority queue.")
			}
			pair := list.NewListCap[any](2)
			pair.Add(element)
			pair.Add(priority)
			return NewLoxList(pair), nil
		})
	case "peekPriority":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			_, priority, ok := l.peek()
			if !ok {
				return nil, nil
			}
			return priority, nil
		})
	case "peekPriorityErr":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			_, priority, ok := l.peek()
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"Cannot peek from empty priority queue.")
			}
			return priority, nil
		})
	case "peekValue":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, _, ok := l.peek()
			if !ok {
				return nil, nil
			}
			return element, nil
		})
	case "peekValueErr":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			element, _, ok := l.peek()
			if !ok {
				return nil, loxerror.RuntimeError(name,
					"Cannot peek from empty priority queue.")
			}
			return element, nil
		})
	case "print":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			fmt.Println(l.queue)
			return nil, nil
		})
	case "priorities":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			set := EmptyLoxSet()
			for priority := range l.priorities {
				_, errStr := set.add(priority)
				if len(errStr) > 0 {
					return nil, loxerror.RuntimeError(name, errStr)
				}
			}
			return set, nil
		})
	case "prioritiesIter":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			queueIter := l.queue.Iterator()
			firstIter := true
			iterator := ProtoIterator{}
			if l.allowDupPriorities {
				var priority int64
				seenPriorities := map[int64]struct{}{}
				if l.isReversed {
					iterator.hasNextMethod = func() bool {
						for {
							var ok bool
							if firstIter {
								firstIter = false
								ok = queueIter.First()
							} else {
								ok = queueIter.Next()
							}
							if !ok {
								return false
							}
							priority = queueIter.Value().(LoxPriorityQueueElement).priority
							if _, ok = seenPriorities[priority]; ok {
								continue
							}
							seenPriorities[priority] = struct{}{}
							return true
						}
					}
				} else {
					iterator.hasNextMethod = func() bool {
						for {
							var ok bool
							if firstIter {
								firstIter = false
								ok = queueIter.Last()
							} else {
								ok = queueIter.Prev()
							}
							if !ok {
								return false
							}
							priority = queueIter.Value().(LoxPriorityQueueElement).priority
							if _, ok = seenPriorities[priority]; ok {
								continue
							}
							seenPriorities[priority] = struct{}{}
							return true
						}
					}
				}
				iterator.nextMethod = func() any {
					return priority
				}
			} else {
				if l.isReversed {
					iterator.hasNextMethod = func() bool {
						if firstIter {
							firstIter = false
							return queueIter.First()
						}
						return queueIter.Next()
					}
				} else {
					iterator.hasNextMethod = func() bool {
						if firstIter {
							firstIter = false
							return queueIter.Last()
						}
						return queueIter.Prev()
					}
				}
				iterator.nextMethod = func() any {
					return queueIter.Value().(LoxPriorityQueueElement).priority
				}
			}
			return NewLoxIterator(iterator), nil
		})
	case "prioritiesList":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxList(l.prioritiesListAny()), nil
		})
	case "reset":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.reset()
			return nil, nil
		})
	case "resetReversed":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.setIsReversedReset(true)
			return nil, nil
		})
	case "str":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxStringQuote(l.queue.String()), nil
		})
	case "toDict":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			for priority, element := range l.priorities {
				switch element := element.(type) {
				case LoxPriorityQueueDup:
					elements := *element.list
					lst := list.NewListCap[any](int64(len(elements)))
					for _, e := range elements {
						lst.Add(e)
					}
					dict.setKeyValue(priority, NewLoxList(lst))
				default:
					dict.setKeyValue(priority, element)
				}
			}
			return dict, nil
		})
	case "toList":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			pairs := list.NewListCap[any](l.Length())
			l.forEach(func(element any, priority int64, _ int64) {
				pair := list.NewListCap[any](2)
				pair.Add(element)
				pair.Add(priority)
				pairs.Add(NewLoxList(pair))
			})
			return NewLoxList(pairs), nil
		})
	case "values":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxList(l.values()), nil
		})
	case "valuesIter":
		return priQueueFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			queueIter := l.queue.Iterator()
			firstIter := true
			iterator := ProtoIterator{}
			iterator.hasNextMethod = func() bool {
				if firstIter {
					firstIter = false
					return queueIter.First()
				}
				return queueIter.Next()
			}
			iterator.nextMethod = func() any {
				return queueIter.Value().(LoxPriorityQueueElement).element
			}
			return NewLoxIterator(iterator), nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Priority queues have no property called '"+methodName+"'.")
}

func (l *LoxPriorityQueue) Iterator() interfaces.Iterator {
	iter := l.queue.Iterator()
	return &LoxPriorityQueueIterator{
		iter:      &iter,
		firstIter: true,
	}
}

func (l *LoxPriorityQueue) Length() int64 {
	return int64(l.queue.Size())
}

func (l *LoxPriorityQueue) String() string {
	str := "priority queue"
	if l.isReversed || l.allowDupPriorities {
		str += " ("
		if l.isReversed {
			str += "reversed"
		}
		if l.allowDupPriorities {
			if l.isReversed {
				str += ", "
			}
			str += "duplicate priorities"
		}
		str += ")"
	}
	return fmt.Sprintf("<%v at %p>", str, l)
}

func (l *LoxPriorityQueue) Type() string {
	return "priority queue"
}

type LoxPriorityQueueBuilder struct {
	opts    LoxPriorityQueueOpts
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxPriorityQueueBuilder() *LoxPriorityQueueBuilder {
	return &LoxPriorityQueueBuilder{
		opts:    loxPriorityQueueOpts_def,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func (l *LoxPriorityQueueBuilder) Get(name *token.Token) (any, error) {
	methodName := name.Lexeme
	if method, ok := l.methods[methodName]; ok {
		return method, nil
	}
	builderFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native priority queue builder fn %v at %p>", methodName, s)
		}
		if _, ok := l.methods[methodName]; !ok {
			l.methods[methodName] = s
		}
		return s, nil
	}
	argMustBeType := func(theType string) (any, error) {
		errStr := fmt.Sprintf("Argument to 'priority queue builder.%v' must be a %v.", methodName, theType)
		return nil, loxerror.RuntimeError(name, errStr)
	}
	switch methodName {
	case "allowDupPriorities":
		return builderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if allowDupPriorities, ok := args[0].(bool); ok {
				l.opts.allowDupPriorities = allowDupPriorities
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "build":
		return builderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			return NewLoxPriorityQueue(l.opts), nil
		})
	case "buildArgs":
		return builderFunc(-1, func(_ *Interpreter, args list.List[any]) (any, error) {
			argsLen := len(args)
			if argsLen < 2 {
				return nil, loxerror.RuntimeError(
					name,
					fmt.Sprintf(
						"priority queue builder.buildArgs: expected at least 2 arguments but got %v.",
						argsLen,
					),
				)
			}
			if argsLen%2 != 0 {
				return nil, loxerror.RuntimeError(name,
					"priority queue builder.buildArgs: number of arguments cannot be an odd number.")
			}
			m := map[int64]any{}
			var intVar int64
			for i, arg := range args {
				if i%2 == 0 {
					switch arg := arg.(type) {
					case int64:
						intVar = arg
					default:
						return nil, loxerror.RuntimeError(
							name,
							fmt.Sprintf(
								"priority queue builder.buildArgs: argument #%v must be an integer.",
								i+1,
							),
						)
					}
				} else {
					m[intVar] = arg
				}
			}
			return NewLoxPriorityQueueMap(m, l.opts), nil
		})
	case "buildDict":
		return builderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if loxDict, ok := args[0].(*LoxDict); ok {
				m := map[int64]any{}
				it := loxDict.Iterator()
				for it.HasNext() {
					pair := it.Next().(*LoxList).elements
					var key int64
					switch pairKey := pair[0].(type) {
					case int64:
						key = pairKey
					default:
						return nil, loxerror.RuntimeError(name,
							"Dictionary argument to 'priority queue builder.buildDict' must only have integer keys.")
					}
					value := pair[1]
					m[key] = value
				}
				return NewLoxPriorityQueueMap(m, l.opts), nil
			}
			return argMustBeType("dictionary")
		})
	case "isReversed":
		return builderFunc(1, func(_ *Interpreter, args list.List[any]) (any, error) {
			if isReversed, ok := args[0].(bool); ok {
				if isReversed {
					l.opts.cmp = loxPriorityQueue_cmpReversed
				} else {
					l.opts.cmp = loxPriorityQueue_cmp
				}
				l.opts.isReversed = isReversed
				return l, nil
			}
			return argMustBeType("boolean")
		})
	case "toDict":
		return builderFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			dict := EmptyLoxDict()
			dict.setKeyValue(
				NewLoxString("allowDupPriorities", '\''),
				l.opts.allowDupPriorities,
			)
			dict.setKeyValue(
				NewLoxString("isReversed", '\''),
				l.opts.isReversed,
			)
			return dict, nil
		})
	}
	return nil, loxerror.RuntimeError(name, "Priority queue builders have no property called '"+methodName+"'.")
}

func (l *LoxPriorityQueueBuilder) String() string {
	return fmt.Sprintf("<priority queue builder at %p>", l)
}

func (l *LoxPriorityQueueBuilder) Type() string {
	return "priority queue builder"
}
