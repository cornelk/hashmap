package hashmap

import (
	"strconv"
	"sync"
	"testing"

	"golang.org/x/sync/syncmap"
)

const benchmarkItemCount = 1 << 10 // 1024

func setupHashMap(b *testing.B) *HashMap {
	m := &HashMap{}
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Set(i, i)
	}

	b.ResetTimer()
	return m
}

func setupHashMapString(b *testing.B) *HashMap {
	m := &HashMap{}
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m.Set(s, s)
	}

	b.ResetTimer()
	return m
}

func writeHashMap(m *HashMap, i uintptr) {
	m.Set(i, i)
}

func setupHashMapHashedKey(b *testing.B) *HashMap {
	m := &HashMap{}
	log := log2(uintptr(benchmarkItemCount))
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		hash := i << (strconv.IntSize - log)
		m.SetHashedKey(hash, i)
	}

	b.ResetTimer()
	return m
}

func setupGoMap(b *testing.B) map[uintptr]uintptr {
	m := make(map[uintptr]uintptr)
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m[i] = i
	}

	b.ResetTimer()
	return m
}

func writeGoMap(m map[uintptr]uintptr, i uintptr, mtx *sync.RWMutex) {
	mtx.Lock()
	m[i] = i
	mtx.Unlock()
}

func setupGoSyncMap(b *testing.B) *syncmap.Map {
	m := &syncmap.Map{}
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}

	b.ResetTimer()
	return m
}

func writeGoSyncMap(m *syncmap.Map, i uintptr) {
	m.Store(i, i)
}

func setupGoMapString(b *testing.B) map[string]string {
	m := make(map[string]string)
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m[s] = s
	}
	b.ResetTimer()
	return m
}

func BenchmarkReadHashMapUint(b *testing.B) {
	m := setupHashMap(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.GetUintKey(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapWithWritesUint(b *testing.B) {
	m := setupHashMap(b)

	go func(bn int) {
		for n := 0; n < bn; n++ {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				writeHashMap(m, i)
			}
		}
	}(b.N)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.GetUintKey(i)
				if j != i {
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
				if sVal != s {
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
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Get(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHashMapHashedKey(b *testing.B) {
	m := setupHashMapHashedKey(b)
	log := log2(uintptr(benchmarkItemCount))

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				hash := i << (strconv.IntSize - log)
				j, _ := m.GetHashedKey(hash)
				if j != i {
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
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j := m[i]
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
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				l.RLock()
				j := m[i]
				l.RUnlock()
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapWithWritesUintMutex(b *testing.B) {
	m := setupGoMap(b)
	l := &sync.RWMutex{}

	go func(bn int) {
		for n := 0; n < bn; n++ {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				writeGoMap(m, i, l)
			}
		}
	}(b.N)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				l.RLock()
				j := m[i]
				l.RUnlock()
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoSyncMapUint(b *testing.B) {
	m := setupGoSyncMap(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoSyncMapWithWritesUint(b *testing.B) {
	m := setupGoSyncMap(b)

	go func(bn int) {
		for n := 0; n < bn; n++ {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				writeGoSyncMap(m, i)
			}
		}
	}(b.N)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uintptr(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
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
				sVal := m[s]
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
				sVal := m[s]
				l.RUnlock()
				if s != sVal {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkWriteHashMapUint(b *testing.B) {
	m := &HashMap{}

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Set(i, i)
		}
	}
}

func BenchmarkWriteHashMapHashedKey(b *testing.B) {
	m := &HashMap{}
	log := log2(uintptr(benchmarkItemCount))

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			hash := i << (strconv.IntSize - log)
			m.SetHashedKey(hash, i)
		}
	}
}

func BenchmarkWriteGoMapMutexUint(b *testing.B) {
	m := make(map[uintptr]uintptr)
	l := &sync.RWMutex{}

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			l.Lock()
			m[i] = i
			l.Unlock()
		}
	}
}

func BenchmarkWriteGoSyncMapUint(b *testing.B) {
	m := &syncmap.Map{}

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Store(i, i)
		}
	}
}
