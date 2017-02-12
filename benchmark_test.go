package hashmap

import (
	"strconv"
	"sync"
	"testing"
	"unsafe"
)

const benchmarkItemCount = 1 << 9

func setupHashMap(b *testing.B) *HashMap {
	b.StopTimer()
	m := New()

	for i := uint64(0); i < benchmarkItemCount; i++ {
		j := uintptr(i)
		m.Set(i, unsafe.Pointer(&j))
	}

	b.StartTimer()
	return m
}

func setupHashMapString(b *testing.B) *HashMap {
	b.StopTimer()
	m := New()

	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m.Set(s, unsafe.Pointer(&s))
	}

	b.StartTimer()
	return m
}

func setupHashMapHashedKey(b *testing.B) *HashMap {
	b.StopTimer()
	m := New()
	log := log2(uint64(benchmarkItemCount))

	for i := uint64(0); i < benchmarkItemCount; i++ {
		hash := i << (64 - log)
		j := uintptr(i)
		m.SetHashedKey(hash, unsafe.Pointer(&j))
	}

	b.StartTimer()
	return m
}

func setupGoMap(b *testing.B) map[uint64]uint64 {
	b.StopTimer()
	m := make(map[uint64]uint64)
	for i := uint64(0); i < benchmarkItemCount; i++ {
		m[i] = i
	}
	b.StartTimer()
	return m
}

func setupGoMapString(b *testing.B) map[string]string {
	b.StopTimer()
	m := make(map[string]string)
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m[s] = s
	}
	b.StartTimer()
	return m
}

func BenchmarkReadHashMapUint(b *testing.B) {
	m := setupHashMap(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				j, _ := m.GetUintKey(i)
				if *(*uint64)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapString(b *testing.B) {
	m := setupHashMapString(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < benchmarkItemCount; i++ {
				s := strconv.Itoa(i)
				sVal, _ := m.GetStringKey(s)
				if *(*string)(sVal) != s {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapInterface(b *testing.B) {
	m := setupHashMap(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				j, _ := m.Get(i)
				if *(*uint64)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapHashedKey(b *testing.B) {
	m := setupHashMapHashedKey(b)
	log := log2(uint64(benchmarkItemCount))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				hash := i << (64 - log)
				j, _ := m.GetHashedKey(hash)
				if *(*uint64)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapUintUnsafe(b *testing.B) {
	m := setupGoMap(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				j, _ := m[i]
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapUintMutex(b *testing.B) {
	m := setupGoMap(b)
	l := &sync.RWMutex{}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				l.RLock()
				j, _ := m[i]
				l.RUnlock()
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapStringUnsafe(b *testing.B) {
	m := setupGoMapString(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < benchmarkItemCount; i++ {
				s := strconv.Itoa(i)
				sVal, _ := m[s]
				if s != sVal {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapStringMutex(b *testing.B) {
	m := setupGoMapString(b)
	l := &sync.RWMutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < benchmarkItemCount; i++ {
				s := strconv.Itoa(i)
				l.RLock()
				sVal, _ := m[s]
				l.RUnlock()
				if s != sVal {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkWriteHashMap(b *testing.B) {
	m := New()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				j := uintptr(i)
				m.SetHashedKey(i, unsafe.Pointer(&j))
			}
		}
	})
}

func BenchmarkWriteHashMapHashedKey(b *testing.B) {
	m := New()
	log := log2(uint64(benchmarkItemCount))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				hash := i << (64 - log)
				j := uintptr(i)
				m.SetHashedKey(hash, unsafe.Pointer(&j))
			}
		}
	})
}

func BenchmarkWriteGoMap(b *testing.B) {
	m := make(map[uint64]uint64)
	l := &sync.RWMutex{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				l.Lock()
				m[i] = i
				l.Unlock()
			}
		}
	})
}

func BenchmarkUnsafePointer(b *testing.B) {
	b.StopTimer()
	var m [benchmarkItemCount]unsafe.Pointer
	for i := 0; i < benchmarkItemCount; i++ {
		item := &Animal{strconv.Itoa(i)}
		m[i] = unsafe.Pointer(item)
	}
	b.StartTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				item := m[i]
				animal := (*Animal)(item)
				if animal == nil {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkInterface(b *testing.B) {
	b.StopTimer()
	var m [benchmarkItemCount]interface{}
	for i := 0; i < benchmarkItemCount; i++ {
		item := &Animal{strconv.Itoa(i)}
		m[i] = item
	}
	b.StartTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				item := m[i]
				_, ok := item.(*Animal)
				if !ok {
					b.Fail()
				}
			}
		}
	})
}
