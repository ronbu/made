package mapset

import (
	"testing"
)

func Test(t *testing.T) {
	a, b := NewSet([]interface{}{1, 2, 3}), NewSet([]interface{}{2, 3, 4})

	if !NewSet([]interface{}{1, 2}).Equals(NewSet([]interface{}{1, 2})) {
		t.Fatal("Equals does not work")
	}
	var e MapSet
	e = NewSet([]interface{}{1, 2, 3, 4})
	if a := a.Union(b); !a.Equals(e) {
		t.Fatal("Expected: ")
	}

	e = NewSet([]interface{}{2, 3})
	if a := a.Inter(b); !a.Equals(e) {
		t.Fatal("Expected: ")
	}
	e = NewSet([]interface{}{1})
	if a := a.Diff(b); !a.Equals(e) {
		t.Fatal("Expected: ")
	}
	e = NewSet([]interface{}{1, 4})
	if a := a.SymDiff(b); !a.Equals(e) {
		t.Fatal("Expected: ")
	}
	if a.Super(b) {
		t.Fatal("Super fails")
	}
	if a.Sub(b) {
		t.Fatal("Sub fails")
	}
	if !NewSet([]interface{}{1, 2}).Sub(
		NewSet([]interface{}{1, 2, 3})) {

		t.Fatal("Sub fails")
	}
	if !NewSet([]interface{}{1, 2, 3}).Super(
		NewSet([]interface{}{1, 2})) {

		t.Fatal("Super fails")
	}
}
