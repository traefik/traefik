// Copyright 2015 Brett Vickers.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package etree provides XML services through an Element Tree
// abstraction.
package etree

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	// NoIndent is used with Indent to disable all indenting.
	NoIndent = -1
)

// ErrXML is returned when XML parsing fails due to incorrect formatting.
var ErrXML = errors.New("etree: invalid XML format")

// ReadSettings allow for changing the default behavior of the ReadFrom*
// methods.
type ReadSettings struct {
	// CharsetReader to be passed to standard xml.Decoder. Default: nil.
	CharsetReader func(charset string, input io.Reader) (io.Reader, error)

	// Permissive allows input containing common mistakes such as missing tags
	// or attribute values. Default: false.
	Permissive bool
}

// newReadSettings creates a default ReadSettings record.
func newReadSettings() ReadSettings {
	return ReadSettings{}
}

// WriteSettings allow for changing the serialization behavior of the WriteTo*
// methods.
type WriteSettings struct {
	// CanonicalEndTags forces the production of XML end tags, even for
	// elements that have no child elements. Default: false.
	CanonicalEndTags bool

	// CanonicalText forces the production of XML character references for
	// text data characters &, <, and >. If false, XML character references
	// are also produced for " and '. Default: false.
	CanonicalText bool

	// CanonicalAttrVal forces the production of XML character references for
	// attribute value characters &, < and ". If false, XML character
	// references are also produced for > and '. Default: false.
	CanonicalAttrVal bool
}

// newWriteSettings creates a default WriteSettings record.
func newWriteSettings() WriteSettings {
	return WriteSettings{
		CanonicalEndTags: false,
		CanonicalText:    false,
		CanonicalAttrVal: false,
	}
}

// A Token is an empty interface that represents an Element, CharData,
// Comment, Directive, or ProcInst.
type Token interface {
	Parent() *Element
	dup(parent *Element) Token
	setParent(parent *Element)
	writeTo(w *bufio.Writer, s *WriteSettings)
}

// A Document is a container holding a complete XML hierarchy. Its embedded
// element contains zero or more children, one of which is usually the root
// element.  The embedded element may include other children such as
// processing instructions or BOM CharData tokens.
type Document struct {
	Element
	ReadSettings  ReadSettings
	WriteSettings WriteSettings
}

// An Element represents an XML element, its attributes, and its child tokens.
type Element struct {
	Space, Tag string   // namespace and tag
	Attr       []Attr   // key-value attribute pairs
	Child      []Token  // child tokens (elements, comments, etc.)
	parent     *Element // parent element
}

// An Attr represents a key-value attribute of an XML element.
type Attr struct {
	Space, Key string // The attribute's namespace and key
	Value      string // The attribute value string
}

// CharData represents character data within XML.
type CharData struct {
	Data       string
	parent     *Element
	whitespace bool
}

// A Comment represents an XML comment.
type Comment struct {
	Data   string
	parent *Element
}

// A Directive represents an XML directive.
type Directive struct {
	Data   string
	parent *Element
}

// A ProcInst represents an XML processing instruction.
type ProcInst struct {
	Target string
	Inst   string
	parent *Element
}

// NewDocument creates an XML document without a root element.
func NewDocument() *Document {
	return &Document{
		Element{Child: make([]Token, 0)},
		newReadSettings(),
		newWriteSettings(),
	}
}

// Copy returns a recursive, deep copy of the document.
func (d *Document) Copy() *Document {
	return &Document{*(d.dup(nil).(*Element)), d.ReadSettings, d.WriteSettings}
}

// Root returns the root element of the document, or nil if there is no root
// element.
func (d *Document) Root() *Element {
	for _, t := range d.Child {
		if c, ok := t.(*Element); ok {
			return c
		}
	}
	return nil
}

