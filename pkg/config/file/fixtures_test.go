package file

type bar string

type Yo struct {
	Foo string
	Fii string
	Fuu string
	Yi  *Yi `file:"allowEmpty"`
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
