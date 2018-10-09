package pluralforms

type math interface {
	calc(n uint32) uint32
}

type mod struct {
	value uint32
}

func (m mod) calc(n uint32) uint32 {
	return n % m.value
}
