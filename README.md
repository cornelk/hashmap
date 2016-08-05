# hashmap [![Build Status](https://travis-ci.org/cornelk/hashmap.svg?branch=master)](https://travis-ci.org/cornelk/hashmap) [![GoDoc](https://godoc.org/github.com/cornelk/hashmap?status.svg)](https://godoc.org/github.com/cornelk/hashmap) [![Go Report Card](https://goreportcard.com/badge/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap) [![codecov](https://codecov.io/gh/cornelk/hashmap/branch/master/graph/badge.svg)](https://codecov.io/gh/cornelk/hashmap)

A Golang thread-safe HashMap optimized for lock-free read access.

Reading from the hash map in a thread-safe way is faster than reading from standard go map in an unsafe way:

```
BenchmarkReadHashMap-4       	  200000	      7682 ns/op
BenchmarkReadGoMapUnsafe-4   	  200000	      8675 ns/op
BenchmarkReadGoMap-4         	   20000	     57842 ns/op
```

On a machine with more cores:

```
BenchmarkReadHashMap-32          1000000              1309 ns/op
BenchmarkReadGoMapUnsafe-32      1000000              1569 ns/op
BenchmarkReadGoMap-32              20000             77803 ns/op
```
