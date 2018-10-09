// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
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
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/osutil"
)

func ensureGPGHomeDirectory() (string, error) {
	real, err := osutil.RealUser()
	if err != nil {
		return "", err
	}

	uid, gid, err := osutil.UidGid(real)
	if err != nil {
		return "", err
	}

	homedir := os.Getenv("SNAP_GNUPG_HOME")
	if homedir == "" {
		homedir = filepath.Join(real.HomeDir, ".snap", "gnupg")
	}

	if err := osutil.MkdirAllChown(homedir, 0700, uid, gid); err != nil {
		return "", err
	}
	return homedir, nil
}

// findGPGCommand returns the path to a suitable GnuPG binary to use.
// GnuPG 2 is mainly intended for desktop use, and is hard for us to use
// here: in particular, it's extremely difficult to use it to delete a
// secret key without a pinentry prompt (which would be necessary in our
// test suite).  GnuPG 1 is still supported so it's reasonable to continue
// using that for now.
func findGPGCommand() (string, error) {
	if path := os.Getenv("SNAP_GNUPG_CMD"); path != "" {
		return path, nil
	}

	path, err := exec.LookPath("gpg1")
	if err != nil {
		path, err = exec.LookPath("gpg")
	}
	return path, err
}

func runGPGImpl(input []byte, args ...string) ([]byte, error) {
	homedir, err := ensureGPGHomeDirectory()
	if err != nil {
		return nil, err
	}

	// Ensure the gpg-agent knows what tty to talk to to ask for
	// the passphrase. This is needed because we drive gpg over
	// a pipe and if the agent is not already started it will
	// fail to be able to ask for a password.
	if os.Getenv("GPG_TTY") == "" {
		tty, err := os.Readlink("/proc/self/fd/0")
		if err != nil {
			return nil, err
		}
		os.Setenv("GPG_TTY", tty)
	}

	general := []string{"--homedir", homedir, "-q", "--no-auto-check-trustdb"}
	allArgs := append(general, args...)

	path, err := findGPGCommand()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(path, allArgs...)
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	if len(input) != 0 {
		cmd.Stdin = bytes.NewBuffer(input)
	}

	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s failed: %v (%q)", path, strings.Join(args, " "), err, errBuf.Bytes())
	}

	return outBuf.Bytes(), nil
}

var runGPG = runGPGImpl

// A key pair manager backed by a local GnuPG setup.
type GPGKeypairManager struct{}

func (gkm *GPGKeypairManager) gpg(input []byte, args ...string) ([]byte, error) {
	return runGPG(input, args...)
}

// NewGPGKeypairManager creates a new key pair manager backed by a local GnuPG setup.
// Importing keys through the keypair manager interface is not
// suppored.
// Main purpose is allowing signing using keys from a GPG setup.
func NewGPGKeypairManager() *GPGKeypairManager {
	return &GPGKeypairManager{}
}

func (gkm *GPGKeypairManager) retrieve(fpr string) (PrivateKey, error) {
	out, err := gkm.gpg(nil, "--batch", "--export", "--export-options", "export-minimal,export-clean,no-export-attributes", "0x"+fpr)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("cannot retrieve key with fingerprint %q in GPG keyring", fpr)
	}

	pubKeyBuf := bytes.NewBuffer(out)
	privKey, err := newExtPGPPrivateKey(pubKeyBuf, "GPG", func(content []byte) ([]byte, error) {
		return gkm.sign(fpr, content)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot load GPG public key with fingerprint %q: %v", fpr, err)
	}
	gotFingerprint := privKey.fingerprint()
	if gotFingerprint != fpr {
		return nil, fmt.Errorf("got wrong public key from GPG, expected fingerprint %q: %s", fpr, gotFingerprint)
	}
	return privKey, nil
}

