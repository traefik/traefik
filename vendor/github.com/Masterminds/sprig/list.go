package sprig

import (
	"reflect"
	"sort"
)

func list(v ...interface{}) []interface{} {
	return v
}

func push(list []interface{}, v interface{}) []interface{} {
	return append(list, v)
}

func prepend(list []interface{}, v interface{}) []interface{} {
	return append([]interface{}{v}, list...)
}

func last(list []interface{}) interface{} {
	l := len(list)
	if l == 0 {
		return nil
	}
	return list[l-1]
}

func first(list []interface{}) interface{} {
	if len(list) == 0 {
		return nil
	}
	return list[0]
}

func rest(list []interface{}) []interface{} {
	if len(list) == 0 {
		return list
	}
	return list[1:]
}

func initial(list []interface{}) []interface{} {
	l := len(list)
	if l == 0 {
		return list
	}
	return list[:l-1]
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

func reverse(v []interface{}) []interface{} {
	// We do not sort in place because the incomming array should not be altered.
	l := len(v)
	c := make([]interface{}, l)
	for i := 0; i < l; i++ {
		c[l-i-1] = v[i]
	}
	return c
}

func compact(list []interface{}) []interface{} {
	res := []interface{}{}
	for _, item := range list {
		if !empty(item) {
			res = append(res, item)
		}
	}
	return res
}

func uniq(list []interface{}) []interface{} {
	dest := []interface{}{}
	for _, item := range list {
		if !inList(dest, item) {
			dest = append(dest, item)
		}
	}
	return dest
}

func inList(haystack []interface{}, needle interface{}) bool {
	for _, h := range haystack {
		if reflect.DeepEqual(needle, h) {
			return true
		}
	}
	return false
}

func without(list []interface{}, omit ...interface{}) []interface{} {
	res := []interface{}{}
	for _, i := range list {
		if !inList(omit, i) {
			res = append(res, i)
		}
	}
	return res
}
