// Copyright 2015 Brett Vickers.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etree

import (
	"strconv"
	"strings"
)

/*
A Path is an object that represents an optimized version of an
XPath-like search string.  Although path strings are XPath-like,
only the following limited syntax is supported:

    .               Selects the current element
    ..              Selects the parent of the current element
    *               Selects all child elements
    //              Selects all descendants of the current element
    tag             Selects all child elements with the given tag
    [#]             Selects the element of the given index (1-based,
                      negative starts from the end)
    [@attrib]       Selects all elements with the given attribute
    [@attrib='val'] Selects all elements with the given attribute set to val
    [tag]           Selects all elements with a child element named tag
    [tag='val']     Selects all elements with a child element named tag
                      and text matching val
    [text()]        Selects all elements with non-empty text
    [text()='val']  Selects all elements whose text matches val

Examples:

Select the title elements of all descendant book elements having a
'category' attribute of 'WEB':
    //book[@category='WEB']/title

Select the first book element with a title child containing the text
'Great Expectations':
    .//book[title='Great Expectations'][1]

Starting from the current element, select all children of book elements
with an attribute 'language' set to 'english':
    ./book/*[@language='english']

Starting from the current element, select all children of book elements
containing the text 'special':
    ./book/*[text()='special']

Select all descendant book elements whose title element has an attribute
'language' set to 'french':
    //book/title[@language='french']/..

*/
type Path struct {
	segments []segment
}

// ErrPath is returned by path functions when an invalid etree path is provided.
type ErrPath string

// Error returns the string describing a path error.
func (err ErrPath) Error() string {
	return "etree: " + string(err)
}

// CompilePath creates an optimized version of an XPath-like string that
// can be used to query elements in an element tree.
func CompilePath(path string) (Path, error) {
	var comp compiler
	segments := comp.parsePath(path)
	if comp.err != ErrPath("") {
		return Path{nil}, comp.err
	}
	return Path{segments}, nil
}

// MustCompilePath creates an optimized version of an XPath-like string that
// can be used to query elements in an element tree.  Panics if an error
// occurs.  Use this function to create Paths when you know the path is
// valid (i.e., if it's hard-coded).
func MustCompilePath(path string) Path {
	p, err := CompilePath(path)
	if err != nil {
		panic(err)
	}
	return p
}

// A segment is a portion of a path between "/" characters.
// It contains one selector and zero or more [filters].
type segment struct {
	sel     selector
	filters []filter
}

func (seg *segment) apply(e *Element, p *pather) {
	seg.sel.apply(e, p)
	for _, f := range seg.filters {
		f.apply(p)
	}
}

// A selector selects XML elements for consideration by the
// path traversal.
type selector interface {
	apply(e *Element, p *pather)
}

// A filter pares down a list of candidate XML elements based
// on a path filter in [brackets].
type filter interface {
	apply(p *pather)
}

// A pather is helper object that traverses an element tree using
// a Path object.  It collects and deduplicates all elements matching
// the path query.
type pather struct {
	queue      fifo
	results    []*Element
	inResults  map[*Element]bool
	candidates []*Element
	scratch    []*Element // used by filters
}

// A node represents an element and the remaining path segments that
// should be applied against it by the pather.
type node struct {
	e        *Element
	segments []segment
}

func newPather() *pather {
	return &pather{
		results:    make([]*Element, 0),
		inResults:  make(map[*Element]bool),
		candidates: make([]*Element, 0),
		scratch:    make([]*Element, 0),
	}
}

// traverse follows the path from the element e, collecting
// and then returning all elements that match the path's selectors
// and filters.
func (p *pather) traverse(e *Element, path Path) []*Element {
	for p.queue.add(node{e, path.segments}); p.queue.len() > 0; {
		p.eval(p.queue.remove().(node))
	}
	return p.results
}

// eval evalutes the current path node by applying the remaining
// path's selector rules against the node's element.
func (p *pather) eval(n node) {
	p.candidates = p.candidates[0:0]
	seg, remain := n.segments[0], n.segments[1:]
	seg.apply(n.e, p)

	if len(remain) == 0 {
		for _, c := range p.candidates {
			if in := p.inResults[c]; !in {
				p.inResults[c] = true
				p.results = append(p.results, c)
			}
		}
	} else {
		for _, c := range p.candidates {
			p.queue.add(node{c, remain})
		}
	}
}

// A compiler generates a compiled path from a path string.
type compiler struct {
	err ErrPath
}