// Walk iterates over all the RSA private keys in the local GPG setup calling the provided callback until this returns an error
func (gkm *GPGKeypairManager) Walk(consider func(privk PrivateKey, fingerprint string, uid string) error) error {
	// see GPG source doc/DETAILS
	out, err := gkm.gpg(nil, "--batch", "--list-secret-keys", "--fingerprint", "--with-colons", "--fixed-list-mode")
	if err != nil {
		return err
	}
	lines := strings.Split(string(out), "\n")
	n := len(lines)
	if n > 0 && lines[n-1] == "" {
		n--
	}
	if n == 0 {
		return nil
	}
	lines = lines[:n]
	for j := 0; j < n; j++ {
		// sec: line
		line := lines[j]
		if !strings.HasPrefix(line, "sec:") {
			continue
		}
		secFields := strings.Split(line, ":")
		if len(secFields) < 5 {
			continue
		}
		if secFields[3] != "1" { // not RSA
			continue
		}
		keyID := secFields[4]
		uid := ""
		fpr := ""
		var privKey PrivateKey
		// look for fpr:, uid: lines, order may vary and gpg2.1
		// may springle additional lines in (like gpr:)
	Loop:
		for k := j + 1; k < n && !strings.HasPrefix(lines[k], "sec:"); k++ {
			switch {
			case strings.HasPrefix(lines[k], "fpr:"):
				fprFields := strings.Split(lines[k], ":")
				// extract "Field 10 - User-ID"
				// A FPR record stores the fingerprint here.
				if len(fprFields) < 10 {
					break Loop
				}
				fpr = fprFields[9]
				if !strings.HasSuffix(fpr, keyID) {
					break // strange, skip
				}
				privKey, err = gkm.retrieve(fpr)
				if err != nil {
					return err
				}
			case strings.HasPrefix(lines[k], "uid:"):
				uidFields := strings.Split(lines[k], ":")
				// extract "*** Field 10 - User-ID"
				if len(uidFields) < 10 {
					break Loop
				}
				uid = uidFields[9]
			}
		}
		// sanity checking
		if privKey == nil || uid == "" {
			continue
		}
		// collected it all
		err = consider(privKey, fpr, uid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gkm *GPGKeypairManager) Put(privKey PrivateKey) error {
	// NOTE: we don't need this initially at least and this keypair mgr is not for general arbitrary usage
	return fmt.Errorf("cannot import private key into GPG keyring")
}

func (gkm *GPGKeypairManager) Get(keyID string) (PrivateKey, error) {
	stop := errors.New("stop marker")
	var hit PrivateKey
	match := func(privk PrivateKey, fpr string, uid string) error {
		if privk.PublicKey().ID() == keyID {
			hit = privk
			return stop
		}
		return nil
	}
	err := gkm.Walk(match)
	if err == stop {
		return hit, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("cannot find key %q in GPG keyring", keyID)
}

func (gkm *GPGKeypairManager) sign(fingerprint string, content []byte) ([]byte, error) {
	out, err := gkm.gpg(content, "--personal-digest-preferences", "SHA512", "--default-key", "0x"+fingerprint, "--detach-sign")
	if err != nil {
		return nil, fmt.Errorf("cannot sign using GPG: %v", err)
	}
	return out, nil
}

type gpgKeypairInfo struct {
	privKey     PrivateKey
	fingerprint string
}

func (gkm *GPGKeypairManager) findByName(name string) (*gpgKeypairInfo, error) {
	stop := errors.New("stop marker")
	var hit *gpgKeypairInfo
	match := func(privk PrivateKey, fpr string, uid string) error {
		if uid == name {
			hit = &gpgKeypairInfo{
				privKey:     privk,
				fingerprint: fpr,
			}
			return stop
		}
		return nil
	}
	err := gkm.Walk(match)
	if err == stop {
		return hit, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("cannot find key named %q in GPG keyring", name)
}

// GetByName looks up a private key by name and returns it.
func (gkm *GPGKeypairManager) GetByName(name string) (PrivateKey, error) {
	keyInfo, err := gkm.findByName(name)
	if err != nil {
		return nil, err
	}
	return keyInfo.privKey, nil
}

var generateTemplate = `
Key-Type: RSA
Key-Length: 4096
Name-Real: %s
Creation-Date: seconds=%d
Preferences: SHA512
`

func (gkm *GPGKeypairManager) parametersForGenerate(passphrase string, name string) string {
	fixedCreationTime := v1FixedTimestamp.Unix()
	generateParams := fmt.Sprintf(generateTemplate, name, fixedCreationTime)
	if passphrase != "" {
		generateParams += "Passphrase: " + passphrase + "\n"
	}
	return generateParams
}

// Generate creates a new key with the given passphrase and name.
func (gkm *GPGKeypairManager) Generate(passphrase string, name string) error {
	_, err := gkm.findByName(name)
	if err == nil {
		return fmt.Errorf("key named %q already exists in GPG keyring", name)
	}
	generateParams := gkm.parametersForGenerate(passphrase, name)
	_, err = gkm.gpg([]byte(generateParams), "--batch", "--gen-key")
	if err != nil {
		return err
	}
	return nil
}

// Export returns the encoded text of the named public key.
func (gkm *GPGKeypairManager) Export(name string) ([]byte, error) {
	keyInfo, err := gkm.findByName(name)
	if err != nil {
		return nil, err
	}
	return EncodePublicKey(keyInfo.privKey.PublicKey())
}

// Delete removes the named key pair from GnuPG's storage.
func (gkm *GPGKeypairManager) Delete(name string) error {
	keyInfo, err := gkm.findByName(name)
	if err != nil {
		return err
	}
	_, err = gkm.gpg(nil, "--batch", "--delete-secret-and-public-key", "0x"+keyInfo.fingerprint)
	if err != nil {
		return err
	}
	return nil
}
