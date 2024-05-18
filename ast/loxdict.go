package ast

type LoxDict struct {
	entries map[any]any
	methods map[string]*struct{ ProtoLoxCallable }
}

func NewLoxDict(entries map[any]any) *LoxDict {
	return &LoxDict{
		entries: entries,
		methods: make(map[string]*struct{ ProtoLoxCallable }),
	}
}

func EmptyLoxDict() *LoxDict {
	return NewLoxDict(make(map[any]any))
}

func (l *LoxDict) String() string {
	return getResult(l, true)
}
