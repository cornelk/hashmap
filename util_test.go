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
	type testFixture struct {
		input  interface{}
		output uintptr
	}
	var fixtures = []testFixture{
		{input: "123", output: 9076048966884696828},
		{input: []byte("123"), output: 9076048966884696828},
		{input: int(1), output: 1754102016959854353},
		{input: int8(1), output: 1754102016959854353},
		{input: int16(-1), output: 6588593453755867710},
		{input: int32(math.MaxInt32), output: 15638166383137924496},
		{input: int64(math.MaxInt64), output: 4889285460913276945},
		{input: uint(0), output: 9257401834698437112},
		{input: uint8(123), output: 2662021623061770816},
		{input: uint16(1234), output: 1663804089773015140},
		{input: uint32(12345), output: 11667327197262824396},
		{input: uint64(123456), output: 9063688366117729139},
		{input: uintptr(1234567), output: 14770111569646361914},
	}

	for _, f := range fixtures {
		output := getKeyHash(f.input)
		if output != f.output {
			t.Errorf("Key hash of %v and type %v should have been %d but was %d", f.input, reflect.TypeOf(f.input), f.output, output)
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
