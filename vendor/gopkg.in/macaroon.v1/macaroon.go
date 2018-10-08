// The macaroon package implements macaroons as described in
// the paper "Macaroons: Cookies with Contextual Caveats for
// Decentralized Authorization in the Cloud"
// (http://theory.stanford.edu/~ataly/Papers/macaroons.pdf)
//
// See the macaroon bakery packages at http://godoc.org/gopkg.in/macaroon-bakery.v0
// for higher level services and operations that use macaroons.
package macaroon // import "gopkg.in/macaroon.v1"

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"fmt"
	"io"
)

// Macaroon holds a macaroon.
// See Fig. 7 of http://theory.stanford.edu/~ataly/Papers/macaroons.pdf
// for a description of the data contained within.
// Macaroons are mutable objects - use Clone as appropriate
// to avoid unwanted mutation.
type Macaroon struct {
	// data holds the binary-marshalled form
	// of the macaroon data.
	data []byte

	location packet
	id       packet
	caveats  []caveat
	sig      [hashLen]byte
}

// caveat holds a first person or third party caveat.
type caveat struct {
	location       packet
	caveatId       packet
	verificationId packet
}

type Caveat struct {
	Id       string
	Location string
}

// isThirdParty reports whether the caveat must be satisfied
// by some third party (if not, it's a first person caveat).
func (cav *caveat) isThirdParty() bool {
	return cav.verificationId.len() > 0
}

// New returns a new macaroon with the given root key,
// identifier and location.
func New(rootKey []byte, id, loc string) (*Macaroon, error) {
	var m Macaroon
	if err := m.init(id, loc); err != nil {
		return nil, err
	}
	derivedKey := makeKey(rootKey)
	m.sig = *keyedHash(derivedKey, m.dataBytes(m.id))
	return &m, nil
}

func (m *Macaroon) init(id, loc string) error {
	var ok bool
	m.location, ok = m.appendPacket(fieldLocation, []byte(loc))
	if !ok {
		return fmt.Errorf("macaroon location too big")
	}
	m.id, ok = m.appendPacket(fieldIdentifier, []byte(id))
	if !ok {
		return fmt.Errorf("macaroon identifier too big")
	}
	return nil
}

// Clone returns a copy of the receiving macaroon.
func (m *Macaroon) Clone() *Macaroon {
	m1 := *m
	// Ensure that if any data is appended to the new
	// macaroon, it will copy data and caveats.
	m1.data = m1.data[0:len(m1.data):len(m1.data)]
	m1.caveats = m1.caveats[0:len(m1.caveats):len(m1.caveats)]
	return &m1
}

// Location returns the macaroon's location hint. This is
// not verified as part of the macaroon.
func (m *Macaroon) Location() string {
	return m.dataStr(m.location)
}

// Id returns the id of the macaroon. This can hold
// arbitrary information.
func (m *Macaroon) Id() string {
	return m.dataStr(m.id)
}

// Signature returns the macaroon's signature.
func (m *Macaroon) Signature() []byte {
	// sig := m.sig
	// return sig[:]
	// Work around https://github.com/golang/go/issues/9537
	sig := new([hashLen]byte)
	*sig = m.sig
	return sig[:]
}

// Caveats returns the macaroon's caveats.
// This method will probably change, and it's important not to change the returned caveat.
func (m *Macaroon) Caveats() []Caveat {
	caveats := make([]Caveat, len(m.caveats))
	for i, cav := range m.caveats {
		caveats[i] = Caveat{
			Id:       m.dataStr(cav.caveatId),
			Location: m.dataStr(cav.location),
		}
	}
	return caveats
}

// appendCaveat appends a caveat without modifying the macaroon's signature.
func (m *Macaroon) appendCaveat(caveatId string, verificationId []byte, loc string) (*caveat, error) {
	var cav caveat
	var ok bool
	if caveatId != "" {
		cav.caveatId, ok = m.appendPacket(fieldCaveatId, []byte(caveatId))
		if !ok {
			return nil, fmt.Errorf("caveat identifier too big")
		}
	}
	if len(verificationId) > 0 {
		cav.verificationId, ok = m.appendPacket(fieldVerificationId, verificationId)
		if !ok {
			return nil, fmt.Errorf("caveat verification id too big")
		}
	}
	if loc != "" {
		cav.location, ok = m.appendPacket(fieldCaveatLocation, []byte(loc))
		if !ok {
			return nil, fmt.Errorf("caveat location too big")
		}
	}
	m.caveats = append(m.caveats, cav)
	return &m.caveats[len(m.caveats)-1], nil
}

func (m *Macaroon) addCaveat(caveatId string, verificationId []byte, loc string) error {
	cav, err := m.appendCaveat(caveatId, verificationId, loc)
	if err != nil {
		return err
	}
	m.sig = *keyedHash2(&m.sig, m.dataBytes(cav.verificationId), m.dataBytes(cav.caveatId))
	return nil
}

