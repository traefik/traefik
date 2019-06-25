package env

type Ya struct {
	Foo     *Yaa
	Field1  string
	Field2  bool
	Field3  int
	Field4  map[string]string
	Field5  map[string]int
	Field6  map[string]struct{ Field string }
	Field7  map[string]struct{ Field map[string]string }
	Field8  map[string]*struct{ Field string }
	Field9  map[string]*struct{ Field map[string]string }
	Field10 struct{ Field string }
	Field11 *struct{ Field string }
	Field12 *string
	Field13 *bool
	Field14 *int
	Field15 []int
}

type Yaa struct {
	FieldIn1  string
	FieldIn2  bool
	FieldIn3  int
	FieldIn4  map[string]string
	FieldIn5  map[string]int
	FieldIn6  map[string]struct{ Field string }
	FieldIn7  map[string]struct{ Field map[string]string }
	FieldIn8  map[string]*struct{ Field string }
	FieldIn9  map[string]*struct{ Field map[string]string }
	FieldIn10 struct{ Field string }
	FieldIn11 *struct{ Field string }
	FieldIn12 *string
	FieldIn13 *bool
	FieldIn14 *int
}

type Yo struct {
	Foo string `description:"Foo description"`
	Fii string `description:"Fii description"`
	Fuu string `description:"Fuu description"`
	Yi  *Yi    `label:"allowEmpty"`
	Yu  *Yi
}

func (y *Yo) SetDefaults() {
	y.Foo = "foo"
	y.Fii = "fii"
}

type Yi struct {
	Foo string
	Fii string
	Fuu string
}

func (y *Yi) SetDefaults() {
	y.Foo = "foo"
	y.Fii = "fii"
}

type Yu struct {
	Yi
}

type Ye struct {
	*Yi
}
