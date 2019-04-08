package sprig

import (
	"fmt"
	"reflect"
	"sort"
)

// Reflection is used in these functions so that slices and arrays of strings,
// ints, and other types not implementing []interface{} can be worked with.
// For example, this is useful if you need to work on the output of regexs.

func list(v ...interface{}) []interface{} {
	return v
}

func push(list interface{}, v interface{}) []interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append(nl, v)

	default:
		panic(fmt.Sprintf("Cannot push on type %s", tp))
	}
}

func prepend(list interface{}, v interface{}) []interface{} {
	//return append([]interface{}{v}, list...)

	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append([]interface{}{v}, nl...)

	default:
		panic(fmt.Sprintf("Cannot prepend on type %s", tp))
	}
}

func last(list interface{}) interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil
		}

		return l2.Index(l - 1).Interface()
	default:
		panic(fmt.Sprintf("Cannot find last on type %s", tp))
	}
}

func first(list interface{}) interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil
		}

		return l2.Index(0).Interface()
	default:
		panic(fmt.Sprintf("Cannot find first on type %s", tp))
	}
}

func rest(list interface{}) []interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil
		}

		nl := make([]interface{}, l-1)
		for i := 1; i < l; i++ {
			nl[i-1] = l2.Index(i).Interface()
		}

		return nl
	default:
		panic(fmt.Sprintf("Cannot find rest on type %s", tp))
	}
}

func initial(list interface{}) []interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil
		}

		nl := make([]interface{}, l-1)
		for i := 0; i < l-1; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return nl
	default:
		panic(fmt.Sprintf("Cannot find initial on type %s", tp))
	}
}

func sortAlpha(list interface{}) []string {
	k := reflect.Indirect(reflect.ValueOf(list)).Kind()
	switch k {
	case reflect.Slice, reflect.Array:
		a := strslice(list)
		s := sort.StringSlice(a)
		s.Sort()
		return s
	}
	return []string{strval(list)}
}

func reverse(v interface{}) []interface{} {
	tp := reflect.TypeOf(v).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(v)

		l := l2.Len()
		// We do not sort in place because the incoming array should not be altered.
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[l-i-1] = l2.Index(i).Interface()
		}

		return nl
	default:
		panic(fmt.Sprintf("Cannot find reverse on type %s", tp))
	}
}

func compact(list interface{}) []interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := []interface{}{}
		var item interface{}
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if !empty(item) {
				nl = append(nl, item)
			}
		}

		return nl
	default:
		panic(fmt.Sprintf("Cannot compact on type %s", tp))
	}
}

func uniq(list interface{}) []interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		dest := []interface{}{}
		var item interface{}
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if !inList(dest, item) {
				dest = append(dest, item)
			}
		}

		return dest
	default:
		panic(fmt.Sprintf("Cannot find uniq on type %s", tp))
	}
}

func inList(haystack []interface{}, needle interface{}) bool {
	for _, h := range haystack {
		if reflect.DeepEqual(needle, h) {
			return true
		}
	}
	return false
}

func without(list interface{}, omit ...interface{}) []interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		res := []interface{}{}
		var item interface{}
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if !inList(omit, item) {
				res = append(res, item)
			}
		}

		return res
	default:
		panic(fmt.Sprintf("Cannot find without on type %s", tp))
	}
}

func has(needle interface{}, haystack interface{}) bool {
	if haystack == nil {
		return false
	}
	tp := reflect.TypeOf(haystack).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(haystack)
		var item interface{}
		l := l2.Len()
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if reflect.DeepEqual(needle, item) {
				return true
			}
		}

		return false
	default:
		panic(fmt.Sprintf("Cannot find has on type %s", tp))
	}
}

// $list := [1, 2, 3, 4, 5]
// slice $list     -> list[0:5] = list[:]
// slice $list 0 3 -> list[0:3] = list[:3]
// slice $list 3 5 -> list[3:5]
// slice $list 3   -> list[3:5] = list[3:]
func slice(list interface{}, indices ...interface{}) interface{} {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil
		}

		var start, end int
		if len(indices) > 0 {
			start = toInt(indices[0])
		}
		if len(indices) < 2 {
			end = l
		} else {
			end = toInt(indices[1])
		}

		return l2.Slice(start, end).Interface()
	default:
		panic(fmt.Sprintf("list should be type of slice or array but %s", tp))
	}
}
