// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package asserts

import (
	"bufio"
	"bytes"
	"crypto"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type typeFlags int

const (
	noAuthority typeFlags = iota + 1
)

// AssertionType describes a known assertion type with its name and metadata.
type AssertionType struct {
	// Name of the type.
	Name string
	// PrimaryKey holds the names of the headers that constitute the
	// unique primary key for this assertion type.
	PrimaryKey []string

	assembler func(assert assertionBase) (Assertion, error)
	flags     typeFlags
}

// MaxSupportedFormat returns the maximum supported format iteration for the type.
func (at *AssertionType) MaxSupportedFormat() int {
	return maxSupportedFormat[at.Name]
}

// Understood assertion types.
var (
	AccountType         = &AssertionType{"account", []string{"account-id"}, assembleAccount, 0}
	AccountKeyType      = &AssertionType{"account-key", []string{"public-key-sha3-384"}, assembleAccountKey, 0}
	RepairType          = &AssertionType{"repair", []string{"brand-id", "repair-id"}, assembleRepair, 0}
	ModelType           = &AssertionType{"model", []string{"series", "brand-id", "model"}, assembleModel, 0}
	SerialType          = &AssertionType{"serial", []string{"brand-id", "model", "serial"}, assembleSerial, 0}
	BaseDeclarationType = &AssertionType{"base-declaration", []string{"series"}, assembleBaseDeclaration, 0}
	SnapDeclarationType = &AssertionType{"snap-declaration", []string{"series", "snap-id"}, assembleSnapDeclaration, 0}
	SnapBuildType       = &AssertionType{"snap-build", []string{"snap-sha3-384"}, assembleSnapBuild, 0}
	SnapRevisionType    = &AssertionType{"snap-revision", []string{"snap-sha3-384"}, assembleSnapRevision, 0}
	SnapDeveloperType   = &AssertionType{"snap-developer", []string{"snap-id", "publisher-id"}, assembleSnapDeveloper, 0}
	SystemUserType      = &AssertionType{"system-user", []string{"brand-id", "email"}, assembleSystemUser, 0}
	ValidationType      = &AssertionType{"validation", []string{"series", "snap-id", "approved-snap-id", "approved-snap-revision"}, assembleValidation, 0}
	StoreType           = &AssertionType{"store", []string{"store"}, assembleStore, 0}

// ...
)

// Assertion types without a definite authority set (on the wire and/or self-signed).
var (
	DeviceSessionRequestType = &AssertionType{"device-session-request", []string{"brand-id", "model", "serial"}, assembleDeviceSessionRequest, noAuthority}
	SerialRequestType        = &AssertionType{"serial-request", nil, assembleSerialRequest, noAuthority}
	AccountKeyRequestType    = &AssertionType{"account-key-request", []string{"public-key-sha3-384"}, assembleAccountKeyRequest, noAuthority}
)

var typeRegistry = map[string]*AssertionType{
	AccountType.Name:         AccountType,
	AccountKeyType.Name:      AccountKeyType,
	ModelType.Name:           ModelType,
	SerialType.Name:          SerialType,
	BaseDeclarationType.Name: BaseDeclarationType,
	SnapDeclarationType.Name: SnapDeclarationType,
	SnapBuildType.Name:       SnapBuildType,
	SnapRevisionType.Name:    SnapRevisionType,
	SnapDeveloperType.Name:   SnapDeveloperType,
	SystemUserType.Name:      SystemUserType,
	ValidationType.Name:      ValidationType,
	RepairType.Name:          RepairType,
	StoreType.Name:           StoreType,
	// no authority
	DeviceSessionRequestType.Name: DeviceSessionRequestType,
	SerialRequestType.Name:        SerialRequestType,
	AccountKeyRequestType.Name:    AccountKeyRequestType,
}

// Type returns the AssertionType with name or nil
func Type(name string) *AssertionType {
	return typeRegistry[name]
}

// TypeNames returns a sorted list of known assertion type names.
func TypeNames() []string {
	names := make([]string, 0, len(typeRegistry))
	for k := range typeRegistry {
		names = append(names, k)
	}

	sort.Strings(names)

	return names
}

var maxSupportedFormat = map[string]int{}

func init() {
	// register maxSupportedFormats while breaking initialisation loop

	// 1: plugs and slots
	// 2: support for $SLOT()/$PLUG()/$MISSING
	// 3: support for on-store/on-brand/on-model device scope constraints
	maxSupportedFormat[SnapDeclarationType.Name] = 3
}

func MockMaxSupportedFormat(assertType *AssertionType, maxFormat int) (restore func()) {
	prev := maxSupportedFormat[assertType.Name]
	maxSupportedFormat[assertType.Name] = maxFormat
	return func() {
		maxSupportedFormat[assertType.Name] = prev
	}
}

var formatAnalyzer = map[*AssertionType]func(headers map[string]interface{}, body []byte) (formatnum int, err error){
	SnapDeclarationType: snapDeclarationFormatAnalyze,
}

// SuggestFormat returns a minimum format that supports the features that would be used by an assertion with the given components.
func SuggestFormat(assertType *AssertionType, headers map[string]interface{}, body []byte) (formatnum int, err error) {
	analyzer := formatAnalyzer[assertType]
	if analyzer == nil {
		// no analyzer, format 0 is all there is
		return 0, nil
	}
	formatnum, err = analyzer(headers, body)
	if err != nil {
		return 0, fmt.Errorf("assertion %s: %v", assertType.Name, err)
	}
	return formatnum, nil
}

// HeadersFromPrimaryKey constructs a headers mapping from the
// primaryKey values and the assertion type, it errors if primaryKey
// has the wrong length.
func HeadersFromPrimaryKey(assertType *AssertionType, primaryKey []string) (headers map[string]string, err error) {
	if len(primaryKey) != len(assertType.PrimaryKey) {
		return nil, fmt.Errorf("primary key has wrong length for %q assertion", assertType.Name)
	}
	headers = make(map[string]string, len(assertType.PrimaryKey))
	for i, name := range assertType.PrimaryKey {
		keyVal := primaryKey[i]
		if keyVal == "" {
			return nil, fmt.Errorf("primary key %q header cannot be empty", name)
		}
		headers[name] = keyVal
	}
	return headers, nil
}

// PrimaryKeyFromHeaders extracts the tuple of values from headers
// corresponding to a primary key under the assertion type, it errors
// if there are missing primary key headers.
func PrimaryKeyFromHeaders(assertType *AssertionType, headers map[string]string) (primaryKey []string, err error) {
	primaryKey = make([]string, len(assertType.PrimaryKey))
	for i, k := range assertType.PrimaryKey {
		keyVal := headers[k]
		if keyVal == "" {
			return nil, fmt.Errorf("must provide primary key: %v", k)
		}
		primaryKey[i] = keyVal
	}
	return primaryKey, nil
}

// Ref expresses a reference to an assertion.
type Ref struct {
	Type       *AssertionType
	PrimaryKey []string
}

func (ref *Ref) String() string {
	pkStr := "-"
	n := len(ref.Type.PrimaryKey)
	if n != len(ref.PrimaryKey) {
		pkStr = "???"
	} else if n > 0 {
		pkStr = ref.PrimaryKey[n-1]
		if n > 1 {
			sfx := []string{pkStr + ";"}
			for i, k := range ref.Type.PrimaryKey[:n-1] {
				sfx = append(sfx, fmt.Sprintf("%s:%s", k, ref.PrimaryKey[i]))
			}
			pkStr = strings.Join(sfx, " ")
		}
	}
	return fmt.Sprintf("%s (%s)", ref.Type.Name, pkStr)
}

// Unique returns a unique string representing the reference that can be used as a key in maps.
func (ref *Ref) Unique() string {
	return fmt.Sprintf("%s/%s", ref.Type.Name, strings.Join(ref.PrimaryKey, "/"))
}

// Resolve resolves the reference using the given find function.
func (ref *Ref) Resolve(find func(assertType *AssertionType, headers map[string]string) (Assertion, error)) (Assertion, error) {
	headers, err := HeadersFromPrimaryKey(ref.Type, ref.PrimaryKey)
	if err != nil {
		return nil, fmt.Errorf("%q assertion reference primary key has the wrong length (expected %v): %v", ref.Type.Name, ref.Type.PrimaryKey, ref.PrimaryKey)
	}
	return find(ref.Type, headers)
}

// Assertion represents an assertion through its general elements.
type Assertion interface {
	// Type returns the type of this assertion
	Type() *AssertionType
	// Format returns the format iteration of this assertion
	Format() int
	// SupportedFormat returns whether the assertion uses a supported
	// format iteration. If false the assertion might have been only
	// partially parsed.
	SupportedFormat() bool
	// Revision returns the revision of this assertion
	Revision() int
	// AuthorityID returns the authority that signed this assertion
	AuthorityID() string

	// Header retrieves the header with name
	Header(name string) interface{}

	// Headers returns the complete headers
	Headers() map[string]interface{}

	// HeaderString retrieves the string value of header with name or ""
	HeaderString(name string) string

	// Body returns the body of this assertion
	Body() []byte

	// Signature returns the signed content and its unprocessed signature
	Signature() (content, signature []byte)

	// SignKeyID returns the key id for the key that signed this assertion.
	SignKeyID() string

	// Prerequisites returns references to the prerequisite assertions for the validity of this one.
	Prerequisites() []*Ref

	// Ref returns a reference representing this assertion.
	Ref() *Ref
}

// customSigner represents an assertion with special arrangements for its signing key (e.g. self-signed), rather than the usual case where an assertion is signed by its authority.
type customSigner interface {
	// signKey returns the public key material for the key that signed this assertion.  See also SignKeyID.
	signKey() PublicKey
}

// MediaType is the media type for encoded assertions on the wire.
const MediaType = "application/x.ubuntu.assertion"

// assertionBase is the concrete base to hold representation data for actual assertions.
type assertionBase struct {
	headers map[string]interface{}
	body    []byte
	// parsed format iteration
	format int
	// parsed revision
	revision int
	// preserved content
	content []byte
	// unprocessed signature
	signature []byte
}

// HeaderString retrieves the string value of header with name or ""
func (ab *assertionBase) HeaderString(name string) string {
	s, _ := ab.headers[name].(string)
	return s
}

// Type returns the assertion type.
func (ab *assertionBase) Type() *AssertionType {
	return Type(ab.HeaderString("type"))
}

// Format returns the assertion format iteration.
func (ab *assertionBase) Format() int {
	return ab.format
}

// SupportedFormat returns whether the assertion uses a supported
// format iteration. If false the assertion might have been only
// partially parsed.
func (ab *assertionBase) SupportedFormat() bool {
	return ab.format <= maxSupportedFormat[ab.HeaderString("type")]
}

// Revision returns the assertion revision.
func (ab *assertionBase) Revision() int {
	return ab.revision
}

// AuthorityID returns the authority-id a.k.a the signer id of the assertion.
func (ab *assertionBase) AuthorityID() string {
	return ab.HeaderString("authority-id")
}

// Header returns the value of an header by name.
func (ab *assertionBase) Header(name string) interface{} {
	v := ab.headers[name]
	if v == nil {
		return nil
	}
	return copyHeader(v)
}

// Headers returns the complete headers.
func (ab *assertionBase) Headers() map[string]interface{} {
	return copyHeaders(ab.headers)
}

// Body returns the body of the assertion.
func (ab *assertionBase) Body() []byte {
	return ab.body
}

// Signature returns the signed content and its unprocessed signature.
func (ab *assertionBase) Signature() (content, signature []byte) {
	return ab.content, ab.signature
}

// SignKeyID returns the key id for the key that signed this assertion.
func (ab *assertionBase) SignKeyID() string {
	return ab.HeaderString("sign-key-sha3-384")
}

// Prerequisites returns references to the prerequisite assertions for the validity of this one.
func (ab *assertionBase) Prerequisites() []*Ref {
	return nil
}

// Ref returns a reference representing this assertion.
func (ab *assertionBase) Ref() *Ref {
	assertType := ab.Type()
	primKey := make([]string, len(assertType.PrimaryKey))
	for i, name := range assertType.PrimaryKey {
		primKey[i] = ab.HeaderString(name)
	}
	return &Ref{
		Type:       assertType,
		PrimaryKey: primKey,
	}
}

// sanity check
var _ Assertion = (*assertionBase)(nil)

// Decode parses a serialized assertion.
//
// The expected serialisation format looks like:
//
//   HEADER ("\n\n" BODY?)? "\n\n" SIGNATURE
//
// where:
//
//    HEADER is a set of header entries separated by "\n"
//    BODY can be arbitrary text,
//    SIGNATURE is the signature
//
// Both BODY and HEADER must be UTF8.
//
// A header entry for a single line value (no '\n' in it) looks like:
//
//   NAME ": " SIMPLEVALUE
//
// The format supports multiline text values (with '\n's in them) and
// lists or maps, possibly nested, with string scalars in them.
//
// For those a header entry looks like:
//
//   NAME ":\n" MULTI(baseindent)
//
// where MULTI can be
//
// * (baseindent + 4)-space indented value (multiline text)
//
// * entries of a list each of the form:
//
//     " "*baseindent "  -"  ( " " SIMPLEVALUE | "\n" MULTI )
//
// * entries of map each of the form:
//
//     " "*baseindent "  " NAME ":"  ( " " SIMPLEVALUE | "\n" MULTI )
//
// baseindent starts at 0 and then grows with nesting matching the
// previous level introduction (e.g. the " "*baseindent " -" bit)
// length minus 1.
//
// In general the following headers are mandatory:
//
//   type
//   authority-id (except for on the wire/self-signed assertions like serial-request)
//
// Further for a given assertion type all the primary key headers
// must be non empty and must not contain '/'.
//
// The following headers expect string representing integer values and
// if omitted otherwise are assumed to be 0:
//
//   revision (a positive int)
//   body-length (expected to be equal to the length of BODY)
//   format (a positive int for the format iteration of the type used)
//
// Times are expected to be in the RFC3339 format: "2006-01-02T15:04:05Z07:00".
//
func Decode(serializedAssertion []byte) (Assertion, error) {
	// copy to get an independent backstorage that can't be mutated later
	assertionSnapshot := make([]byte, len(serializedAssertion))
	copy(assertionSnapshot, serializedAssertion)
	contentSignatureSplit := bytes.LastIndex(assertionSnapshot, nlnl)
	if contentSignatureSplit == -1 {
		return nil, fmt.Errorf("assertion content/signature separator not found")
	}
	content := assertionSnapshot[:contentSignatureSplit]
	signature := assertionSnapshot[contentSignatureSplit+2:]

	headersBodySplit := bytes.Index(content, nlnl)
	var body, head []byte
	if headersBodySplit == -1 {
		head = content
	} else {
		body = content[headersBodySplit+2:]
		if len(body) == 0 {
			body = nil
		}
		head = content[:headersBodySplit]
	}

	headers, err := parseHeaders(head)
	if err != nil {
		return nil, fmt.Errorf("parsing assertion headers: %v", err)
	}

	return assemble(headers, body, content, signature)
}

// Maximum assertion component sizes.
const (
	MaxBodySize      = 2 * 1024 * 1024
	MaxHeadersSize   = 128 * 1024
	MaxSignatureSize = 128 * 1024
)

// Decoder parses a stream of assertions bundled by separating them with double newlines.
type Decoder struct {
	rd             io.Reader
	initialBufSize int
	b              *bufio.Reader
	err            error
	maxHeadersSize int
	maxSigSize     int

	defaultMaxBodySize int
	typeMaxBodySize    map[*AssertionType]int
}

// initBuffer finishes a Decoder initialization by setting up the bufio.Reader,
// it returns the *Decoder for convenience of notation.
func (d *Decoder) initBuffer() *Decoder {
	d.b = bufio.NewReaderSize(d.rd, d.initialBufSize)
	return d
}

const defaultDecoderBufSize = 4096

// NewDecoder returns a Decoder to parse the stream of assertions from the reader.
func NewDecoder(r io.Reader) *Decoder {
	return (&Decoder{
		rd:                 r,
		initialBufSize:     defaultDecoderBufSize,
		maxHeadersSize:     MaxHeadersSize,
		maxSigSize:         MaxSignatureSize,
		defaultMaxBodySize: MaxBodySize,
	}).initBuffer()
}

// NewDecoderWithTypeMaxBodySize returns a Decoder to parse the stream of assertions from the reader enforcing optional per type max body sizes or the default one as fallback.
func NewDecoderWithTypeMaxBodySize(r io.Reader, typeMaxBodySize map[*AssertionType]int) *Decoder {
	return (&Decoder{
		rd:                 r,
		initialBufSize:     defaultDecoderBufSize,
		maxHeadersSize:     MaxHeadersSize,
		maxSigSize:         MaxSignatureSize,
		defaultMaxBodySize: MaxBodySize,
		typeMaxBodySize:    typeMaxBodySize,
	}).initBuffer()
}

func (d *Decoder) peek(size int) ([]byte, error) {
	buf, err := d.b.Peek(size)
	if err == bufio.ErrBufferFull {
		rebuf, reerr := d.b.Peek(d.b.Buffered())
		if reerr != nil {
			panic(reerr)
		}
		mr := io.MultiReader(bytes.NewBuffer(rebuf), d.rd)
		d.b = bufio.NewReaderSize(mr, (size/d.initialBufSize+1)*d.initialBufSize)
		buf, err = d.b.Peek(size)
	}
	if err != nil && d.err == nil {
		d.err = err
	}
	return buf, d.err
}

// NB: readExact and readUntil use peek underneath and their returned
// buffers are valid only until the next reading call

func (d *Decoder) readExact(size int) ([]byte, error) {
	buf, err := d.peek(size)
	d.b.Discard(len(buf))
	if len(buf) == size {
		return buf, nil
	}
	if err == io.EOF {
		return buf, io.ErrUnexpectedEOF
	}
	return buf, err
}

func (d *Decoder) readUntil(delim []byte, maxSize int) ([]byte, error) {
	last := 0
	size := d.initialBufSize
	for {
		buf, err := d.peek(size)
		if i := bytes.Index(buf[last:], delim); i >= 0 {
			d.b.Discard(last + i + len(delim))
			return buf[:last+i+len(delim)], nil
		}
		// report errors only once we have consumed what is buffered
		if err != nil && len(buf) == d.b.Buffered() {
			d.b.Discard(len(buf))
			return buf, err
		}
		last = size - len(delim) + 1
		size *= 2
		if size > maxSize {
			return nil, fmt.Errorf("maximum size exceeded while looking for delimiter %q", delim)
		}
	}
}

// Decode parses the next assertion from the stream.
// It returns the error io.EOF at the end of a well-formed stream.
func (d *Decoder) Decode() (Assertion, error) {
	// read the headers and the nlnl separator after them
	headAndSep, err := d.readUntil(nlnl, d.maxHeadersSize)
	if err != nil {
		if err == io.EOF {
			if len(headAndSep) != 0 {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, io.EOF
		}
		return nil, fmt.Errorf("error reading assertion headers: %v", err)
	}

	headLen := len(headAndSep) - len(nlnl)
	headers, err := parseHeaders(headAndSep[:headLen])
	if err != nil {
		return nil, fmt.Errorf("parsing assertion headers: %v", err)
	}

	typeStr, _ := headers["type"].(string)
	typ := Type(typeStr)

	length, err := checkIntWithDefault(headers, "body-length", 0)
	if err != nil {
		return nil, fmt.Errorf("assertion: %v", err)
	}
	if typMaxBodySize := d.typeMaxBodySize[typ]; typMaxBodySize != 0 && length > typMaxBodySize {
		return nil, fmt.Errorf("assertion body length %d exceeds maximum body size %d for %q assertions", length, typMaxBodySize, typ.Name)
	} else if length > d.defaultMaxBodySize {
		return nil, fmt.Errorf("assertion body length %d exceeds maximum body size", length)
	}

	// save the headers before we try to read more, and setup to capture
	// the whole content in a buffer
	contentBuf := bytes.NewBuffer(make([]byte, 0, len(headAndSep)+length))
	contentBuf.Write(headAndSep)

	if length > 0 {
		// read the body if length != 0
		body, err := d.readExact(length)
		if err != nil {
			return nil, err
		}
		contentBuf.Write(body)
	}

	// try to read the end of body a.k.a content/signature separator
	endOfBody, err := d.readUntil(nlnl, d.maxSigSize)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading assertion trailer: %v", err)
	}

	var sig []byte
	if bytes.Equal(endOfBody, nlnl) {
		// we got the nlnl content/signature separator, read the signature now and the assertion/assertion nlnl separation
		sig, err = d.readUntil(nlnl, d.maxSigSize)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading assertion signature: %v", err)
		}
	} else {
		// we got the signature directly which is a ok format only if body length == 0
		if length > 0 {
			return nil, fmt.Errorf("missing content/signature separator")
		}
		sig = endOfBody
		contentBuf.Truncate(headLen)
	}

	// normalize sig ending newlines
	if bytes.HasSuffix(sig, nlnl) {
		sig = sig[:len(sig)-1]
	}

	finalContent := contentBuf.Bytes()
	var finalBody []byte
	if length > 0 {
		finalBody = finalContent[headLen+len(nlnl):]
	}

	finalSig := make([]byte, len(sig))
	copy(finalSig, sig)

	return assemble(headers, finalBody, finalContent, finalSig)
}