func keyedHash2(key *[keyLen]byte, d1, d2 []byte) *[hashLen]byte {
	if len(d1) == 0 {
		return keyedHash(key, d2)
	}
	var data [hashLen * 2]byte
	copy(data[0:], keyedHash(key, d1)[:])
	copy(data[hashLen:], keyedHash(key, d2)[:])
	return keyedHash(key, data[:])
}

// Bind prepares the macaroon for being used to discharge the
// macaroon with the given signature sig. This must be
// used before it is used in the discharges argument to Verify.
func (m *Macaroon) Bind(sig []byte) {
	m.sig = *bindForRequest(sig, &m.sig)
}

// AddFirstPartyCaveat adds a caveat that will be verified
// by the target service.
func (m *Macaroon) AddFirstPartyCaveat(caveatId string) error {
	return m.addCaveat(caveatId, nil, "")
}

// AddThirdPartyCaveat adds a third-party caveat to the macaroon,
// using the given shared root key, caveat id and location hint.
// The caveat id should encode the root key in some
// way, either by encrypting it with a key known to the third party
// or by holding a reference to it stored in the third party's
// storage.
func (m *Macaroon) AddThirdPartyCaveat(rootKey []byte, caveatId string, loc string) error {
	return m.addThirdPartyCaveatWithRand(rootKey, caveatId, loc, rand.Reader)
}

func (m *Macaroon) addThirdPartyCaveatWithRand(rootKey []byte, caveatId string, loc string, r io.Reader) error {
	derivedKey := makeKey(rootKey)
	verificationId, err := encrypt(&m.sig, derivedKey, r)
	if err != nil {
		return err
	}
	return m.addCaveat(caveatId, verificationId, loc)
}

var zeroKey [hashLen]byte

// bindForRequest binds the given macaroon
// to the given signature of its parent macaroon.
func bindForRequest(rootSig []byte, dischargeSig *[hashLen]byte) *[hashLen]byte {
	if bytes.Equal(rootSig, dischargeSig[:]) {
		return dischargeSig
	}
	return keyedHash2(&zeroKey, rootSig, dischargeSig[:])
}

// Verify verifies that the receiving macaroon is valid.
// The root key must be the same that the macaroon was originally
// minted with. The check function is called to verify each
// first-party caveat - it should return an error if the
// condition is not met.
//
// The discharge macaroons should be provided in discharges.
//
// Verify returns nil if the verification succeeds.
func (m *Macaroon) Verify(rootKey []byte, check func(caveat string) error, discharges []*Macaroon) error {
	derivedKey := makeKey(rootKey)
	// TODO(rog) consider distinguishing between classes of
	// check error - some errors may be resolved by minting
	// a new macaroon; others may not.
	used := make([]int, len(discharges))
	if err := m.verify(&m.sig, derivedKey, check, discharges, used); err != nil {
		return err
	}
	for i, dm := range discharges {
		switch used[i] {
		case 0:
			return fmt.Errorf("discharge macaroon %q was not used", dm.Id())
		case 1:
			continue
		default:
			// Should be impossible because of check in verify, but be defensive.
			return fmt.Errorf("discharge macaroon %q was used more than once", dm.Id())
		}
	}
	return nil
}

func (m *Macaroon) verify(rootSig *[hashLen]byte, rootKey *[hashLen]byte, check func(caveat string) error, discharges []*Macaroon, used []int) error {
	caveatSig := keyedHash(rootKey, m.dataBytes(m.id))
	for i, cav := range m.caveats {
		if cav.isThirdParty() {
			cavKey, err := decrypt(caveatSig, m.dataBytes(cav.verificationId))
			if err != nil {
				return fmt.Errorf("failed to decrypt caveat %d signature: %v", i, err)
			}
			// We choose an arbitrary error from one of the
			// possible discharge macaroon verifications
			// if there's more than one discharge macaroon
			// with the required id.
			found := false
			for di, dm := range discharges {
				if !bytes.Equal(dm.dataBytes(dm.id), m.dataBytes(cav.caveatId)) {
					continue
				}
				found = true

				// It's important that we do this before calling verify,
				// as it prevents potentially infinite recursion.
				if used[di]++; used[di] > 1 {
					return fmt.Errorf("discharge macaroon %q was used more than once", dm.Id())
				}
				if err := dm.verify(rootSig, cavKey, check, discharges, used); err != nil {
					return err
				}
				break
			}
			if !found {
				return fmt.Errorf("cannot find discharge macaroon for caveat %q", m.dataBytes(cav.caveatId))
			}
		} else {
			if err := check(string(m.dataBytes(cav.caveatId))); err != nil {
				return err
			}
		}
		caveatSig = keyedHash2(caveatSig, m.dataBytes(cav.verificationId), m.dataBytes(cav.caveatId))
	}
	// TODO perhaps we should actually do this check before doing
	// all the potentially expensive caveat checks.
	boundSig := bindForRequest(rootSig[:], caveatSig)
	if !hmac.Equal(boundSig[:], m.sig[:]) {
		return fmt.Errorf("signature mismatch after caveat verification")
	}
	return nil
}

type Verifier interface {
	Verify(m *Macaroon, rootKey []byte) (bool, error)
}
