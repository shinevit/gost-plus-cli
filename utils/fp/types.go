package fp

// Transforms a value of type T into a value of type R.
type Functor[T any, R any] func(T) R

// Transforms a value of type T into a slice of R.
type Monad[T any, R any] func(T) []R

// Filtering function of condition
type Predicate[T any] = Functor[T, bool]

// Type of numeric values
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}
type NumericSelector[T any, R Numeric] = Functor[T, R]

// Select or transform, T -> R
// Returns the result of applying the function f to the input value
func Map[T any, R any](value T, fn Functor[T, R]) R {
	return fn(value)
}