func checkIteration(headers map[string]interface{}, name string) (int, error) {
	iternum, err := checkIntWithDefault(headers, name, 0)
	if err != nil {
		return -1, err
	}
	if iternum < 0 {
		return -1, fmt.Errorf("%s should be positive: %v", name, iternum)
	}
	return iternum, nil
}

func checkFormat(headers map[string]interface{}) (int, error) {
	return checkIteration(headers, "format")
}

func checkRevision(headers map[string]interface{}) (int, error) {
	return checkIteration(headers, "revision")
}

// Assemble assembles an assertion from its components.
func Assemble(headers map[string]interface{}, body, content, signature []byte) (Assertion, error) {
	err := checkHeaders(headers)
	if err != nil {
		return nil, err
	}
	return assemble(headers, body, content, signature)
}

// assemble is the internal variant of Assemble, assumes headers are already checked for supported types
func assemble(headers map[string]interface{}, body, content, signature []byte) (Assertion, error) {
	length, err := checkIntWithDefault(headers, "body-length", 0)
	if err != nil {
		return nil, fmt.Errorf("assertion: %v", err)
	}
	if length != len(body) {
		return nil, fmt.Errorf("assertion body length and declared body-length don't match: %v != %v", len(body), length)
	}

	if !utf8.Valid(body) {
		return nil, fmt.Errorf("body is not utf8")
	}

	if _, err := checkDigest(headers, "sign-key-sha3-384", crypto.SHA3_384); err != nil {
		return nil, fmt.Errorf("assertion: %v", err)
	}

	typ, err := checkNotEmptyString(headers, "type")
	if err != nil {
		return nil, fmt.Errorf("assertion: %v", err)
	}
	assertType := Type(typ)
	if assertType == nil {
		return nil, fmt.Errorf("unknown assertion type: %q", typ)
	}

	if assertType.flags&noAuthority == 0 {
		if _, err := checkNotEmptyString(headers, "authority-id"); err != nil {
			return nil, fmt.Errorf("assertion: %v", err)
		}
	} else {
		_, ok := headers["authority-id"]
		if ok {
			return nil, fmt.Errorf("%q assertion cannot have authority-id set", assertType.Name)
		}
	}

	formatnum, err := checkFormat(headers)
	if err != nil {
		return nil, fmt.Errorf("assertion: %v", err)
	}

	for _, primKey := range assertType.PrimaryKey {
		if _, err := checkPrimaryKey(headers, primKey); err != nil {
			return nil, fmt.Errorf("assertion %s: %v", assertType.Name, err)
		}
	}

	revision, err := checkRevision(headers)
	if err != nil {
		return nil, fmt.Errorf("assertion: %v", err)
	}

	if len(signature) == 0 {
		return nil, fmt.Errorf("empty assertion signature")
	}

	assert, err := assertType.assembler(assertionBase{
		headers:   headers,
		body:      body,
		format:    formatnum,
		revision:  revision,
		content:   content,
		signature: signature,
	})
	if err != nil {
		return nil, fmt.Errorf("assertion %s: %v", assertType.Name, err)
	}
	return assert, nil
}

