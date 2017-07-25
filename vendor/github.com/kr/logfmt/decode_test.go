package logfmt

import (
	"reflect"
	"testing"
	"time"
)

type pair struct {
	k, v string
}

type coll struct {
	a []pair
}

func (c *coll) HandleLogfmt(key, val []byte) error {
	c.a = append(c.a, pair{string(key), string(val)})
	return nil
}

func TestDecodeCustom(t *testing.T) {
	data := []byte(`a=foo b=10ms c=cat E="123" d foo= emp=`)

	g := new(coll)
	if err := Unmarshal(data, g); err != nil {
		t.Fatal(err)
	}

	w := []pair{
		{"a", "foo"},
		{"b", "10ms"},
		{"c", "cat"},
		{"E", "123"},
		{"d", ""},
		{"foo", ""},
		{"emp", ""},
	}

	if !reflect.DeepEqual(w, g.a) {
		t.Errorf("\nwant %v\n got %v", w, g)
	}
}

func TestDecodeDefault(t *testing.T) {
	var g struct {
		Float  float64
		NFloat *float64
		String string
		Int    int
		D      time.Duration
		NB     *[]byte
		Here   bool
		This   int `logfmt:"that"`
	}

	em, err := NewStructHandler(&g)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name string
		val  string
		want interface{}
	}{
		{"float", "3.14", 3.14},
		{"nfloat", "123", float64(123)},
		{"string", "foobar", "foobar"},
		{"inT", "10", 10},
		{"d", "1h", 1 * time.Hour},
		{"nb", "bytes!", []byte("bytes!")},
		{"here", "", true},
		{"that", "5", 5},
	}

	rv := reflect.Indirect(reflect.ValueOf(&g))
	for i, test := range tests {
		err = em.HandleLogfmt([]byte(test.name), []byte(test.val))
		if err != nil {
			t.Error(err)
			continue
		}

		fv := reflect.Indirect(rv.Field(i))
		if !fv.IsValid() {
			ft := rv.Type().Field(i)
			t.Errorf("%s is invalid", ft.Name)
			continue
		}

		gv := fv.Interface()
		if !reflect.DeepEqual(gv, test.want) {
			t.Errorf("want %T %#v, got %T %#v", test.want, test.want, gv, gv)
		}
	}

	if g.Float != 3.14 {
		t.Errorf("want %v, got %v", 3.14, g.Float)
	}

	err = em.HandleLogfmt([]byte("nfloat"), []byte("123"))
	if err != nil {
		t.Fatal(err)
	}

	if g.NFloat == nil || *g.NFloat != 123 {
		t.Errorf("want %v, got %v", 123, *g.NFloat)
	}
}
