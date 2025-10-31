package slice

import (
	"github.com/go-gost/gost.plus/utils/fp"
)

func Map[T any, R any](input []T, fn fp.Functor[T, R]) []R {
	var result []R = []R{}
	for _, item := range input {
		mapped := fn(item)
		result = append(result, mapped)
	}
	return result
}

// Returns: T[T[]] -> R[]
// Always returns an empty slice (not nil) for empty input to maintain consistency
func FlatMap[T any, R any](input []T, mapper fp.Monad[T, R]) []R {
	var result []R = []R{}
	for _, item := range input {
		mapped := mapper(item)
		result = append(result, mapped...)
	}
	return result
}

// Filter items with a condition defined on the predicate
// Always returns an empty slice (not nil) for empty input to maintain consistency
func Filter[T any](input []T, predicate fp.Predicate[T]) []T {
	var result []T = []T{}
	for _, item := range input {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Reversed Filter - filter all items except satisfied by the predicate
// Always returns an empty slice (not nil) for empty input to maintain consistency
func FilterNot[T any](input []T, predicate fp.Predicate[T]) []T {
	var result []T = []T{}
	for _, item := range input {
		if !predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Calculate number of items satisfied with a predicate function
func Count[T any](input []T, predicate fp.Predicate[T]) uint64 {
	var counter uint64
	for _, item := range input {
		if predicate(item) {
			counter++
		}
	}
	return counter
}

// Calculate sum of numeric projection values
func Sum[T any, R fp.Numeric](input []T, selector fp.NumericSelector[T, R]) R {
	var sum R
	for _, item := range input {
		sum += selector(item)
	}
	return sum
}