func writeHeader(buf *bytes.Buffer, headers map[string]interface{}, name string) {
	appendEntry(buf, fmt.Sprintf("%s:", name), headers[name], 0)
}

func assembleAndSign(assertType *AssertionType, headers map[string]interface{}, body []byte, privKey PrivateKey) (Assertion, error) {
	err := checkAssertType(assertType)
	if err != nil {
		return nil, err
	}

	withAuthority := assertType.flags&noAuthority == 0

	err = checkHeaders(headers)
	if err != nil {
		return nil, err
	}

	// there's no hint at all that we will need non-textual bodies,
	// make sure we actually enforce that
	if !utf8.Valid(body) {
		return nil, fmt.Errorf("assertion body is not utf8")
	}

	finalHeaders := copyHeaders(headers)
	bodyLength := len(body)
	finalBody := make([]byte, bodyLength)
	copy(finalBody, body)
	finalHeaders["type"] = assertType.Name
	finalHeaders["body-length"] = strconv.Itoa(bodyLength)
	finalHeaders["sign-key-sha3-384"] = privKey.PublicKey().ID()

	if withAuthority {
		if _, err := checkNotEmptyString(finalHeaders, "authority-id"); err != nil {
			return nil, err
		}
	} else {
		_, ok := finalHeaders["authority-id"]
		if ok {
			return nil, fmt.Errorf("%q assertion cannot have authority-id set", assertType.Name)
		}
	}

	formatnum, err := checkFormat(finalHeaders)
	if err != nil {
		return nil, err
	}

	if formatnum > assertType.MaxSupportedFormat() {
		return nil, fmt.Errorf("cannot sign %q assertion with format %d higher than max supported format %d", assertType.Name, formatnum, assertType.MaxSupportedFormat())
	}

	suggestedFormat, err := SuggestFormat(assertType, finalHeaders, finalBody)
	if err != nil {
		return nil, err
	}

	if suggestedFormat > formatnum {
		return nil, fmt.Errorf("cannot sign %q assertion with format set to %d lower than min format %d covering included features", assertType.Name, formatnum, suggestedFormat)
	}

	revision, err := checkRevision(finalHeaders)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString("type: ")
	buf.WriteString(assertType.Name)

	if formatnum > 0 {
		writeHeader(buf, finalHeaders, "format")
	} else {
		delete(finalHeaders, "format")
	}

	if withAuthority {
		writeHeader(buf, finalHeaders, "authority-id")
	}

	if revision > 0 {
		writeHeader(buf, finalHeaders, "revision")
	} else {
		delete(finalHeaders, "revision")
	}
	written := map[string]bool{
		"type":              true,
		"format":            true,
		"authority-id":      true,
		"revision":          true,
		"body-length":       true,
		"sign-key-sha3-384": true,
	}
	for _, primKey := range assertType.PrimaryKey {
		if _, err := checkPrimaryKey(finalHeaders, primKey); err != nil {
			return nil, err
		}
		writeHeader(buf, finalHeaders, primKey)
		written[primKey] = true
	}

	// emit other headers in lexicographic order
	otherKeys := make([]string, 0, len(finalHeaders))
	for name := range finalHeaders {
		if !written[name] {
			otherKeys = append(otherKeys, name)
		}
	}
	sort.Strings(otherKeys)
	for _, k := range otherKeys {
		writeHeader(buf, finalHeaders, k)
	}

	// body-length and body
	if bodyLength > 0 {
		writeHeader(buf, finalHeaders, "body-length")
	} else {
		delete(finalHeaders, "body-length")
	}

	// signing key reference
	writeHeader(buf, finalHeaders, "sign-key-sha3-384")

	if bodyLength > 0 {
		buf.Grow(bodyLength + 2)
		buf.Write(nlnl)
		buf.Write(finalBody)
	} else {
		finalBody = nil
	}
	content := buf.Bytes()

	signature, err := signContent(content, privKey)
	if err != nil {
		return nil, fmt.Errorf("cannot sign assertion: %v", err)
	}
	// be 'cat' friendly, add a ignored newline to the signature which is the last part of the encoded assertion
	signature = append(signature, '\n')

	assert, err := assertType.assembler(assertionBase{
		headers:   finalHeaders,
		body:      finalBody,
		format:    formatnum,
		revision:  revision,
		content:   content,
		signature: signature,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot assemble assertion %s: %v", assertType.Name, err)
	}
	return assert, nil
}

// SignWithoutAuthority assembles an assertion without a set authority with the provided information and signs it with the given private key.
func SignWithoutAuthority(assertType *AssertionType, headers map[string]interface{}, body []byte, privKey PrivateKey) (Assertion, error) {
	if assertType.flags&noAuthority == 0 {
		return nil, fmt.Errorf("cannot sign assertions needing a definite authority with SignWithoutAuthority")
	}
	return assembleAndSign(assertType, headers, body, privKey)
}

// Encode serializes an assertion.
func Encode(assert Assertion) []byte {
	content, signature := assert.Signature()
	needed := len(content) + 2 + len(signature)
	buf := bytes.NewBuffer(make([]byte, 0, needed))
	buf.Write(content)
	buf.Write(nlnl)
	buf.Write(signature)
	return buf.Bytes()
}

// Encoder emits a stream of assertions bundled by separating them with double newlines.
type Encoder struct {
	wr      io.Writer
	nextSep []byte
}

// NewEncoder returns a Encoder to emit a stream of assertions to a writer.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{wr: w}
}

