# hashmap  [![GoDoc](https://godoc.org/github.com/cornelk/hashmap?status.svg)](https://godoc.org/github.com/cornelk/hashmap) [![Go Report Card](https://goreportcard.com/badge/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap)

A Golang thread-safe HashMap optimized for lock-free read access.

Reading from the hash map in a thread-safe way is faster than reading from standard go map in an unsafe way:

```
go test -bench=.

BenchmarkReadHashMap-4            300000              5503 ns/op
BenchmarkReadGoMap-4              200000              6609 ns/op
BenchmarkReadGoMapSafe-4           30000             42261 ns/op
```

On a machine with more cores:

```
BenchmarkReadHashMap-32          1000000              1384 ns/op
BenchmarkReadGoMap-32            1000000              1540 ns/op
BenchmarkReadGoMapSafe-32          20000             76930 ns/op
```

