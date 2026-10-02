package hashmap

// defaultSize is the default size for a map.
const defaultSize = 8

// maxFillRate is the maximum fill rate for the slice before a resize will happen.
const maxFillRate = 50

// support all numeric, pointer and string types and aliases of those.
type hashable interface {
	comparable
}