// SetRoot replaces the document's root element with e. If the document
// already has a root when this function is called, then the document's
// original root is unbound first. If the element e is bound to another
// document (or to another element within a document), then it is unbound
// first.
func (d *Document) SetRoot(e *Element) {
	if e.parent != nil {
		e.parent.RemoveChild(e)
	}
	e.setParent(&d.Element)

	for i, t := range d.Child {
		if _, ok := t.(*Element); ok {
			t.setParent(nil)
			d.Child[i] = e
			return
		}
	}
	d.Child = append(d.Child, e)
}

// ReadFrom reads XML from the reader r into the document d. It returns the
// number of bytes read and any error encountered.
func (d *Document) ReadFrom(r io.Reader) (n int64, err error) {
	return d.Element.readFrom(r, d.ReadSettings)
}

// ReadFromFile reads XML from the string s into the document d.
func (d *Document) ReadFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = d.ReadFrom(f)
	return err
}

// ReadFromBytes reads XML from the byte slice b into the document d.
func (d *Document) ReadFromBytes(b []byte) error {
	_, err := d.ReadFrom(bytes.NewReader(b))
	return err
}

// ReadFromString reads XML from the string s into the document d.
func (d *Document) ReadFromString(s string) error {
	_, err := d.ReadFrom(strings.NewReader(s))
	return err
}

// WriteTo serializes an XML document into the writer w. It
// returns the number of bytes written and any error encountered.
func (d *Document) WriteTo(w io.Writer) (n int64, err error) {
	cw := newCountWriter(w)
	b := bufio.NewWriter(cw)
	for _, c := range d.Child {
		c.writeTo(b, &d.WriteSettings)
	}
	err, n = b.Flush(), cw.bytes
	return
}

// WriteToFile serializes an XML document into the file named
// filename.
func (d *Document) WriteToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = d.WriteTo(f)
	return err
}

// WriteToBytes serializes the XML document into a slice of
// bytes.
func (d *Document) WriteToBytes() (b []byte, err error) {
	var buf bytes.Buffer
	if _, err = d.WriteTo(&buf); err != nil {
		return
	}
	return buf.Bytes(), nil
}

// WriteToString serializes the XML document into a string.
func (d *Document) WriteToString() (s string, err error) {
	var b []byte
	if b, err = d.WriteToBytes(); err != nil {
		return
	}
	return string(b), nil
}

type indentFunc func(depth int) string

// Indent modifies the document's element tree by inserting CharData entities
// containing carriage returns and indentation. The amount of indentation per
// depth level is given as spaces. Pass etree.NoIndent for spaces if you want
// no indentation at all.
func (d *Document) Indent(spaces int) {
	var indent indentFunc
	switch {
	case spaces < 0:
		indent = func(depth int) string { return "" }
	default:
		indent = func(depth int) string { return crIndent(depth*spaces, crsp) }
	}
	d.Element.indent(0, indent)
}

// IndentTabs modifies the document's element tree by inserting CharData
// entities containing carriage returns and tabs for indentation.  One tab is
// used per indentation level.
func (d *Document) IndentTabs() {
	indent := func(depth int) string { return crIndent(depth, crtab) }
	d.Element.indent(0, indent)
}

// NewElement creates an unparented element with the specified tag. The tag
// may be prefixed by a namespace and a colon.
func NewElement(tag string) *Element {
	space, stag := spaceDecompose(tag)
	return newElement(space, stag, nil)
}

// newElement is a helper function that creates an element and binds it to
// a parent element if possible.
func newElement(space, tag string, parent *Element) *Element {
	e := &Element{
		Space:  space,
		Tag:    tag,
		Attr:   make([]Attr, 0),
		Child:  make([]Token, 0),
		parent: parent,
	}
	if parent != nil {
		parent.addChild(e)
	}
	return e
}

// Copy creates a recursive, deep copy of the element and all its attributes
// and children. The returned element has no parent but can be parented to a
// another element using AddElement, or to a document using SetRoot.
func (e *Element) Copy() *Element {
	var parent *Element
	return e.dup(parent).(*Element)
}

