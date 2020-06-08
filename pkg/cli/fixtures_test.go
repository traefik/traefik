package cli

type Yo struct {
	Foo string `description:"Foo description"`
	Fii string `description:"Fii description"`
	Fuu string `description:"Fuu description"`
	Yi  *Yi    `label:"allowEmpty" file:"allowEmpty"`
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
