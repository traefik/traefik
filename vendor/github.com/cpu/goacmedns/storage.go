package goacmedns

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// Storage is an interface describing the required functions for an ACME DNS
// Account storage mechanism.
type Storage interface {
	// Save will persist the `Account` data that has been `Put` so far
	Save() error
	// Put will add an `Account` for the given domain to the storage. It may not
	// be persisted until `Save` is called.
	Put(string, Account) error
	// Fetch will retrieve an `Account` for the given domain from the storage. If
	// the provided domain does not have an `Account` saved in the storage
	// `ErrDomainNotFound` will be returned
	Fetch(string) (Account, error)
}

var (
	// ErrDomainNotFound is returned from `Fetch` when the provided domain is not
	// present in the storage.
	ErrDomainNotFound = errors.New("requested domain is not present in storage")
)

// fileStorage implements the `Storage` interface and persists `Accounts` to
// a JSON file on disk.
type fileStorage struct {
	// path is the filepath that the `accounts` are persisted to when the `Save`
	// function is called.
	path string
	// mode is the file mode used when the `path` JSON file must be created
	mode os.FileMode
	// accounts holds the `Account` data that has been `Put` into the storage
	accounts map[string]Account
}

// NewFileStorage returns a `Storage` implementation backed by JSON content
// saved into the provided `path` on disk. The file at `path` will be created if
// required. When creating a new file the provided `mode` is used to set the
// permissions.
func NewFileStorage(path string, mode os.FileMode) Storage {
	fs := fileStorage{
		path:     path,
		mode:     mode,
		accounts: make(map[string]Account),
	}
	// Opportunistically try to load the account data. Return an empty account if
	// any errors occur.
	if jsonData, err := ioutil.ReadFile(path); err == nil {
		if err := json.Unmarshal(jsonData, &fs.accounts); err != nil {
			return fs
		}
	}
	return fs
}

// Save persists the `Account` data to the fileStorage's configured path. The
// file at that path will be created with the fileStorage's mode if required.
func (f fileStorage) Save() error {
	if serialized, err := json.Marshal(f.accounts); err != nil {
		return err
	} else if err = ioutil.WriteFile(f.path, serialized, f.mode); err != nil {
		return err
	}
	return nil
}

// Put saves an `Account` for the given `Domain` into the in-memory accounts of
// the fileStorage instance. The `Account` data will not be written to disk
// until the `Save` function is called
func (f fileStorage) Put(domain string, acct Account) error {
	f.accounts[domain] = acct
	return nil
}

// Fetch retrieves the `Account` object for the given `domain` from the
// fileStorage in-memory accounts. If the `domain` provided does not have an
// `Account` in the storage an `ErrDomainNotFound` error is returned.
func (f fileStorage) Fetch(domain string) (Account, error) {
	if acct, exists := f.accounts[domain]; exists {
		return acct, nil
	}
	return Account{}, ErrDomainNotFound
}