// Text returns the characters immediately following the element's
// opening tag.
func (e *Element) Text() string {
	if len(e.Child) == 0 {
		return ""
	}
	if cd, ok := e.Child[0].(*CharData); ok {
		return cd.Data
	}
	return ""
}

// SetText replaces an element's subsidiary CharData text with a new string.
func (e *Element) SetText(text string) {
	if len(e.Child) > 0 {
		if cd, ok := e.Child[0].(*CharData); ok {
			cd.Data = text
			return
		}
	}
	cd := newCharData(text, false, e)
	copy(e.Child[1:], e.Child[0:])
	e.Child[0] = cd
}

// CreateElement creates an element with the specified tag and adds it as the
// last child element of the element e. The tag may be prefixed by a namespace
// and a colon.
func (e *Element) CreateElement(tag string) *Element {
	space, stag := spaceDecompose(tag)
	return newElement(space, stag, e)
}

// AddChild adds the token t as the last child of element e. If token t was
// already the child of another element, it is first removed from its current
// parent element.
func (e *Element) AddChild(t Token) {
	if t.Parent() != nil {
		t.Parent().RemoveChild(t)
	}
	t.setParent(e)
	e.addChild(t)
}

// InsertChild inserts the token t before e's existing child token ex. If ex
// is nil (or if ex is not a child of e), then t is added to the end of e's
// child token list. If token t was already the child of another element, it
// is first removed from its current parent element.
func (e *Element) InsertChild(ex Token, t Token) {
	if t.Parent() != nil {
		t.Parent().RemoveChild(t)
	}
	t.setParent(e)

	for i, c := range e.Child {
		if c == ex {
			e.Child = append(e.Child, nil)
			copy(e.Child[i+1:], e.Child[i:])
			e.Child[i] = t
			return
		}
	}
	e.addChild(t)
}

// RemoveChild attempts to remove the token t from element e's list of
// children. If the token t is a child of e, then it is returned. Otherwise,
// nil is returned.
func (e *Element) RemoveChild(t Token) Token {
	for i, c := range e.Child {
		if c == t {
			e.Child = append(e.Child[:i], e.Child[i+1:]...)
			c.setParent(nil)
			return t
		}
	}
	return nil
}

// ReadFrom reads XML from the reader r and stores the result as a new child
// of element e.
func (e *Element) readFrom(ri io.Reader, settings ReadSettings) (n int64, err error) {
	r := newCountReader(ri)
	dec := xml.NewDecoder(r)
	dec.CharsetReader = settings.CharsetReader
	dec.Strict = !settings.Permissive
	var stack stack
	stack.push(e)
	for {
		t, err := dec.RawToken()
		switch {
		case err == io.EOF:
			return r.bytes, nil
		case err != nil:
			return r.bytes, err
		case stack.empty():
			return r.bytes, ErrXML
		}

		top := stack.peek().(*Element)

		switch t := t.(type) {
		case xml.StartElement:
			e := newElement(t.Name.Space, t.Name.Local, top)
			for _, a := range t.Attr {
				e.createAttr(a.Name.Space, a.Name.Local, a.Value)
			}
			stack.push(e)
		case xml.EndElement:
			stack.pop()
		case xml.CharData:
			data := string(t)
			newCharData(data, isWhitespace(data), top)
		case xml.Comment:
			newComment(string(t), top)
		case xml.Directive:
			newDirective(string(t), top)
		case xml.ProcInst:
			newProcInst(t.Target, string(t.Inst), top)
		}
	}
}

// SelectAttr finds an element attribute matching the requested key and
// returns it if found. Returns nil if no matching attribute is found. The key
// may be prefixed by a namespace and a colon.
func (e *Element) SelectAttr(key string) *Attr {
	space, skey := spaceDecompose(key)
	for i, a := range e.Attr {
		if spaceMatch(space, a.Space) && skey == a.Key {
			return &e.Attr[i]
		}
	}
	return nil
}

