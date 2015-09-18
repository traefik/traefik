package fun

import (
	"testing"
)

func TestSet(t *testing.T) {
	a := []string{"andrew", "plato", "andrew", "cauchy", "cauchy", "andrew"}
	set := Set(a).(map[string]bool)

	assertDeep(t, set, map[string]bool{
		"andrew": true,
		"plato":  true,
		"cauchy": true,
	})
}

func TestUnion(t *testing.T) {
	a := map[string]bool{
		"springsteen": true,
		"jgeils":      true,
		"seger":       true,
		"metallica":   true,
	}
	b := map[string]bool{
		"metallica": true,
		"chesney":   true,
		"mcgraw":    true,
		"cash":      true,
	}
	c := Union(a, b).(map[string]bool)

	assertDeep(t, c, map[string]bool{
		"springsteen": true,
		"jgeils":      true,
		"seger":       true,
		"metallica":   true,
		"chesney":     true,
		"mcgraw":      true,
		"cash":        true,
	})
}

func TestIntersection(t *testing.T) {
	a := map[string]bool{
		"springsteen": true,
		"jgeils":      true,
		"seger":       true,
		"metallica":   true,
	}
	b := map[string]bool{
		"metallica": true,
		"chesney":   true,
		"mcgraw":    true,
		"cash":      true,
	}
	c := Intersection(a, b).(map[string]bool)

	assertDeep(t, c, map[string]bool{
		"metallica": true,
	})
}

func TestDifference(t *testing.T) {
	a := map[string]bool{
		"springsteen": true,
		"jgeils":      true,
		"seger":       true,
		"metallica":   true,
	}
	b := map[string]bool{
		"metallica": true,
		"chesney":   true,
		"mcgraw":    true,
		"cash":      true,
	}
	c := Difference(a, b).(map[string]bool)

	assertDeep(t, c, map[string]bool{
		"springsteen": true,
		"jgeils":      true,
		"seger":       true,
	})
}
