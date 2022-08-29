package hashmap

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()
	m := New[uintptr, uintptr]()
	assert.Zero(t, m.Len())
}

func TestSetString(t *testing.T) {
	t.Parallel()
	m := New[int, string]()
	elephant := "elephant"
	monkey := "monkey"

	m.Set(1, elephant) // insert
	value, ok := m.Get(1)
	require.True(t, ok)
	assert.Equal(t, elephant, value)

	m.Set(1, monkey) // overwrite
	value, ok = m.Get(1)
	require.True(t, ok)
	assert.Equal(t, monkey, value)

	assert.Equal(t, 1, m.Len())

	m.Set(2, elephant) // insert
	assert.Equal(t, 2, m.Len())
	value, ok = m.Get(2)
	require.True(t, ok)
	assert.Equal(t, elephant, value)
}

func TestSetUint8(t *testing.T) {
	t.Parallel()
	m := New[uint8, int]()

	m.Set(1, 128) // insert
	value, ok := m.Get(1)
	require.True(t, ok)
	assert.Equal(t, 128, value)

	m.Set(2, 200) // insert
	assert.Equal(t, 2, m.Len())
	value, ok = m.Get(2)
	require.True(t, ok)
	assert.Equal(t, 200, value)
}

func TestSetUint16(t *testing.T) {
	t.Parallel()
	m := New[uint16, int]()

	m.Set(1, 128) // insert
	value, ok := m.Get(1)
	require.True(t, ok)
	assert.Equal(t, 128, value)

	m.Set(2, 200) // insert
	assert.Equal(t, 2, m.Len())
	value, ok = m.Get(2)
	require.True(t, ok)
	assert.Equal(t, 200, value)
}

func TestSetFloat32(t *testing.T) {
	t.Parallel()
	m := New[float32, int]()

	m.Set(1.1, 128) // insert
	value, ok := m.Get(1.1)
	require.True(t, ok)
	assert.Equal(t, 128, value)

	m.Set(2.2, 200) // insert
	assert.Equal(t, 2, m.Len())
	value, ok = m.Get(2.2)
	require.True(t, ok)
	assert.Equal(t, 200, value)
}

func TestSetInt64(t *testing.T) {
	t.Parallel()
	m := New[int64, int]()

	m.Set(1, 128) // insert
	value, ok := m.Get(1)
	require.True(t, ok)
	assert.Equal(t, 128, value)

	m.Set(2, 200) // insert
	assert.Equal(t, 2, m.Len())
	value, ok = m.Get(2)
	require.True(t, ok)
	assert.Equal(t, 200, value)
}

func TestSetComplex128(t *testing.T) {
	t.Parallel()
	m := New[complex128, int]()

	comp1 := complex(float64(5), float64(10))
	m.Set(comp1, 128) // insert
	value, ok := m.Get(comp1)
	require.True(t, ok)
	assert.Equal(t, 128, value)

	comp2 := complex(float64(5), float64(20))
	m.Set(comp2, 200) // insert
	assert.Equal(t, 2, m.Len())
	value, ok = m.Get(comp2)
	require.True(t, ok)
	assert.Equal(t, 200, value)
}

func TestInsert(t *testing.T) {
	t.Parallel()
	m := New[int, string]()
	elephant := "elephant"
	monkey := "monkey"

	inserted := m.Insert(1, elephant)
	require.True(t, inserted)
	value, ok := m.Get(1)
	require.True(t, ok)
	assert.Equal(t, elephant, value)

	inserted = m.Insert(1, monkey)
	require.False(t, inserted)
	value, ok = m.Get(1)
	require.True(t, ok)
	assert.Equal(t, elephant, value)

	assert.Equal(t, 1, m.Len())

	inserted = m.Insert(2, monkey)
	require.True(t, inserted)
	assert.Equal(t, 2, m.Len())
	value, ok = m.Get(2)
	require.True(t, ok)
	assert.Equal(t, monkey, value)
}