// SelectAttrValue finds an element attribute matching the requested key and
// returns its value if found. The key may be prefixed by a namespace and a
// colon. If the key is not found, the dflt value is returned instead.
func (e *Element) SelectAttrValue(key, dflt string) string {
	space, skey := spaceDecompose(key)
	for _, a := range e.Attr {
		if spaceMatch(space, a.Space) && skey == a.Key {
			return a.Value
		}
	}
	return dflt
}

// ChildElements returns all elements that are children of element e.
func (e *Element) ChildElements() []*Element {
	var elements []*Element
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok {
			elements = append(elements, c)
		}
	}
	return elements
}

// SelectElement returns the first child element with the given tag. The tag
// may be prefixed by a namespace and a colon. Returns nil if no element with
// a matching tag was found.
func (e *Element) SelectElement(tag string) *Element {
	space, stag := spaceDecompose(tag)
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok && spaceMatch(space, c.Space) && stag == c.Tag {
			return c
		}
	}
	return nil
}

// SelectElements returns a slice of all child elements with the given tag.
// The tag may be prefixed by a namespace and a colon.
func (e *Element) SelectElements(tag string) []*Element {
	space, stag := spaceDecompose(tag)
	var elements []*Element
	for _, t := range e.Child {
		if c, ok := t.(*Element); ok && spaceMatch(space, c.Space) && stag == c.Tag {
			elements = append(elements, c)
		}
	}
	return elements
}

// FindElement returns the first element matched by the XPath-like path
// string. Returns nil if no element is found using the path. Panics if an
// invalid path string is supplied.
func (e *Element) FindElement(path string) *Element {
	return e.FindElementPath(MustCompilePath(path))
}

// FindElementPath returns the first element matched by the XPath-like path
// string. Returns nil if no element is found using the path.
func (e *Element) FindElementPath(path Path) *Element {
	p := newPather()
	elements := p.traverse(e, path)
	switch {
	case len(elements) > 0:
		return elements[0]
	default:
		return nil
	}
}

// FindElements returns a slice of elements matched by the XPath-like path
// string. Panics if an invalid path string is supplied.
func (e *Element) FindElements(path string) []*Element {
	return e.FindElementsPath(MustCompilePath(path))
}

// FindElementsPath returns a slice of elements matched by the Path object.
func (e *Element) FindElementsPath(path Path) []*Element {
	p := newPather()
	return p.traverse(e, path)
}

// GetPath returns the absolute path of the element.
func (e *Element) GetPath() string {
	path := []string{}
	for seg := e; seg != nil; seg = seg.Parent() {
		if seg.Tag != "" {
			path = append(path, seg.Tag)
		}
	}

	// Reverse the path.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return "/" + strings.Join(path, "/")
}

// GetRelativePath returns the path of the element relative to the source
// element. If the two elements are not part of the same element tree, then
// GetRelativePath returns the empty string.
func (e *Element) GetRelativePath(source *Element) string {
	var path []*Element

	if source == nil {
		return ""
	}

	// Build a reverse path from the element toward the root. Stop if the
	// source element is encountered.
	var seg *Element
	for seg = e; seg != nil && seg != source; seg = seg.Parent() {
		path = append(path, seg)
	}

	// If we found the source element, reverse the path and compose the
	// string.
	if seg == source {
		if len(path) == 0 {
			return "."
		}
		parts := []string{}
		for i := len(path) - 1; i >= 0; i-- {
			parts = append(parts, path[i].Tag)
		}
		return "./" + strings.Join(parts, "/")
	}

	// The source wasn't encountered, so climb from the source element toward
	// the root of the tree until an element in the reversed path is
	// encountered.

	findPathIndex := func(e *Element, path []*Element) int {
		for i, ee := range path {
			if e == ee {
				return i
			}
		}
		return -1
	}

	climb := 0
	for seg = source; seg != nil; seg = seg.Parent() {
		i := findPathIndex(seg, path)
		if i >= 0 {
			path = path[:i] // truncate at found segment
			break
		}
		climb++
	}

	// No element in the reversed path was encountered, so the two elements
	// must not be part of the same tree.
	if seg == nil {
		return ""
	}

	// Reverse the (possibly truncated) path and prepend ".." segments to
	// climb.
	parts := []string{}
	for i := 0; i < climb; i++ {
		parts = append(parts, "..")
	}
	for i := len(path) - 1; i >= 0; i-- {
		parts = append(parts, path[i].Tag)
	}
	return strings.Join(parts, "/")
}