func (enc *Encoder) writeSep(last byte) error {
	if last != '\n' {
		_, err := enc.wr.Write(nl)
		if err != nil {
			return err
		}
	}
	enc.nextSep = nl
	return nil
}

// WriteEncoded writes the encoded assertion into the stream with the required separator.
func (enc *Encoder) WriteEncoded(encoded []byte) error {
	sz := len(encoded)
	if sz == 0 {
		return fmt.Errorf("internal error: encoded assertion cannot be empty")
	}

	_, err := enc.wr.Write(enc.nextSep)
	if err != nil {
		return err
	}

	_, err = enc.wr.Write(encoded)
	if err != nil {
		return err
	}

	return enc.writeSep(encoded[sz-1])
}

// WriteContentSignature writes the content and signature of an assertion into the stream with all the required separators.
func (enc *Encoder) WriteContentSignature(content, signature []byte) error {
	if len(content) == 0 {
		return fmt.Errorf("internal error: content cannot be empty")
	}

	sz := len(signature)
	if sz == 0 {
		return fmt.Errorf("internal error: signature cannot be empty")
	}

	_, err := enc.wr.Write(enc.nextSep)
	if err != nil {
		return err
	}

	_, err = enc.wr.Write(content)
	if err != nil {
		return err
	}
	_, err = enc.wr.Write(nlnl)
	if err != nil {
		return err
	}
	_, err = enc.wr.Write(signature)
	if err != nil {
		return err
	}

	return enc.writeSep(signature[sz-1])
}

// Encode emits the assertion into the stream with the required separator.
// Errors here are always about writing given that Encode() itself cannot error.
func (enc *Encoder) Encode(assert Assertion) error {
	return enc.WriteContentSignature(assert.Signature())
}

// SignatureCheck checks the signature of the assertion against the given public key. Useful for assertions with no authority.
func SignatureCheck(assert Assertion, pubKey PublicKey) error {
	content, encodedSig := assert.Signature()
	sig, err := decodeSignature(encodedSig)
	if err != nil {
		return err
	}
	err = pubKey.verify(content, sig)
	if err != nil {
		return fmt.Errorf("failed signature verification: %v", err)
	}
	return nil
}