func TestGetNonExistingItem(t *testing.T) {
	t.Parallel()
	m := New[int, string]()
	value, ok := m.Get(1)
	require.False(t, ok)
	assert.Equal(t, "", value)
}

func TestGrow(t *testing.T) {
	t.Parallel()
	m := New[int, string]()
	m.Grow(uintptr(63))

	for { // make sure to wait for resize operation to finish
		if m.resizing.Load() == 0 {
			break
		}
		time.Sleep(time.Microsecond * 50)
	}

	store := m.store.Load()
	log := int(math.Log2(64))
	expectedSize := uintptr(strconv.IntSize - log)
	assert.EqualValues(t, expectedSize, store.keyShifts)
}

func TestResize(t *testing.T) {
	t.Parallel()
	m := NewSized[uintptr, string](2)
	itemCount := uintptr(50)

	for i := uintptr(0); i < itemCount; i++ {
		m.Set(i, strconv.Itoa(int(i)))
	}

	assert.EqualValues(t, itemCount, m.Len())

	for { // make sure to wait for resize operation to finish
		if m.resizing.Load() == 0 {
			break
		}
		time.Sleep(time.Microsecond * 50)
	}

	assert.True(t, m.FillRate() > 30)

	for i := uintptr(0); i < itemCount; i++ {
		value, ok := m.Get(i)
		require.True(t, ok)
		expected := strconv.Itoa(int(i))
		assert.Equal(t, expected, value)
	}
}

func TestStringer(t *testing.T) {
	t.Parallel()
	m := New[int, string]()
	elephant := "elephant"
	monkey := "monkey"

	assert.Equal(t, "[]", m.String())

	m.Set(1, elephant)
	hashedKey0 := m.hasher(1)
	expected := fmt.Sprintf("[%v]", hashedKey0)
	assert.Equal(t, expected, m.String())

	m.Set(2, monkey)
	hashedKey1 := m.hasher(2)
	if hashedKey0 < hashedKey1 {
		expected = fmt.Sprintf("[%v,%v]", hashedKey0, hashedKey1)
	} else {
		expected = fmt.Sprintf("[%v,%v]", hashedKey1, hashedKey0)
	}
	assert.Equal(t, expected, m.String())
}

func TestDelete(t *testing.T) {
	t.Parallel()
	m := New[int, string]()
	elephant := "elephant"
	monkey := "monkey"

	deleted := m.Del(1)
	assert.False(t, deleted)

	m.Set(1, elephant)
	m.Set(2, monkey)

	deleted = m.Del(0)
	assert.False(t, deleted)
	deleted = m.Del(3)
	assert.False(t, deleted)
	assert.Equal(t, 2, m.Len())

	deleted = m.Del(1)
	assert.True(t, deleted)
	deleted = m.Del(1)
	assert.False(t, deleted)
	deleted = m.Del(2)
	assert.True(t, deleted)
	assert.Equal(t, 0, m.Len())
}

func TestRange(t *testing.T) {
	t.Parallel()
	m := New[int, string]()

	items := map[int]string{}
	m.Range(func(key int, value string) bool {
		items[key] = value
		return true
	})
	assert.Equal(t, 0, len(items))

	itemCount := 16
	for i := itemCount; i > 0; i-- {
		m.Set(i, strconv.Itoa(i))
	}

	items = map[int]string{}
	m.Range(func(key int, value string) bool {
		items[key] = value
		return true
	})

	assert.Len(t, items, itemCount)
	for i := 1; i <= itemCount; i++ {
		value, ok := items[i]
		require.True(t, ok)
		expected := strconv.Itoa(i)
		assert.Equal(t, expected, value)
	}

	items = map[int]string{} // test aborting range
	m.Range(func(key int, value string) bool {
		items[key] = value
		return false
	})
	assert.Len(t, items, 1)
}

