package hashmap

import "testing"

func TestListNew32(t *testing.T) {
	l := NewList32()
	n := l.First()
	if n != nil {
		t.Error("First item of list should be nil.")
	}

	n = l.root.Next()
	if n != l.root {
		t.Error("Next element of empty list should be root.")
	}
}
