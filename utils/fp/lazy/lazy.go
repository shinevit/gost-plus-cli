package lazy

import "sync"

type Lazy[T any] struct {
	once  sync.Once
	value *T
	init  func() T
}

func NewLazy[T any](initializer func() T) *Lazy[T] {
	return &Lazy[T]{init: initializer}
}

func (l *Lazy[T]) GetPtr() *T {
	l.once.Do(func() {
		v := l.init()
		l.value = &v
	})
	return l.value
}

func (l *Lazy[T]) Get() T {
	return *l.GetPtr()
}