// nolint: funlen, gocognit
func TestHashMap_parallel(t *testing.T) {
	m := New[int, int]()

	max := 10
	dur := 2 * time.Second

	do := func(t *testing.T, max int, d time.Duration, fn func(*testing.T, int)) <-chan error {
		t.Helper()
		done := make(chan error)
		var times int64
		// This goroutines will terminate test in case if closure hangs.
		go func() {
			for {
				select {
				case <-time.After(d + 500*time.Millisecond):
					if atomic.LoadInt64(&times) == 0 {
						done <- fmt.Errorf("closure was not executed even once, something blocks it")
					}
					close(done)
				case <-done:
				}
			}
		}()
		go func() {
			timer := time.NewTimer(d)
			defer timer.Stop()
		InfLoop:
			for {
				for i := 0; i < max; i++ {
					select {
					case <-timer.C:
						break InfLoop
					default:
					}
					fn(t, i)
					atomic.AddInt64(&times, 1)
				}
			}
			close(done)
		}()
		return done
	}

	wait := func(t *testing.T, done <-chan error) {
		t.Helper()
		if err := <-done; err != nil {
			t.Error(err)
		}
	}

	// Initial fill.
	for i := 0; i < max; i++ {
		m.Set(i, i)
	}
	t.Run("set_get", func(t *testing.T) {
		doneSet := do(t, max, dur, func(t *testing.T, i int) {
			t.Helper()
			m.Set(i, i)
		})
		doneGet := do(t, max, dur, func(t *testing.T, i int) {
			t.Helper()
			if _, ok := m.Get(i); !ok {
				t.Errorf("missing value for key: %d", i)
			}
		})
		wait(t, doneSet)
		wait(t, doneGet)
	})
	t.Run("get-or-insert-and-delete", func(t *testing.T) {
		doneGetOrInsert := do(t, max, dur, func(t *testing.T, i int) {
			t.Helper()
			m.GetOrInsert(i, i)
		})
		doneDel := do(t, max, dur, func(t *testing.T, i int) {
			t.Helper()
			m.Del(i)
		})
		wait(t, doneGetOrInsert)
		wait(t, doneDel)
	})
}

func TestHashMap_SetConcurrent(t *testing.T) {
	t.Parallel()
	m := New[string, int]()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			m.Set(strconv.Itoa(i), i)

			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				m.Get(strconv.Itoa(i))
			}(i)
		}(i)
	}

	wg.Wait()
}

func TestConcurrentInsertDelete(t *testing.T) {
	t.Parallel()

	for i := 0; i < 200; i++ {
		el1 := &ListElement[int, int]{
			key:     111,
			keyHash: 111,
		}
		el2 := &ListElement[int, int]{
			key:     222,
			keyHash: 222,
		}
		el3 := &ListElement[int, int]{
			key:     333,
			keyHash: 333,
		}
		newIl := &ListElement[int, int]{
			key:     223,
			keyHash: 223,
		}
		l := NewList[int, int]()
		l.Add(nil, el1.keyHash, el1.key, 111)
		l.Add(nil, el2.keyHash, el2.key, 222)
		l.Add(nil, el3.keyHash, el3.key, 333)
		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			rand.Seed(int64(time.Now().Nanosecond()))
			time.Sleep(time.Duration(rand.Intn(10)))
			l.Delete(el2)
			wg.Done()
		}()
		go func() {
			defer wg.Done()
			rand.Seed(int64(time.Now().Nanosecond()))
			time.Sleep(time.Duration(rand.Intn(10)))
			for {
				if _, _, inserted := l.Add(nil, newIl.keyHash, newIl.key, 223); inserted {
					return
				}
			}
		}()
		wg.Wait()

		assert.Equal(t, 3, l.Len())
		_, found, _ := l.search(nil, newIl.keyHash, newIl.key)
		assert.NotNil(t, found)
	}
}
