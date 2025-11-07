package option

import "github.com/go-gost/gost.plus/utils/fp"

type Option[T any] struct {
	value *T
}
type SomeFn[T any] func(*T)
type NoneFn[T any] func()

func Some[T any](v *T) Option[T] {
	if v == nil {
		return None[T]()
	}
	return Option[T]{value: v}
}

func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

func (opt Option[T]) OrElse(fallback Option[T]) Option[T] {
	return Fold(opt, fallback, func(v T) Option[T] {
		return opt
	})
}

func (opt Option[T]) GetOrElse(defaultValue T) T {
	return Fold(opt, defaultValue, func(v T) T {
		return v
	})
}

func (opt Option[T]) Match(some SomeFn[T], none NoneFn[T]) {
	if opt.value == nil {
		none()
		return
	}
	some(opt.value)
}

func (opt Option[T]) Filter(predicate fp.Predicate[T]) Option[T] {
	return Fold(opt, None[T](), func(v T) Option[T] {
		if predicate(v) {
			return opt
		}
		return None[T]()
	})
}

func Map[A, B any](opt Option[A], fn fp.Functor[A, B]) Option[B] {
	return Fold(opt, None[B](), func(v A) Option[B] {
		result := fn(*opt.value)
		return Some(&result)
	})
}

func FlatMap[A, B any](opt Option[A], fn fp.Functor[A, Option[B]]) Option[B] {
	return Fold(opt, None[B](), func(v A) Option[B] {
		return fn(*opt.value)
	})
}

func Fold[T, R any](opt Option[T], ifEmpty R, f func(T) R) R {
	if opt.value == nil {
		return ifEmpty
	}
	return f(*opt.value)
}

// Cond returns Some(v) if the condition is true, otherwise None.
func Cond[T any](condition bool, value T) Option[T] {
	if condition {
		return Some(&value)
	}
	return None[T]()
}