// indent recursively inserts proper indentation between an
// XML element's child tokens.
func (e *Element) indent(depth int, indent indentFunc) {
	e.stripIndent()
	n := len(e.Child)
	if n == 0 {
		return
	}

	oldChild := e.Child
	e.Child = make([]Token, 0, n*2+1)
	isCharData, firstNonCharData := false, true
	for _, c := range oldChild {

		// Insert CR+indent before child if it's not character data.
		// Exceptions: when it's the first non-character-data child, or when
		// the child is at root depth.
		_, isCharData = c.(*CharData)
		if !isCharData {
			if !firstNonCharData || depth > 0 {
				newCharData(indent(depth), true, e)
			}
			firstNonCharData = false
		}

		e.addChild(c)

		// Recursively process child elements.
		if ce, ok := c.(*Element); ok {
			ce.indent(depth+1, indent)
		}
	}

	// Insert CR+indent before the last child.
	if !isCharData {
		if !firstNonCharData || depth > 0 {
			newCharData(indent(depth-1), true, e)
		}
	}
}

// stripIndent removes any previously inserted indentation.
func (e *Element) stripIndent() {
	// Count the number of non-indent child tokens
	n := len(e.Child)
	for _, c := range e.Child {
		if cd, ok := c.(*CharData); ok && cd.whitespace {
			n--
		}
	}
	if n == len(e.Child) {
		return
	}

	// Strip out indent CharData
	newChild := make([]Token, n)
	j := 0
	for _, c := range e.Child {
		if cd, ok := c.(*CharData); ok && cd.whitespace {
			continue
		}
		newChild[j] = c
		j++
	}
	e.Child = newChild
}

// dup duplicates the element.
func (e *Element) dup(parent *Element) Token {
	ne := &Element{
		Space:  e.Space,
		Tag:    e.Tag,
		Attr:   make([]Attr, len(e.Attr)),
		Child:  make([]Token, len(e.Child)),
		parent: parent,
	}
	for i, t := range e.Child {
		ne.Child[i] = t.dup(ne)
	}
	for i, a := range e.Attr {
		ne.Attr[i] = a
	}
	return ne
}

// Parent returns the element token's parent element, or nil if it has no
// parent.
func (e *Element) Parent() *Element {
	return e.parent
}

// setParent replaces the element token's parent.
func (e *Element) setParent(parent *Element) {
	e.parent = parent
}

// writeTo serializes the element to the writer w.
func (e *Element) writeTo(w *bufio.Writer, s *WriteSettings) {
	w.WriteByte('<')
	if e.Space != "" {
		w.WriteString(e.Space)
		w.WriteByte(':')
	}
	w.WriteString(e.Tag)
	for _, a := range e.Attr {
		w.WriteByte(' ')
		a.writeTo(w, s)
	}
	if len(e.Child) > 0 {
		w.WriteString(">")
		for _, c := range e.Child {
			c.writeTo(w, s)
		}
		w.Write([]byte{'<', '/'})
		if e.Space != "" {
			w.WriteString(e.Space)
			w.WriteByte(':')
		}
		w.WriteString(e.Tag)
		w.WriteByte('>')
	} else {
		if s.CanonicalEndTags {
			w.Write([]byte{'>', '<', '/'})
			if e.Space != "" {
				w.WriteString(e.Space)
				w.WriteByte(':')
			}
			w.WriteString(e.Tag)
			w.WriteByte('>')
		} else {
			w.Write([]byte{'/', '>'})
		}
	}
}

