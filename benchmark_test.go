package hashmap

import (
	"sync"
	"testing"
)

func setupHashMap(b *testing.B) *HashMap {
	b.StopTimer()
	m := New()
	for i := 0; i < 1024; i++ {
		m.Add(uint64(i), uint64(i))
	}
	b.StartTimer()
	return m
}

func setupGoMap(b *testing.B) map[uint64]uint64 {
	b.StopTimer()
	m := make(map[uint64]uint64)
	for i := 0; i < 1024; i++ {
		m[uint64(i)] = uint64(i)
	}
	b.StartTimer()
	return m
}

func BenchmarkReadHashMap(b *testing.B) {
	m := setupHashMap(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < 1024; i++ {
				j, _ := m.Get(uint64(i))
				if j.(uint64) != uint64(i) {
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
			for i := 0; i < 1024; i++ {
				j, _ := m[uint64(i)]
				if j != uint64(i) {
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
			for i := 0; i < 1024; i++ {
				l.RLock()
				j, _ := m[uint64(i)]
				l.RUnlock()
				if j != uint64(i) {
					b.Fail()
				}
			}
		}
	})
}
