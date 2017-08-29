package hashmap

import (
	"testing"
)

func TestLog2(t *testing.T) {
	var fixtures = map[uintptr]uintptr{
		0: 0,
		1: 0,
		2: 1,
		3: 2,
		4: 2,
		5: 3,
	}

	for input, result := range fixtures {
		output := log2(input)
		if output != result {
			t.Errorf("Log2 of %d should have been %d but was %d", input, result, output)
		}
	}
}
