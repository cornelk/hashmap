package hashmap

import "testing"

func TestListNew(t *testing.T) {
	l := NewList()
	n := l.First()
	if n != nil {
		t.Error("First item of list should be nil.")
	}

	n = l.head.Next()
	if n != nil {
		t.Error("Next element of empty list should be nil.")
	}
}
