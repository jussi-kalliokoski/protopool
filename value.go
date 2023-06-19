package protopool

// Resetter represents a value that can be reset.
type Resetter interface {
	Reset()
}

// Value is a container for allocations, to be used within a pool value.
type Value struct {
	messages stores
	lists    stores
	maps     stores
}

// Reset resets the Value, allowing all allocated objects to be reused from
// there on.
func (v *Value) Reset() {
	v.messages.Reset()
	v.lists.Reset()
	v.maps.Reset()
}

// New returns a new pointer to T, or reuses an existing one from before the
// previous Reset.
//
// If *T implements Resetter, Reset will be called on the value before it's
// returned. Otherwise, the value will be zeroed.
func New[T any](v *Value) *T {
	return getStore[T](&v.messages).get()
}

// NewList returns a new List wrapper that will be modifying the slice at dst.
//
// If an existing list wrapper from before the previous Reset for the same type
// T is available, that will be used instead of allocating a new List.
func NewList[T any](v *Value, dst *[]T) *List[T] {
	list := getStore[List[T]](&v.lists).get()
	list.setTarget(dst)
	return list
}

// List is a wrapper for slices.
type List[T any] struct {
	storage []T
	target  *[]T
}

// Append appends a value to the underlying slice, and sets the slice at dst to
// the same value.
func (l *List[T]) Append(values ...T) {
	l.storage = append(l.storage, values...)
	l.flush()
}

// Reset sets the length of the underlying slice to zero. Has no effect on the
// slice pointed by dst.
func (l *List[T]) Reset() {
	*l = List[T]{storage: l.storage[:0]}
}

func (l *List[T]) setTarget(target *[]T) {
	l.storage = l.storage[:0]
	l.target = target
	l.flush()
}

func (l *List[T]) flush() {
	*l.target = l.storage
}

// Put puts a value v by the key k to the map at m.
//
// If the map is nil, a new one is allocated first (or an existing one from
// before previous Reset), and set to m.
func Put[K comparable, V any](v *Value, m *map[K]V, key K, value V) {
	if *m == nil {
		mw := getStore[mapWrapper[K, V]](&v.maps).get()
		if mw.m == nil {
			mw.m = map[K]V{}
		}
		*m = mw.m
	}
	mm := *m
	mm[key] = value
}

type mapWrapper[K comparable, V any] struct {
	m map[K]V
}

func (mw *mapWrapper[K, V]) Reset() {
	for k := range mw.m {
		delete(mw.m, k)
	}
}

type stores struct {
	byType map[any]Resetter
}

func (ss *stores) Reset() {
	for _, v := range ss.byType {
		v.Reset()
	}
}

func getStore[T any](ss *stores) *store[T] {
	if ss.byType == nil {
		ss.byType = map[any]Resetter{}
	}
	var t *T
	s, _ := ss.byType[t].(*store[T])
	if s == nil {
		s = &store[T]{}
		ss.byType[t] = s
	}
	return s
}

type store[T any] struct {
	messages []*T
	offset   int
}

func (s *store[T]) get() *T {
	if s.offset >= len(s.messages) {
		v := newVal[T]()
		s.messages = append(s.messages, v)
		s.offset++
		return v
	}

	v := s.messages[s.offset]
	s.offset++

	if resetter, ok := any(v).(Resetter); ok {
		resetter.Reset()
	} else {
		var empty T
		*v = empty
	}

	return v
}

func (s *store[T]) Reset() {
	s.offset = 0
}

func newVal[T any]() *T {
	var empty T
	return &empty
}
