package anonymize

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Courgette struct {
	Ji string
	Ho string
}
type Tomate struct {
	Ji string
	Ho string
}

type Carotte struct {
	Name        string
	Value       int
	Courgette   Courgette
	ECourgette  Courgette `export:"true"`
	Pourgette   *Courgette
	EPourgette  *Courgette `export:"true"`
	Aubergine   map[string]string
	EAubergine  map[string]string `export:"true"`
	SAubergine  map[string]Tomate
	ESAubergine map[string]Tomate `export:"true"`
	PSAubergine map[string]*Tomate
	EPAubergine map[string]*Tomate `export:"true"`
}

func Test_doOnStruct(t *testing.T) {
	testCase := []struct {
		name     string
		base     *Carotte
		expected *Carotte
		hasError bool
	}{
		{
			name: "primitive",
			base: &Carotte{
				Name:  "koko",
				Value: 666,
			},
			expected: &Carotte{
				Name: "xxxx",
			},
		},
		{
			name: "struct",
			base: &Carotte{
				Name: "koko",
				Courgette: Courgette{
					Ji: "huu",
				},
			},
			expected: &Carotte{
				Name: "xxxx",
			},
		},
		{
			name: "pointer",
			base: &Carotte{
				Name: "koko",
				Pourgette: &Courgette{
					Ji: "hoo",
				},
			},
			expected: &Carotte{
				Name:      "xxxx",
				Pourgette: nil,
			},
		},
		{
			name: "export struct",
			base: &Carotte{
				Name: "koko",
				ECourgette: Courgette{
					Ji: "huu",
				},
			},
			expected: &Carotte{
				Name: "xxxx",
				ECourgette: Courgette{
					Ji: "xxxx",
				},
			},
		},
		{
			name: "export pointer struct",
			base: &Carotte{
				Name: "koko",
				ECourgette: Courgette{
					Ji: "huu",
				},
			},
			expected: &Carotte{
				Name: "xxxx",
				ECourgette: Courgette{
					Ji: "xxxx",
				},
			},
		},
		{
			name: "export map string/string",
			base: &Carotte{
				Name: "koko",
				EAubergine: map[string]string{
					"foo": "bar",
				},
			},
			expected: &Carotte{
				Name: "xxxx",
				EAubergine: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			name: "export map string/pointer",
			base: &Carotte{
				Name: "koko",
				EPAubergine: map[string]*Tomate{
					"foo": {
						Ji: "fdskljf",
					},
				},
			},
			expected: &Carotte{
				Name: "xxxx",
				EPAubergine: map[string]*Tomate{
					"foo": {
						Ji: "xxxx",
					},
				},
			},
		},
		{
			name: "export map string/struct (UNSAFE)",
			base: &Carotte{
				Name: "koko",
				ESAubergine: map[string]Tomate{
					"foo": {
						Ji: "JiJiJi",
					},
				},
			},
			expected: &Carotte{
				Name: "xxxx",
				ESAubergine: map[string]Tomate{
					"foo": {
						Ji: "JiJiJi",
					},
				},
			},
			hasError: true,
		},
	}

	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			val := reflect.ValueOf(test.base).Elem()
			err := doOnStruct(val)
			if !test.hasError && err != nil {
				t.Fatal(err)
			}
			if test.hasError && err == nil {
				t.Fatal("Got no error but want an error.")
			}

			assert.EqualValues(t, test.expected, test.base)
		})
	}
}
