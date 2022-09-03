package hashmap

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/cornelk/hashmap/assert"
)

// TestAPICounter shows how to use the hashmap to count REST server API calls.
func TestAPICounter(t *testing.T) {
	t.Parallel()
	m := New[string, *int64]()

	for i := 0; i < 100; i++ {
		s := fmt.Sprintf("/api%d/", i%4)

		counter := int64(0)
		actual, _ := m.GetOrInsert(s, &counter)
		atomic.AddInt64(actual, 1)
	}

	s := fmt.Sprintf("/api%d/", 0)
	value, ok := m.Get(s)
	assert.True(t, ok)
	assert.Equal(t, 25, *value)
}

func TestExample(t *testing.T) {
	m := New[string, int]()
	m.Set("amount", 123)
	value, ok := m.Get("amount")
	assert.True(t, ok)
	assert.Equal(t, 123, value)
}
