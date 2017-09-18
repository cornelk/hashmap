package hashmap

import (
	"math"
	"reflect"
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

func TestKeyHash(t *testing.T) {
	var fixtures = map[interface{}]uintptr{
		"123":                9076048966884696828,
		int(1):               1754102016959854353,
		int8(1):              1754102016959854353,
		int16(-1):            6588593453755867710,
		int32(math.MaxInt32): 15638166383137924496,
		int64(math.MaxInt64): 4889285460913276945,
		uint(0):              9257401834698437112,
		uint8(123):           2662021623061770816,
		uint16(1234):         1663804089773015140,
		uint32(12345):        11667327197262824396,
		uint64(123456):       9063688366117729139,
		uintptr(1234567):     14770111569646361914,
	}

	for input, result := range fixtures {
		output := getKeyHash(input)
		if output != result {
			t.Errorf("Key hash of %v and type %v should have been %d but was %d", input, reflect.TypeOf(input), result, output)
		}
	}
}

func TestKeyHashPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Invalid type did not panic")
		}
	}()

	_ = getKeyHash(true)
}
