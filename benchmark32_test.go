package hashmap

import (
	"strconv"
	"sync"
	"testing"
	"unsafe"

	"golang.org/x/sync/syncmap"
)

//const benchmarkItemCount = 1 << 10 // 1024

func setupHashMap32(b *testing.B) *HashMap32 {
	m := &HashMap32{}
	for i := uint32(0); i < benchmarkItemCount; i++ {
		j := uintptr(i)
		m.Set(i, unsafe.Pointer(&j))
	}

	b.ResetTimer()
	return m
}

func setupHashMapString32(b *testing.B) *HashMap32 {
	m := &HashMap32{}
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m.Set(s, unsafe.Pointer(&s))
	}

	b.ResetTimer()
	return m
}

func writeHashMap32(m *HashMap32, i uint32) {
	j := uintptr(i)
	m.Set(i, unsafe.Pointer(&j))
}

func setupHashMapHashedKey32(b *testing.B) *HashMap32 {
	m := &HashMap32{}
	log := log2_32(uint32(benchmarkItemCount))
	for i := uint32(0); i < benchmarkItemCount; i++ {
		hash := i << (32 - log)
		j := uintptr(i)
		m.SetHashedKey(hash, unsafe.Pointer(&j))
	}

	b.ResetTimer()
	return m
}

func setupGoMap32(b *testing.B) map[uint32]uint32 {
	m := make(map[uint32]uint32)
	for i := uint32(0); i < benchmarkItemCount; i++ {
		m[i] = i
	}

	b.ResetTimer()
	return m
}

func writeGoMap32(m map[uint32]uint32, i uint32, mtx *sync.RWMutex) {
	mtx.Lock()
	m[i] = i
	mtx.Unlock()
}

func setupGoSyncMap32(b *testing.B) *syncmap.Map {
	m := &syncmap.Map{}
	for i := uint32(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}

	b.ResetTimer()
	return m
}

func writeGoSyncMap32(m *syncmap.Map, i uint32) {
	m.Store(i, i)
}

func setupGoMapString32(b *testing.B) map[string]string {
	m := make(map[string]string)
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m[s] = s
	}
	b.ResetTimer()
	return m
}

func BenchmarkReadHashMapUint32(b *testing.B) {
	m := setupHashMap32(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				j, _ := m.GetUintKey(i)
				if *(*uint32)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapWithWritesUint32(b *testing.B) {
	m := setupHashMap32(b)

	go func(bn int) {
		for n := 0; n < bn; n++ {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				writeHashMap32(m, i)
			}
		}
	}(b.N)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				j, _ := m.GetUintKey(i)
				if *(*uint32)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapString32(b *testing.B) {
	m := setupHashMapString32(b)

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

func BenchmarkReadHashMapInterface32(b *testing.B) {
	m := setupHashMap32(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				j, _ := m.Get(i)
				if *(*uint32)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapHashedKey32(b *testing.B) {
	m := setupHashMapHashedKey32(b)
	log := log2_32(uint32(benchmarkItemCount))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				hash := i << (32 - log)
				j, _ := m.GetHashedKey(hash)
				if *(*uint32)(j) != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapUintUnsafe32(b *testing.B) {
	m := setupGoMap32(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				j, _ := m[i]
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapUintMutex32(b *testing.B) {
	m := setupGoMap32(b)
	l := &sync.RWMutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
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

func BenchmarkReadGoMapWithWritesUintMutex32(b *testing.B) {
	m := setupGoMap32(b)
	l := &sync.RWMutex{}

	go func(bn int) {
		for n := 0; n < bn; n++ {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				writeGoMap32(m, i, l)
			}
		}
	}(b.N)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
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

func BenchmarkReadGoSyncMapUint32(b *testing.B) {
	m := setupGoSyncMap32(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoSyncMapWithWritesUint32(b *testing.B) {
	m := setupGoSyncMap32(b)

	go func(bn int) {
		for n := 0; n < bn; n++ {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				writeGoSyncMap32(m, i)
			}
		}
	}(b.N)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint32(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapStringUnsafe32(b *testing.B) {
	m := setupGoMapString32(b)
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

func BenchmarkReadGoMapStringMutex32(b *testing.B) {
	m := setupGoMapString32(b)
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

func BenchmarkWriteHashMapUint32(b *testing.B) {
	m := &HashMap32{}

	for n := 0; n < b.N; n++ {
		for i := uint32(0); i < benchmarkItemCount; i++ {
			j := uintptr(i)
			m.Set(i, unsafe.Pointer(&j))
		}
	}
}

func BenchmarkWriteHashMapHashedKey32(b *testing.B) {
	m := &HashMap32{}
	log := log2_32(uint32(benchmarkItemCount))

	for n := 0; n < b.N; n++ {
		for i := uint32(0); i < benchmarkItemCount; i++ {
			hash := i << (64 - log)
			j := uintptr(i)
			m.SetHashedKey(hash, unsafe.Pointer(&j))
		}
	}
}

func BenchmarkWriteGoMapMutexUint32(b *testing.B) {
	m := make(map[uint32]uint32)
	l := &sync.RWMutex{}

	for n := 0; n < b.N; n++ {
		for i := uint32(0); i < benchmarkItemCount; i++ {
			l.Lock()
			m[i] = i
			l.Unlock()
		}
	}
}

func BenchmarkWriteGoSyncMapUint32(b *testing.B) {
	m := &syncmap.Map{}

	for n := 0; n < b.N; n++ {
		for i := uint32(0); i < benchmarkItemCount; i++ {
			m.Store(i, i)
		}
	}
}
