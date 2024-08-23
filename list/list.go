package list

type List[T any] []T

func NewList[T any]() List[T] {
	return NewListLenCap[T](0, 0)
}

func NewListCap[T any](cap int64) List[T] {
	return NewListLenCap[T](0, cap)
}

func NewListCapDouble[T any](cap int64) List[T] {
	return NewListLenCap[T](0, cap*2)
}

func NewListLen[T any](len int64) List[T] {
	return NewListLenCap[T](len, len)
}

func NewListLenCap[T any](len int64, cap int64) List[T] {
	return make(List[T], len, cap)
}

func (l *List[T]) Add(t T) {
	*l = append(*l, t)
}

func (l *List[T]) AddAt(index int64, t T) {
	if int64(len(*l)) == index {
		l.Add(t)
	} else {
		*l = append((*l)[:index+1], (*l)[index:]...)
		(*l)[index] = t
	}
}

func (l *List[T]) Clear() {
	*l = nil
}

func (l *List[T]) IsEmpty() bool {
	return len(*l) == 0
}

func (l *List[T]) Peek() T {
	return (*l)[len(*l)-1]
}

func (l *List[T]) Pop() T {
	return l.RemoveIndex(int64(len(*l) - 1))
}

func (l *List[T]) Push(t T) {
	l.Add(t)
}

func (l *List[T]) RemoveIndex(index int64) T {
	data := (*l)[index]
	*l = append((*l)[:index], (*l)[index+1:]...)
	return data
}
