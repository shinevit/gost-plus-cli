package lazy_test

import (
	"sync"
	"testing"

	"github.com/go-gost/gost.plus/utils/fp/lazy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLazy_Get(t *testing.T) {
	// Test that the value is initialized only once
	callCount := 0
	l := lazy.NewLazy(func() int {
		callCount++
		return 42
	})

	// First call should initialize the value
	val1 := l.Get()
	assert.Equal(t, 42, val1)
	assert.Equal(t, 1, callCount)

	// Subsequent calls should return the same value without reinitializing
	val2 := l.Get()
	assert.Equal(t, 42, val2)
	assert.Equal(t, 1, callCount)
}

func TestLazy_GetPtr(t *testing.T) {
	// Test that the pointer is consistent and value is initialized only once
	callCount := 0
	l := lazy.NewLazy(func() string {
		callCount++
		return "test"
	})

	// First call should initialize the value
	ptr1 := l.GetPtr()
	assert.Equal(t, "test", *ptr1)
	assert.Equal(t, 1, callCount)

	// Subsequent calls should return the same pointer without reinitializing
	ptr2 := l.GetPtr()
	assert.Same(t, ptr1, ptr2)
	assert.Equal(t, 1, callCount)
}

func TestLazy_ConcurrentAccess(t *testing.T) {
	// Test that concurrent access works correctly and initializer is called only once
	const numRoutines = 100
	var callCount int
	l := lazy.NewLazy(func() int {
		callCount++
		return 100
	})

	var wg sync.WaitGroup
	results := make(chan int, numRoutines)

	// Start multiple goroutines that access the lazy value
	for range numRoutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- l.Get()
		}()
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Verify all goroutines got the same value
	for val := range results {
		assert.Equal(t, 100, val)
	}

	// Verify initializer was called exactly once
	assert.Equal(t, 1, callCount)
}

func TestLazy_NilInitializer(t *testing.T) {
	// Test that nil initializer panics
	assert.Panics(t, func() {
		l := &lazy.Lazy[int]{}
		_ = l.Get()
	}, "should panic when initializer is nil")
}

func TestLazy_ZeroValue(t *testing.T) {
	// Test that zero value is handled correctly
	l := lazy.NewLazy(func() int {
		return 0
	})
	val := l.Get()
	assert.Equal(t, 0, val)
}

func TestLazy_SameInstanceSamePointer(t *testing.T) {
	// Test that the same Lazy instance returns the same pointer on multiple GetPtr() calls
	l := lazy.NewLazy(func() []string {
		return []string{"initial", "values"}
	})

	// Get the value multiple times
	val1 := l.Get()
	val2 := l.Get()
	val3 := l.Get()

	// All values should be the same slice
	assert.Equal(t, []string{"initial", "values"}, val1, "Values should match")
	assert.Equal(t, val1, val2, "First and second values should be equal")
	assert.Equal(t, val1, val3, "First and third values should be equal")

	// Get the pointer multiple times
	ptr1 := l.GetPtr()
	ptr2 := l.GetPtr()
	ptr3 := l.GetPtr()

	// All pointers should be the same
	assert.Same(t, ptr1, ptr2, "First and second pointers should be the same")
	assert.Same(t, ptr1, ptr3, "First and third pointers should be the same")
	assert.Same(t, ptr2, ptr3, "Second and third pointers should be the same")

	// Modify the slice through the first pointer
	(*ptr1)[0] = "modified"

	// All pointers should reflect the change
	assert.Equal(t, "modified", (*ptr2)[0], "Second pointer should see the modification")
	assert.Equal(t, "modified", (*ptr3)[0], "Third pointer should see the modification")
}

func TestNewLazy(t *testing.T) {
	// Test that NewLazy returns a non-nil Lazy instance
	initializer := func() string { return "test" }
	l := lazy.NewLazy(initializer)
	require.NotNil(t, l)
}
