package list

type List[T comparable] []T

func NewList[T comparable]() List[T] {
	return NewListCap[T](0)
}

func NewListCap[T comparable](cap int) List[T] {
	return make(List[T], 0, cap)
}

func (l *List[T]) Add(t T) {
	*l = append(*l, t)
}

func (l *List[T]) AddAt(index int, t T) {
	*l = append((*l)[:index+1], (*l)[index:]...)
	(*l)[index] = t
}

func (l *List[T]) Clear() {
	*l = nil
}

func (l *List[T]) Contains(value T) bool {
	return l.IndexOf(value) >= 0
}

func (l *List[T]) IndexOf(value T) int {
	for i, e := range *l {
		if e == value {
			return i
		}
	}
	return -1
}

func (l *List[T]) Peek() T {
	return (*l)[len(*l)-1]
}

func (l *List[T]) Pop() T {
	return l.RemoveIndex(len(*l) - 1)
}

func (l *List[T]) Push(t T) {
	l.Add(t)
}

func (l *List[T]) RemoveIndex(index int) T {
	data := (*l)[index]
	// dataLen := len(*l)
	// for i := index; i < dataLen-1; i++ {
	// 	(*l)[i] = (*l)[i+1]
	// }
	// *l = (*l)[:dataLen-1]
	*l = append((*l)[:index], (*l)[index+1:]...)
	return data
}

func (l *List[T]) RemoveValue(value T) bool {
	i := l.IndexOf(value)
	if i >= 0 {
		l.RemoveIndex(i)
		return true
	}
	return false
}