// parsePath parses an XPath-like string describing a path
// through an element tree and returns a slice of segment
// descriptors.
func (c *compiler) parsePath(path string) []segment {
	// If path starts or ends with //, fix it
	if strings.HasPrefix(path, "//") {
		path = "." + path
	}
	if strings.HasSuffix(path, "//") {
		path = path + "*"
	}

	// Paths cannot be absolute
	if strings.HasPrefix(path, "/") {
		c.err = ErrPath("paths cannot be absolute.")
		return nil
	}

	// Split path into segment objects
	var segments []segment
	for _, s := range splitPath(path) {
		segments = append(segments, c.parseSegment(s))
		if c.err != ErrPath("") {
			break
		}
	}
	return segments
}

func splitPath(path string) []string {
	pieces := make([]string, 0)
	start := 0
	inquote := false
	for i := 0; i+1 <= len(path); i++ {
		if path[i] == '\'' {
			inquote = !inquote
		} else if path[i] == '/' && !inquote {
			pieces = append(pieces, path[start:i])
			start = i + 1
		}
	}
	return append(pieces, path[start:])
}

// parseSegment parses a path segment between / characters.
func (c *compiler) parseSegment(path string) segment {
	pieces := strings.Split(path, "[")
	seg := segment{
		sel:     c.parseSelector(pieces[0]),
		filters: make([]filter, 0),
	}
	for i := 1; i < len(pieces); i++ {
		fpath := pieces[i]
		if fpath[len(fpath)-1] != ']' {
			c.err = ErrPath("path has invalid filter [brackets].")
			break
		}
		seg.filters = append(seg.filters, c.parseFilter(fpath[:len(fpath)-1]))
	}
	return seg
}

// parseSelector parses a selector at the start of a path segment.
func (c *compiler) parseSelector(path string) selector {
	switch path {
	case ".":
		return new(selectSelf)
	case "..":
		return new(selectParent)
	case "*":
		return new(selectChildren)
	case "":
		return new(selectDescendants)
	default:
		return newSelectChildrenByTag(path)
	}
}

// parseFilter parses a path filter contained within [brackets].
func (c *compiler) parseFilter(path string) filter {
	if len(path) == 0 {
		c.err = ErrPath("path contains an empty filter expression.")
		return nil
	}

	// Filter contains [@attr='val'], [text()='val'], or [tag='val']?
	eqindex := strings.Index(path, "='")
	if eqindex >= 0 {
		rindex := nextIndex(path, "'", eqindex+2)
		if rindex != len(path)-1 {
			c.err = ErrPath("path has mismatched filter quotes.")
			return nil
		}
		switch {
		case path[0] == '@':
			return newFilterAttrVal(path[1:eqindex], path[eqindex+2:rindex])
		case strings.HasPrefix(path, "text()"):
			return newFilterTextVal(path[eqindex+2 : rindex])
		default:
			return newFilterChildText(path[:eqindex], path[eqindex+2:rindex])
		}
	}

	// Filter contains [@attr], [N], [tag] or [text()]
	switch {
	case path[0] == '@':
		return newFilterAttr(path[1:])
	case path == "text()":
		return newFilterText()
	case isInteger(path):
		pos, _ := strconv.Atoi(path)
		switch {
		case pos > 0:
			return newFilterPos(pos - 1)
		default:
			return newFilterPos(pos)
		}
	default:
		return newFilterChild(path)
	}
}

// selectSelf selects the current element into the candidate list.
type selectSelf struct{}

func (s *selectSelf) apply(e *Element, p *pather) {
	p.candidates = append(p.candidates, e)
}

// selectParent selects the element's parent into the candidate list.
type selectParent struct{}

func (s *selectParent) apply(e *Element, p *pather) {
	if e.parent != nil {
		p.candidates = append(p.candidates, e.parent)
	}
}

// selectChildren selects the element's child elements into the
// candidate list.
type selectChildren struct{}

func (s *selectChildren) apply(e *Element, p *pather) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok {
			p.candidates = append(p.candidates, c)
		}
	}
}

// selectDescendants selects all descendant child elements
// of the element into the candidate list.
type selectDescendants struct{}

func (s *selectDescendants) apply(e *Element, p *pather) {
	var queue fifo
	for queue.add(e); queue.len() > 0; {
		e := queue.remove().(*Element)
		p.candidates = append(p.candidates, e)
		for _, c := range e.Child {
			if c, ok := c.(*Element); ok {
				queue.add(c)
			}
		}
	}
}

// selectChildrenByTag selects into the candidate list all child
// elements of the element having the specified tag.
type selectChildrenByTag struct {
	space, tag string
}

