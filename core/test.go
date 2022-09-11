package core

var F func(t *Tree, path string, content string)
var expect func(t *Tree, path string)
var deleted func(t *Tree, path string)

func TestTrim() {
	x := NewTree(&Info{})
	F(x, "a", "a")
	F(x, "a2", "a")
	F(x, "b", "b")

	y := NewTree(&Info{})
	F(y, "b", "b")

	// trim x with y
	// Make sure a duplicate file in x won't get deleted by trimming it without that file
	// here y{b} deletes x{b} but not x{a},x{a2} which are duplicates...
	expect(x, "a")
	expect(x, "a2")
	deleted(x, "b")
}