// addChild adds a child token to the element e.
func (e *Element) addChild(t Token) {
	e.Child = append(e.Child, t)
}

// CreateAttr creates an attribute and adds it to element e. The key may be
// prefixed by a namespace and a colon. If an attribute with the key already
// exists, its value is replaced.
func (e *Element) CreateAttr(key, value string) *Attr {
	space, skey := spaceDecompose(key)
	return e.createAttr(space, skey, value)
}

// createAttr is a helper function that creates attributes.
func (e *Element) createAttr(space, key, value string) *Attr {
	for i, a := range e.Attr {
		if space == a.Space && key == a.Key {
			e.Attr[i].Value = value
			return &e.Attr[i]
		}
	}
	a := Attr{space, key, value}
	e.Attr = append(e.Attr, a)
	return &e.Attr[len(e.Attr)-1]
}

// RemoveAttr removes and returns the first attribute of the element whose key
// matches the given key. The key may be prefixed by a namespace and a colon.
// If an equal attribute does not exist, nil is returned.
func (e *Element) RemoveAttr(key string) *Attr {
	space, skey := spaceDecompose(key)
	for i, a := range e.Attr {
		if space == a.Space && skey == a.Key {
			e.Attr = append(e.Attr[0:i], e.Attr[i+1:]...)
			return &a
		}
	}
	return nil
}

var xmlReplacerNormal = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"'", "&apos;",
	`"`, "&quot;",
)

var xmlReplacerCanonicalText = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\r", "&#xD;",
)

var xmlReplacerCanonicalAttrVal = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	`"`, "&quot;",
	"\t", "&#x9;",
	"\n", "&#xA;",
	"\r", "&#xD;",
)

// writeTo serializes the attribute to the writer.
func (a *Attr) writeTo(w *bufio.Writer, s *WriteSettings) {
	if a.Space != "" {
		w.WriteString(a.Space)
		w.WriteByte(':')
	}
	w.WriteString(a.Key)
	w.WriteString(`="`)
	var r *strings.Replacer
	if s.CanonicalAttrVal {
		r = xmlReplacerCanonicalAttrVal
	} else {
		r = xmlReplacerNormal
	}
	w.WriteString(r.Replace(a.Value))
	w.WriteByte('"')
}

// NewCharData creates a parentless XML character data entity.
func NewCharData(data string) *CharData {
	return newCharData(data, false, nil)
}

// newCharData creates an XML character data entity and binds it to a parent
// element. If parent is nil, the CharData token remains unbound.
func newCharData(data string, whitespace bool, parent *Element) *CharData {
	c := &CharData{
		Data:       data,
		whitespace: whitespace,
		parent:     parent,
	}
	if parent != nil {
		parent.addChild(c)
	}
	return c
}

// CreateCharData creates an XML character data entity and adds it as a child
// of element e.
func (e *Element) CreateCharData(data string) *CharData {
	return newCharData(data, false, e)
}

// dup duplicates the character data.
func (c *CharData) dup(parent *Element) Token {
	return &CharData{
		Data:       c.Data,
		whitespace: c.whitespace,
		parent:     parent,
	}
}

// Parent returns the character data token's parent element, or nil if it has
// no parent.
func (c *CharData) Parent() *Element {
	return c.parent
}

// setParent replaces the character data token's parent.
func (c *CharData) setParent(parent *Element) {
	c.parent = parent
}

// writeTo serializes the character data entity to the writer.
func (c *CharData) writeTo(w *bufio.Writer, s *WriteSettings) {
	var r *strings.Replacer
	if s.CanonicalText {
		r = xmlReplacerCanonicalText
	} else {
		r = xmlReplacerNormal
	}
	w.WriteString(r.Replace(c.Data))
}

// NewComment creates a parentless XML comment.
func NewComment(comment string) *Comment {
	return newComment(comment, nil)
}

