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
	log := log2(uint64(benchmarkItemCount))
	for i := uint64(0); i < benchmarkItemCount; i++ {
		hash := i << (64 - log)
		j := uintptr(i)
		m.Set(hash, unsafe.Pointer(&j))
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

func BenchmarkReadHashMap(b *testing.B) {
	m := setupHashMap(b)
	log := log2(uint64(benchmarkItemCount))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				hash := i << (64 - log)
				j, _ := m.Get(hash)
				if *(*uint64)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapUnsafe(b *testing.B) {
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

func BenchmarkReadGoMap(b *testing.B) {
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
