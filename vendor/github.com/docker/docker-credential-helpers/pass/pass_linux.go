// A `pass` based credential helper. Passwords are stored as arguments to pass
// of the form: "$PASS_FOLDER/base64-url(serverURL)/username". We base64-url
// encode the serverURL, because under the hood pass uses files and folders, so
// /s will get translated into additional folders.
package pass

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
)

const PASS_FOLDER = "docker-credential-helpers"

var (
	PassInitialized bool
)

func init() {
	// In principle, we could just run `pass init`. However, pass has a bug
	// where if gpg fails, it doesn't always exit 1. Additionally, pass
	// uses gpg2, but gpg is the default, which may be confusing. So let's
	// just explictily check that pass actually can store and retreive a
	// password.
	password := "pass is initialized"
	name := path.Join(PASS_FOLDER, "docker-pass-initialized-check")

	_, err := runPass(password, "insert", "-f", "-m", name)
	if err != nil {
		return
	}

	stored, err := runPass("", "show", name)
	PassInitialized = err == nil && stored == password

	if PassInitialized {
		runPass("", "rm", "-rf", name)
	}
}

func runPass(stdinContent string, args ...string) (string, error) {
	cmd := exec.Command("pass", args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	defer stdin.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	defer stderr.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer stdout.Close()

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	_, err = stdin.Write([]byte(stdinContent))
	if err != nil {
		return "", err
	}
	stdin.Close()

	errContent, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", fmt.Errorf("error reading stderr: %s", err)
	}

	result, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", fmt.Errorf("Error reading stdout: %s", err)
	}

	cmdErr := cmd.Wait()
	if cmdErr != nil {
		return "", fmt.Errorf("%s: %s", cmdErr, errContent)
	}

	return string(result), nil
}

// Pass handles secrets using Linux secret-service as a store.
type Pass struct{}

// Add adds new credentials to the keychain.
func (h Pass) Add(creds *credentials.Credentials) error {
	if !PassInitialized {
		return errors.New("pass store is uninitialized")
	}

	if creds == nil {
		return errors.New("missing credentials")
	}

	encoded := base64.URLEncoding.EncodeToString([]byte(creds.ServerURL))

	_, err := runPass(creds.Secret, "insert", "-f", "-m", path.Join(PASS_FOLDER, encoded, creds.Username))
	return err
}

// Delete removes credentials from the store.
func (h Pass) Delete(serverURL string) error {
	if !PassInitialized {
		return errors.New("pass store is uninitialized")
	}

	if serverURL == "" {
		return errors.New("missing server url")
	}

	encoded := base64.URLEncoding.EncodeToString([]byte(serverURL))
	_, err := runPass("", "rm", "-rf", path.Join(PASS_FOLDER, encoded))
	return err
}

func getPassDir() string {
	passDir := os.ExpandEnv("$HOME/.password-store")
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) < 2 {
			continue
		}

		if parts[0] != "PASSWORD_STORE_DIR" {
			continue
		}

		passDir = parts[1]
		break
	}

	return passDir
}

// listPassDir lists all the contents of a directory in the password store.
// Pass uses fancy unicode to emit stuff to stdout, so rather than try
// and parse this, let's just look at the directory structure instead.
func listPassDir(args ...string) ([]os.FileInfo, error) {
	passDir := getPassDir()
	p := path.Join(append([]string{passDir, PASS_FOLDER}, args...)...)
	contents, err := ioutil.ReadDir(p)
	if err != nil {
		if os.IsNotExist(err) {
			return []os.FileInfo{}, nil
		}

		return nil, err
	}

	return contents, nil
}

// Get returns the username and secret to use for a given registry server URL.
func (h Pass) Get(serverURL string) (string, string, error) {
	if !PassInitialized {
		return "", "", errors.New("pass store is uninitialized")
	}

	if serverURL == "" {
		return "", "", errors.New("missing server url")
	}

	encoded := base64.URLEncoding.EncodeToString([]byte(serverURL))

	if _, err := os.Stat(path.Join(getPassDir(), PASS_FOLDER, encoded)); err != nil {
		if os.IsNotExist(err) {
			return "", "", nil;
		}

		return "", "", err
	}

	usernames, err := listPassDir(encoded)
	if err != nil {
		return "", "", err
	}

	if len(usernames) < 1 {
		return "", "", fmt.Errorf("no usernames for %s", serverURL)
	}

	actual := strings.TrimSuffix(usernames[0].Name(), ".gpg")
	secret, err := runPass("", "show", path.Join(PASS_FOLDER, encoded, actual))
	return actual, secret, err
}

// List returns the stored URLs and corresponding usernames for a given credentials label
func (h Pass) List() (map[string]string, error) {
	if !PassInitialized {
		return nil, errors.New("pass store is uninitialized")
	}

	servers, err := listPassDir()
	if err != nil {
		return nil, err
	}

	resp := map[string]string{}

	for _, server := range servers {
		if !server.IsDir() {
			continue
		}

		serverURL, err := base64.URLEncoding.DecodeString(server.Name())
		if err != nil {
			return nil, err
		}

		usernames, err := listPassDir(server.Name())
		if err != nil {
			return nil, err
		}

		if len(usernames) < 1 {
			return nil, fmt.Errorf("no usernames for %s", serverURL)
		}

		resp[string(serverURL)] = strings.TrimSuffix(usernames[0].Name(), ".gpg")
	}

	return resp, nil
}
