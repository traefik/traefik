package data

import (
	"fmt"
	"reflect"
	"testing"
)

var pf = fmt.Printf

func TestOrdMap(t *testing.T) {
	omap := OrderedMap(new(string), new(int))

	omap.Put("kaitlyn", 24)
	omap.Put("andrew", 25)
	omap.Put("lauren", 20)
	omap.Put("jen", 24)
	omap.Put("brennan", 25)

	omap.Delete("kaitlyn")

	assertDeep(t, omap.Keys(), []string{"andrew", "lauren", "jen", "brennan"})
	assertDeep(t, omap.Values(), []int{25, 20, 24, 25})
}

func ExampleOrderedMap() {
	omap := OrderedMap(new(string), new([]string))

	omap.Put("Bruce Springsteen",
		[]string{"Thunder Road", "Born to Run", "This Hard Land"})
	omap.Put("J. Geils Band",
		[]string{"Musta Got Lost", "Freeze Frame", "Southside Shuffle"})
	omap.Put("Bob Seger",
		[]string{"Against the Wind", "Roll Me Away", "Night Moves"})

	for _, key := range omap.Keys().([]string) {
		fmt.Println(key)
	}

	omap.Delete("J. Geils Band")
	fmt.Println("\nDeleted 'J. Geils Band'...\n")

	for _, key := range omap.Keys().([]string) {
		fmt.Printf("%s: %v\n", key, omap.Get(key))
	}

	// Output:
	// Bruce Springsteen
	// J. Geils Band
	// Bob Seger
	//
	// Deleted 'J. Geils Band'...
	//
	// Bruce Springsteen: [Thunder Road Born to Run This Hard Land]
	// Bob Seger: [Against the Wind Roll Me Away Night Moves]
}

func assertDeep(t *testing.T, v1, v2 interface{}) {
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("%v != %v", v1, v2)
	}
}
