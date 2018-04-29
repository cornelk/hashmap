package hashmap

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
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

func setupHashMapString(b *testing.B) (*HashMap, []string) {
	m := &HashMap{}
	keys := make([]string, benchmarkItemCount)
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m.Set(s, s)
		keys[i] = s
	}

	b.ResetTimer()
	return m, keys
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

func setupGoSyncMap(b *testing.B) *sync.Map {
	m := &sync.Map{}
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}

	b.ResetTimer()
	return m
}

func setupGoMapString(b *testing.B) (map[string]string, []string) {
	m := make(map[string]string)
	keys := make([]string, benchmarkItemCount)
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m[s] = s
		keys[i] = s
	}
	b.ResetTimer()
	return m, keys
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
	var writer uintptr

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					m.Set(i, i)
				}
			}
		} else {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					j, _ := m.GetUintKey(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func BenchmarkReadHashMapString(b *testing.B) {
	m, keys := setupHashMapString(b)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < benchmarkItemCount; i++ {
				s := keys[i]
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
	var writer uintptr

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					l.Lock()
					m[i] = i
					l.Unlock()
				}
			}
		} else {
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
	var writer uintptr

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					m.Store(i, i)
				}
			}
		} else {
			for pb.Next() {
				for i := uintptr(0); i < benchmarkItemCount; i++ {
					j, _ := m.Load(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func BenchmarkReadGoMapStringUnsafe(b *testing.B) {
	m, keys := setupGoMapString(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < benchmarkItemCount; i++ {
				s := keys[i]
				sVal := m[s]
				if s != sVal {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapStringMutex(b *testing.B) {
	m, keys := setupGoMapString(b)
	l := &sync.RWMutex{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < benchmarkItemCount; i++ {
				s := keys[i]
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
	m := &sync.Map{}

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Store(i, i)
		}
	}
}
