package benchmarks

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alphadose/haxmap"
	"github.com/cornelk/hashmap"
	"github.com/puzpuzpuz/xsync/v2"
	"github.com/zhangyunhao116/skipmap"
)

const benchmarkItemCount = 1024

func setupHashMap(b *testing.B) *hashmap.Map[uintptr, uintptr] {
	b.Helper()

	m := hashmap.New[uintptr, uintptr]()
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Set(i, i)
	}
	return m
}

func setupHaxMap(b *testing.B) *haxmap.Map[uintptr, uintptr] {
	b.Helper()

	m := haxmap.New[uintptr, uintptr]()
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Set(i, i)
	}
	return m
}

func setupSkipMap(b *testing.B) *skipmap.Uint64Map[uint64] {
	b.Helper()

	m := skipmap.NewUint64[uint64]()
	for i := uint64(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}
	return m
}

func setupXsync(b *testing.B) *xsync.MapOf[uint64, uint64] {
	b.Helper()

	m := xsync.NewIntegerMapOf[uint64, uint64]()
	for i := uint64(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}
	return m
}

func setupHashMapString(b *testing.B) (*hashmap.Map[string, string], []string) {
	b.Helper()

	m := hashmap.New[string, string]()
	keys := make([]string, benchmarkItemCount)
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m.Set(s, s)
		keys[i] = s
	}

	return m, keys
}

func setupGoMap(b *testing.B) map[uintptr]uintptr {
	b.Helper()

	m := make(map[uintptr]uintptr)
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m[i] = i
	}

	return m
}

func setupGoSyncMap(b *testing.B) *sync.Map {
	b.Helper()

	m := &sync.Map{}
	for i := uintptr(0); i < benchmarkItemCount; i++ {
		m.Store(i, i)
	}

	return m
}

func setupGoMapString(b *testing.B) (map[string]string, []string) {
	b.Helper()

	m := make(map[string]string)
	keys := make([]string, benchmarkItemCount)
	for i := 0; i < benchmarkItemCount; i++ {
		s := strconv.Itoa(i)
		m[s] = s
		keys[i] = s
	}
	return m, keys
}

func BenchmarkReadHashMapUint(b *testing.B) {
	m := setupHashMap(b)
	b.ResetTimer()

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

func BenchmarkReadHashMapWithWritesUint(b *testing.B) {
	m := setupHashMap(b)
	var writer uintptr
	b.ResetTimer()

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
					j, _ := m.Get(i)
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
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < benchmarkItemCount; i++ {
				s := keys[i]
				sVal, _ := m.Get(s)
				if sVal != s {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadHaxMapUint(b *testing.B) {
	m := setupHaxMap(b)
	b.ResetTimer()

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

func BenchmarkReadHaxMapWithWritesUint(b *testing.B) {
	m := setupHaxMap(b)
	var writer uintptr
	b.ResetTimer()

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
					j, _ := m.Get(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func BenchmarkReadXsyncMapUint(b *testing.B) {
	m := setupXsync(b)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadXsyncMapWithWritesUint(b *testing.B) {
	m := setupXsync(b)
	var writer uintptr
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := uint64(0); i < benchmarkItemCount; i++ {
					m.Store(i, i)
				}
			}
		} else {
			for pb.Next() {
				for i := uint64(0); i < benchmarkItemCount; i++ {
					j, _ := m.Load(i)
					if j != i {
						b.Fail()
					}
				}
			}
		}
	})
}

func BenchmarkReadSkipMapUint(b *testing.B) {
	m := setupSkipMap(b)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := uint64(0); i < benchmarkItemCount; i++ {
				j, _ := m.Load(i)
				if j != i {
					b.Fail()
				}
			}
		}
	})
}

func BenchmarkReadGoMapUintUnsafe(b *testing.B) {
	m := setupGoMap(b)
	b.ResetTimer()
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
	b.ResetTimer()
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
	b.ResetTimer()

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
	b.ResetTimer()
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
	b.ResetTimer()

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
	b.ResetTimer()
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
	b.ResetTimer()
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
	m := hashmap.New[uintptr, uintptr]()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Set(i, i)
		}
	}
}

func BenchmarkWriteGoMapMutexUint(b *testing.B) {
	m := make(map[uintptr]uintptr)
	l := &sync.RWMutex{}
	b.ResetTimer()

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
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := uintptr(0); i < benchmarkItemCount; i++ {
			m.Store(i, i)
		}
	}
}
