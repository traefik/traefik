package route

import (
	"fmt"
	"strings"

	. "gopkg.in/check.v1"
)

type IterSuite struct {
}

var _ = Suite(&IterSuite{})

func (s *IterSuite) TestEmptyOperationsSucceed(c *C) {
	var vals []string
	var seps []byte
	i := newIter(vals, seps)
	_, _, ok := i.next()
	c.Assert(ok, Equals, false)

	_, _, ok = i.next()
	c.Assert(ok, Equals, false)
}

func (s *IterSuite) TestUnwind(c *C) {
	tc := []charTc{
		charTc{
			name:  "Simple iteration",
			input: []string{"hello"},
			sep:   []byte{pathSep},
		},
		charTc{
			name:  "Combined iteration",
			input: []string{"hello", "world", "ha"},
			sep:   []byte{pathSep, domainSep, domainSep},
		},
	}
	for _, t := range tc {
		i := newIter(t.input, t.sep)
		var out []byte
		for {
			ch, _, ok := i.next()
			if !ok {
				break
			}
			out = append(out, ch)
		}
		c.Assert(string(out), Equals, t.String(), Commentf("%v", t.name))
	}
}

func (s *IterSuite) TestRecoverPosition(c *C) {
	i := newIter([]string{"hi", "world"}, []byte{pathSep, domainSep})
	i.next()
	i.next()
	p := i.position()
	i.next()
	i.setPosition(p)

	ch, sep, ok := i.next()
	c.Assert(ok, Equals, true)
	c.Assert(ch, Equals, byte('w'))
	c.Assert(sep, Equals, byte(domainSep))
}

func (s *IterSuite) TestPushBack(c *C) {
	i := newIter([]string{"hi", "world"}, []byte{pathSep, domainSep})
	i.pushBack()
	i.pushBack()
	ch, sep, ok := i.next()
	c.Assert(ok, Equals, true)
	c.Assert(ch, Equals, byte('h'))
	c.Assert(sep, Equals, byte(pathSep))
}

func (s *IterSuite) TestPushBackBoundary(c *C) {
	i := newIter([]string{"hi", "world"}, []byte{pathSep, domainSep})
	i.next()
	i.next()
	i.next()
	i.pushBack()
	i.pushBack()
	ch, sep, ok := i.next()
	c.Assert(ok, Equals, true)
	c.Assert(fmt.Sprintf("%c", ch), Equals, "i")
	c.Assert(fmt.Sprintf("%c", sep), Equals, fmt.Sprintf("%c", pathSep))
}

func (s *IterSuite) TestString(c *C) {
	i := newIter([]string{"hi"}, []byte{pathSep})
	i.next()
	c.Assert(i.String(), Equals, "<1:hi>")
	i.next()
	c.Assert(i.String(), Equals, "<end>")
}

type charTc struct {
	name  string
	input []string
	sep   []byte
}

func (c *charTc) String() string {
	return strings.Join(c.input, "")
}
