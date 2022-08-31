package hashmap

// defaultSize is the default size for a map.
const defaultSize = 8

// maxFillRate is the maximum fill rate for the slice before a resize will happen.
const maxFillRate = 50

// support all numeric types and types whose underlying type is also numeric.
type numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}
