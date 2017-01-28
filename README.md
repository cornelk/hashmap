# hashmap [![Build Status](https://travis-ci.org/cornelk/hashmap.svg?branch=master)](https://travis-ci.org/cornelk/hashmap) [![GoDoc](https://godoc.org/github.com/cornelk/hashmap?status.svg)](https://godoc.org/github.com/cornelk/hashmap) [![Go Report Card](https://goreportcard.com/badge/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap) [![codecov](https://codecov.io/gh/cornelk/hashmap/branch/master/graph/badge.svg)](https://codecov.io/gh/cornelk/hashmap)

A Golang thread-safe HashMap optimized for lock-free read access on 64 bit systems.

Reading from the hash map in a thread-safe way is faster than reading from standard go map in an unsafe way:

```
BenchmarkReadHashMap-8       	 1000000	      1441 ns/op
BenchmarkReadGoMapUnsafe-8   	 1000000	      1976 ns/op
BenchmarkReadGoMap-8         	   50000	     27545 ns/op
```

The library uses a sorted linked list and a slice as an index into the list.

It optimizes the slice access by circumventing the Golang size check when reading from the slice. Once a slice is allocated, the size of it does not change.
The lib limits the index into the slice, therefor the Golang size check is obsolete. When the slice reaches a defined fill rate, a bigger slice is allocated
and all keys recalculated and transferred into the new slice.

The resize operation uses a lock to ensure that only one resize operation is happening. This way, no CPU and memory resources are wasted by multiple goroutines working on the resize.
