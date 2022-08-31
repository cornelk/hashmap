module github.com/cornelk/hashmap/benchmarks

go 1.19

replace github.com/cornelk/hashmap => ../

require (
	github.com/alphadose/haxmap v0.3.1-0.20220831034947-0d601bb44159
	github.com/cornelk/hashmap v1.0.6-0.20220829041708-517efe3afe16
)

require github.com/cespare/xxhash v1.1.0 // indirect
