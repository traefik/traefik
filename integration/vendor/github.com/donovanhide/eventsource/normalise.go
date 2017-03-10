package eventsource

import (
	"io"
)

// A reader which normalises line endings
// "/r" and "/r/n" are converted to "/n"
type normaliser struct {
	r        io.Reader
	lastChar byte
}

func newNormaliser(r io.Reader) *normaliser {
	return &normaliser{r: r}
}

func (norm *normaliser) Read(p []byte) (n int, err error) {
	n, err = norm.r.Read(p)
	for i := 0; i < n; i++ {
		switch {
		case p[i] == '\n' && norm.lastChar == '\r':
			copy(p[i:n], p[i+1:])
			norm.lastChar = p[i]
			n--
			i--
		case p[i] == '\r':
			norm.lastChar = p[i]
			p[i] = '\n'
		default:
			norm.lastChar = p[i]
		}
	}
	return
}
