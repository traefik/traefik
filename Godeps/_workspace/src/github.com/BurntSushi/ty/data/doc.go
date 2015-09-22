/*
Package data tumbles down the rabbit hole into parametric data types.

A parametric data type in this package is a data type that is parameterized by
one or more types discovered at run time. For example, an ordered map is
parameterized by two types: the type of its keys and the type of its values.

Implementation

While all parametric inputs and outputs of each function have a Go type of
`interface{}`, the underlying type is maintained via reflection. In particular,
any operation that interacts with a parametric data type does so via
reflection so that the type safety found in Go at compile time can be
recovered at run time.

For example, consider the case of an ordered map. One might define such a map
as a list of its keys in order and a map of `interface{}` to `interface{}`:

	type OrdMap struct {
		M map[interface{}]interface{}
		Keys []interface{}
	}

And one can interact with this map using standard built-in Go operations:

	// Add a key
	M["key"] = "value"
	Keys = append(Keys, "key")

	// Delete a key
	delete(M, "key")

	// Read a key
	M["key"]

But there is no type safety with such a representation, even at run time:

	// Both of these operations are legal with
	// the aforementioned representation.
	M["key"] = "value"
	M[5] = true

Thus, the contribution of this library is to maintain type safety at run time
by guaranteeing that all operations are consistent with Go typing rules:

	type OrdMap struct {
		M reflect.Value
		Keys reflect.Value
	}

And one must interact with a map using `reflect`:

	key, val := "key", "value"
	rkey := reflect.ValueOf(key)
	rval := reflect.ValueOf(val)

	// Add a key
	M.SetMapIndex(rkey, rval)
	Keys = reflect.Append(Keys, rkey)

	// Delete a key
	M.SetMapIndex(rkey, reflect.Value{})

	// Read a key
	M.MapIndex(rkey)

Which guarantees, at run-time, that the following cannot happen:

	key2, val2 := 5, true
	rkey2 := reflect.ValueOf(key2)
	rval2 := reflect.ValueOf(val2)

	// One or the other operation will be disallowed,
	// assuming `OrdMap` isn't instantiated with
	// `interface{}` as the key and value type.
	M.SetMapIndex(rkey, rval)
	M.SetMapIndex(rkey2, rval2)

The result is much more painful library code but only slightly more painful
client code.
*/
package data
