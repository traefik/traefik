package data

import (
	"reflect"

	"github.com/BurntSushi/ty"
)

// OrdMap has a parametric type `OrdMap<K, V>` where `K` is the type
// of the map's keys and `V` is the type of the map's values.
type OrdMap struct {
	m            reflect.Value
	keys         reflect.Value
	ktype, vtype reflect.Type
}

// OrderedMap returns a new instance of OrdMap instantiated with the key
// and value types given. Namely, the types should be provided via nil
// pointers, e.g., to create a map from strings to integers:
//
//	omap := OrderedMap(new(string), new(int))
//
// An ordered map maintains the insertion order of all keys in the map.
// Namely, `(*OrdMap).Keys()` returns a slice of keys in the order
// they were inserted. The order of a key can *only* be changed if it is
// deleted and added again.
//
// All of the operations on an ordered map have the same time complexity as
// the built-in `map`, except for `Delete` which is O(n) in the number of
// keys.
func OrderedMap(ktype, vtype interface{}) *OrdMap {
	// A giant hack to get `Check` to do all the type construction work for us.
	chk := ty.Check(
		new(func(*ty.A, *ty.B) (ty.A, ty.B, map[ty.A]ty.B, []ty.A)),
		ktype, vtype)
	tkey, tval := chk.Returns[0], chk.Returns[1]
	tmap, tkeys := chk.Returns[2], chk.Returns[3]

	return &OrdMap{
		m:     reflect.MakeMap(tmap),
		keys:  reflect.MakeSlice(tkeys, 0, 10),
		ktype: tkey,
		vtype: tval,
	}
}

// Exists has a parametric type:
//
//	func (om *OrdMap<K, V>) Exists(key K) bool
//
// Exists returns true if `key` is in the map `om`.
func (om *OrdMap) Exists(key interface{}) bool {
	rkey := ty.AssertType(key, om.ktype)
	return om.exists(rkey)
}

func (om *OrdMap) exists(rkey reflect.Value) bool {
	return om.m.MapIndex(rkey).IsValid()
}

// Put has a parametric type:
//
//	func (om *OrdMap<K, V>) Put(key K, val V)
//
// Put adds or overwrites `key` into the map `om` with value `val`.
// If `key` already exists in the map, then its position in the ordering
// of the map is not changed.
func (om *OrdMap) Put(key, val interface{}) {
	rkey := ty.AssertType(key, om.ktype)
	rval := ty.AssertType(val, om.vtype)
	if !om.exists(rkey) {
		om.keys = reflect.Append(om.keys, rkey)
	}
	om.m.SetMapIndex(rkey, rval)
}

// Get has a parametric type:
//
//	func (om *OrdMap<K, V>) Get(key K) V
//
// Get retrieves the value in the map `om` corresponding to `key`. If the
// value does not exist, then the zero value of type `V` is returned.
func (om *OrdMap) Get(key interface{}) interface{} {
	rkey := ty.AssertType(key, om.ktype)
	rval := om.m.MapIndex(rkey)
	if !rval.IsValid() {
		return om.zeroValue().Interface()
	}
	return rval.Interface()
}

// TryGet has a parametric type:
//
//	func (om *OrdMap<K, V>) TryGet(key K) (V, bool)
//
// TryGet retrieves the value in the map `om` corresponding to `key` and
// reports whether the value exists in the map or not. If the value does
// not exist, then the zero value of `V` and `false` are returned.
func (om *OrdMap) TryGet(key interface{}) (interface{}, bool) {
	rkey := ty.AssertType(key, om.ktype)
	rval := om.m.MapIndex(rkey)
	if !rval.IsValid() {
		return om.zeroValue().Interface(), false
	}
	return rval.Interface(), true
}

// Delete has a parametric type:
//
//	func (om *OrdMap<K, V>) Delete(key K)
//
// Delete removes `key` from the map `om`.
//
// N.B. Delete is O(n) in the number of keys.
func (om *OrdMap) Delete(key interface{}) {
	rkey := ty.AssertType(key, om.ktype)

	// Avoid doing work if we don't need to.
	if !om.exists(rkey) {
		return
	}

	keysLen := om.keys.Len()
	for i := 0; i < keysLen; i++ {
		if key == om.keys.Index(i).Interface() {
			// om.keys = append(om.keys[:i], om.keys[i+1:]...)
			om.keys = reflect.AppendSlice(
				om.keys.Slice(0, i), om.keys.Slice(i+1, keysLen))
			break
		}
	}

	// Setting a key to a zero reflect.Value deletes the key from the map.
	om.m.SetMapIndex(rkey, reflect.Value{})
}

// Keys has a parametric type:
//
//	func (om *OrdMap<K, V>) Keys() []K
//
// Keys returns a list of keys in `om` in the order they were inserted.
//
// Behavior is undefined if the list is modified by the caller.
func (om *OrdMap) Keys() interface{} {
	return om.keys.Interface()
}

// Values has a parametric type:
//
//  func (om *OrdMap<K, V>) Values() []V
//
// Values returns a shallow copy of the values in `om` in the order that they
// were inserted.
func (om *OrdMap) Values() interface{} {
	mlen := om.Len()
	tvals := reflect.SliceOf(om.vtype)
	rvals := reflect.MakeSlice(tvals, mlen, mlen)
	for i := 0; i < mlen; i++ {
		rvals.Index(i).Set(om.m.MapIndex(om.keys.Index(i)))
	}
	return rvals.Interface()
}

// Len has a parametric type:
//
//	func (om *OrdMap<K, V>) Len() int
//
// Len returns the number of keys in the map `om`.
func (om *OrdMap) Len() int {
	return om.m.Len()
}

func (om *OrdMap) zeroValue() reflect.Value {
	return reflect.New(om.vtype).Elem()
}