func newSelectChildrenByTag(path string) *selectChildrenByTag {
	s, l := spaceDecompose(path)
	return &selectChildrenByTag{s, l}
}

func (s *selectChildrenByTag) apply(e *Element, p *pather) {
	for _, c := range e.Child {
		if c, ok := c.(*Element); ok && spaceMatch(s.space, c.Space) && s.tag == c.Tag {
			p.candidates = append(p.candidates, c)
		}
	}
}

// filterPos filters the candidate list, keeping only the
// candidate at the specified index.
type filterPos struct {
	index int
}

func newFilterPos(pos int) *filterPos {
	return &filterPos{pos}
}

func (f *filterPos) apply(p *pather) {
	if f.index >= 0 {
		if f.index < len(p.candidates) {
			p.scratch = append(p.scratch, p.candidates[f.index])
		}
	} else {
		if -f.index <= len(p.candidates) {
			p.scratch = append(p.scratch, p.candidates[len(p.candidates)+f.index])
		}
	}
	p.candidates, p.scratch = p.scratch, p.candidates[0:0]
}

// filterAttr filters the candidate list for elements having
// the specified attribute.
type filterAttr struct {
	space, key string
}

func newFilterAttr(str string) *filterAttr {
	s, l := spaceDecompose(str)
	return &filterAttr{s, l}
}

func (f *filterAttr) apply(p *pather) {
	for _, c := range p.candidates {
		for _, a := range c.Attr {
			if spaceMatch(f.space, a.Space) && f.key == a.Key {
				p.scratch = append(p.scratch, c)
				break
			}
		}
	}
	p.candidates, p.scratch = p.scratch, p.candidates[0:0]
}

// filterAttrVal filters the candidate list for elements having
// the specified attribute with the specified value.
type filterAttrVal struct {
	space, key, val string
}

func newFilterAttrVal(str, value string) *filterAttrVal {
	s, l := spaceDecompose(str)
	return &filterAttrVal{s, l, value}
}

func (f *filterAttrVal) apply(p *pather) {
	for _, c := range p.candidates {
		for _, a := range c.Attr {
			if spaceMatch(f.space, a.Space) && f.key == a.Key && f.val == a.Value {
				p.scratch = append(p.scratch, c)
				break
			}
		}
	}
	p.candidates, p.scratch = p.scratch, p.candidates[0:0]
}

// filterText filters the candidate list for elements having text.
type filterText struct{}

func newFilterText() *filterText {
	return &filterText{}
}

func (f *filterText) apply(p *pather) {
	for _, c := range p.candidates {
		if c.Text() != "" {
			p.scratch = append(p.scratch, c)
		}
	}
	p.candidates, p.scratch = p.scratch, p.candidates[0:0]
}

// filterTextVal filters the candidate list for elements having
// text equal to the specified value.
type filterTextVal struct {
	val string
}

func newFilterTextVal(value string) *filterTextVal {
	return &filterTextVal{value}
}

func (f *filterTextVal) apply(p *pather) {
	for _, c := range p.candidates {
		if c.Text() == f.val {
			p.scratch = append(p.scratch, c)
		}
	}
	p.candidates, p.scratch = p.scratch, p.candidates[0:0]
}

// filterChild filters the candidate list for elements having
// a child element with the specified tag.
type filterChild struct {
	space, tag string
}

func newFilterChild(str string) *filterChild {
	s, l := spaceDecompose(str)
	return &filterChild{s, l}
}

func (f *filterChild) apply(p *pather) {
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok &&
				spaceMatch(f.space, cc.Space) &&
				f.tag == cc.Tag {
				p.scratch = append(p.scratch, c)
			}
		}
	}
	p.candidates, p.scratch = p.scratch, p.candidates[0:0]
}

// filterChildText filters the candidate list for elements having
// a child element with the specified tag and text.
type filterChildText struct {
	space, tag, text string
}

func newFilterChildText(str, text string) *filterChildText {
	s, l := spaceDecompose(str)
	return &filterChildText{s, l, text}
}

func (f *filterChildText) apply(p *pather) {
	for _, c := range p.candidates {
		for _, cc := range c.Child {
			if cc, ok := cc.(*Element); ok &&
				spaceMatch(f.space, cc.Space) &&
				f.tag == cc.Tag &&
				f.text == cc.Text() {
				p.scratch = append(p.scratch, c)
			}
		}
	}
	p.candidates, p.scratch = p.scratch, p.candidates[0:0]
}
