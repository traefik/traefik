package redactor

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	Name         string
	EName        string `export:"true"`
	EFName       string `export:"false"`
	Value        int
	EValue       int `export:"true"`
	EFValue      int `export:"false"`
	List         []string
	EList        []string `export:"true"`
	EFList       []string `export:"false"`
	Courgette    Courgette
	ECourgette   Courgette `export:"true"`
	EFCourgette  Courgette `export:"false"`
	Pourgette    *Courgette
	EPourgette   *Courgette `export:"true"`
	EFPourgette  *Courgette `export:"false"`
	Aubergine    map[string]string
	EAubergine   map[string]string `export:"true"`
	EFAubergine  map[string]string `export:"false"`
	SAubergine   map[string]Tomate
	ESAubergine  map[string]Tomate `export:"true"`
	EFSAubergine map[string]Tomate `export:"false"`
	PSAubergine  map[string]*Tomate
	EPAubergine  map[string]*Tomate `export:"true"`
	EFPAubergine map[string]*Tomate `export:"false"`
}

func Test_doOnStruct(t *testing.T) {
	testCase := []struct {
		name             string
		base             *Carotte
		expected         *Carotte
		enabledByDefault bool
	}{
		{
			name: "primitive",
			base: &Carotte{
				Name:   "koko",
				EName:  "kiki",
				Value:  666,
				EValue: 666,
				List:   []string{"test"},
				EList:  []string{"test"},
			},
			expected: &Carotte{
				Name:   "xxxx",
				EName:  "kiki",
				EValue: 666,
				List:   []string{"xxxx"},
				EList:  []string{"test"},
			},
		},
		{
			name: "primitive2",
			base: &Carotte{
				Name:    "koko",
				EFName:  "keke",
				Value:   666,
				EFValue: 777,
				List:    []string{"test"},
				EFList:  []string{"test"},
			},
			expected: &Carotte{
				Name:   "koko",
				EFName: "xxxx",
				Value:  666,
				List:   []string{"test"},
				EFList: []string{"xxxx"},
			},
			enabledByDefault: true,
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
			name: "struct2",
			base: &Carotte{
				Name:   "koko",
				EFName: "keke",
				Courgette: Courgette{
					Ji: "huu",
				},
				EFCourgette: Courgette{
					Ji: "huu",
				},
			},
			expected: &Carotte{
				Name:   "koko",
				EFName: "xxxx",
				Courgette: Courgette{
					Ji: "huu",
					Ho: "",
				},
			},
			enabledByDefault: true,
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
			name: "pointer2",
			base: &Carotte{
				Name:   "koko",
				EFName: "keke",
				Pourgette: &Courgette{
					Ji: "hoo",
				},
				EFPourgette: &Courgette{
					Ji: "hoo",
				},
			},
			expected: &Carotte{
				Name:   "koko",
				EFName: "xxxx",
				Pourgette: &Courgette{
					Ji: "hoo",
				},
				EFPourgette: nil,
			},
			enabledByDefault: true,
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
			name: "export struct 2",
			base: &Carotte{
				Name:   "koko",
				EFName: "keke",
				ECourgette: Courgette{
					Ji: "huu",
				},
				EFCourgette: Courgette{
					Ji: "huu",
				},
			},
			expected: &Carotte{
				Name:   "koko",
				EFName: "xxxx",
				ECourgette: Courgette{
					Ji: "huu",
				},
			},
			enabledByDefault: true,
		},
		{
			name: "export pointer struct",
			base: &Carotte{
				Name: "koko",
				EPourgette: &Courgette{
					Ji: "huu",
				},
			},
			expected: &Carotte{
				Name: "xxxx",
				EPourgette: &Courgette{
					Ji: "xxxx",
				},
			},
		},
		{
			name: "export pointer struct 2",
			base: &Carotte{
				Name:   "koko",
				EFName: "keke",
				EPourgette: &Courgette{
					Ji: "huu",
				},
				EFPourgette: &Courgette{
					Ji: "huu",
				},
			},
			expected: &Carotte{
				Name:   "koko",
				EFName: "xxxx",
				EPourgette: &Courgette{
					Ji: "huu",
				},
				EFPourgette: nil,
			},
			enabledByDefault: true,
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
			name: "export map string/string 2",
			base: &Carotte{
				Name:   "koko",
				EFName: "keke",
				EAubergine: map[string]string{
					"foo": "bar",
				},
				EFAubergine: map[string]string{
					"foo": "bar",
				},
			},
			expected: &Carotte{
				Name:   "koko",
				EFName: "xxxx",
				EAubergine: map[string]string{
					"foo": "bar",
				},
				EFAubergine: map[string]string{},
			},
			enabledByDefault: true,
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
			name: "export map string/pointer 2",
			base: &Carotte{
				Name: "koko",
				EPAubergine: map[string]*Tomate{
					"foo": {
						Ji: "fdskljf",
					},
				},
				EFPAubergine: map[string]*Tomate{
					"foo": {
						Ji: "fdskljf",
					},
				},
			},
			expected: &Carotte{
				Name: "koko",
				EPAubergine: map[string]*Tomate{
					"foo": {
						Ji: "fdskljf",
					},
				},
				EFPAubergine: map[string]*Tomate{},
			},
			enabledByDefault: true,
		},
		{
			name: "export map string/struct",
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
						Ji: "xxxx",
					},
				},
			},
		},
		{
			name: "export map string/struct 2",
			base: &Carotte{
				Name: "koko",
				ESAubergine: map[string]Tomate{
					"foo": {
						Ji: "JiJiJi",
					},
				},
				EFSAubergine: map[string]Tomate{
					"foo": {
						Ji: "JiJiJi",
					},
				},
			},
			expected: &Carotte{
				Name: "koko",
				ESAubergine: map[string]Tomate{
					"foo": {
						Ji: "JiJiJi",
					},
				},
				EFSAubergine: map[string]Tomate{},
			},
			enabledByDefault: true,
		},
	}

	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			val := reflect.ValueOf(test.base).Elem()
			err := doOnStruct(val, exportTag, test.enabledByDefault)
			require.NoError(t, err)

			assert.EqualValues(t, test.expected, test.base)
		})
	}
}