// NewComment creates an XML comment and binds it to a parent element. If
// parent is nil, the Comment remains unbound.
func newComment(comment string, parent *Element) *Comment {
	c := &Comment{
		Data:   comment,
		parent: parent,
	}
	if parent != nil {
		parent.addChild(c)
	}
	return c
}

// CreateComment creates an XML comment and adds it as a child of element e.
func (e *Element) CreateComment(comment string) *Comment {
	return newComment(comment, e)
}

// dup duplicates the comment.
func (c *Comment) dup(parent *Element) Token {
	return &Comment{
		Data:   c.Data,
		parent: parent,
	}
}

// Parent returns comment token's parent element, or nil if it has no parent.
func (c *Comment) Parent() *Element {
	return c.parent
}

// setParent replaces the comment token's parent.
func (c *Comment) setParent(parent *Element) {
	c.parent = parent
}

// writeTo serialies the comment to the writer.
func (c *Comment) writeTo(w *bufio.Writer, s *WriteSettings) {
	w.WriteString("<!--")
	w.WriteString(c.Data)
	w.WriteString("-->")
}

// NewDirective creates a parentless XML directive.
func NewDirective(data string) *Directive {
	return newDirective(data, nil)
}

// newDirective creates an XML directive and binds it to a parent element. If
// parent is nil, the Directive remains unbound.
func newDirective(data string, parent *Element) *Directive {
	d := &Directive{
		Data:   data,
		parent: parent,
	}
	if parent != nil {
		parent.addChild(d)
	}
	return d
}

// CreateDirective creates an XML directive and adds it as the last child of
// element e.
func (e *Element) CreateDirective(data string) *Directive {
	return newDirective(data, e)
}

// dup duplicates the directive.
func (d *Directive) dup(parent *Element) Token {
	return &Directive{
		Data:   d.Data,
		parent: parent,
	}
}

// Parent returns directive token's parent element, or nil if it has no
// parent.
func (d *Directive) Parent() *Element {
	return d.parent
}

// setParent replaces the directive token's parent.
func (d *Directive) setParent(parent *Element) {
	d.parent = parent
}

// writeTo serializes the XML directive to the writer.
func (d *Directive) writeTo(w *bufio.Writer, s *WriteSettings) {
	w.WriteString("<!")
	w.WriteString(d.Data)
	w.WriteString(">")
}

// NewProcInst creates a parentless XML processing instruction.
func NewProcInst(target, inst string) *ProcInst {
	return newProcInst(target, inst, nil)
}

// newProcInst creates an XML processing instruction and binds it to a parent
// element. If parent is nil, the ProcInst remains unbound.
func newProcInst(target, inst string, parent *Element) *ProcInst {
	p := &ProcInst{
		Target: target,
		Inst:   inst,
		parent: parent,
	}
	if parent != nil {
		parent.addChild(p)
	}
	return p
}

// CreateProcInst creates a processing instruction and adds it as a child of
// element e.
func (e *Element) CreateProcInst(target, inst string) *ProcInst {
	return newProcInst(target, inst, e)
}

// dup duplicates the procinst.
func (p *ProcInst) dup(parent *Element) Token {
	return &ProcInst{
		Target: p.Target,
		Inst:   p.Inst,
		parent: parent,
	}
}

// Parent returns processing instruction token's parent element, or nil if it
// has no parent.
func (p *ProcInst) Parent() *Element {
	return p.parent
}

// setParent replaces the processing instruction token's parent.
func (p *ProcInst) setParent(parent *Element) {
	p.parent = parent
}

// writeTo serializes the processing instruction to the writer.
func (p *ProcInst) writeTo(w *bufio.Writer, s *WriteSettings) {
	w.WriteString("<?")
	w.WriteString(p.Target)
	if p.Inst != "" {
		w.WriteByte(' ')
		w.WriteString(p.Inst)
	}
	w.WriteString("?>")
}
