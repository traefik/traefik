package storeconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"

	"github.com/abronan/valkeyrie/store"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/cmd"
	"github.com/containous/traefik/log"
)

// NewCmd builds a new StoreConfig command
func NewCmd(traefikConfiguration *cmd.TraefikConfiguration, traefikPointersConfiguration *cmd.TraefikConfiguration) *flaeg.Command {
	return &flaeg.Command{
		Name:                  "storeconfig",
		Description:           `Store the static traefik configuration into a Key-value stores. Traefik will not start.`,
		Config:                traefikConfiguration,
		DefaultPointersConfig: traefikPointersConfiguration,
		Metadata: map[string]string{
			"parseAllSources": "true",
		},
	}
}

// Run store config in KV
func Run(kv *staert.KvSource, traefikConfiguration *cmd.TraefikConfiguration) func() error {
	return func() error {
		if kv == nil {
			return fmt.Errorf("error using command storeconfig, no Key-value store defined")
		}

		fileConfig := traefikConfiguration.GlobalConfiguration.File
		if fileConfig != nil {
			traefikConfiguration.GlobalConfiguration.File = nil
			if len(fileConfig.Filename) == 0 && len(fileConfig.Directory) == 0 {
				fileConfig.Filename = traefikConfiguration.ConfigFile
			}
		}

		jsonConf, err := json.Marshal(traefikConfiguration.GlobalConfiguration)
		if err != nil {
			return err
		}
		stdlog.Printf("Storing configuration: %s\n", jsonConf)

		err = kv.StoreConfig(traefikConfiguration.GlobalConfiguration)
		if err != nil {
			return err
		}

		if fileConfig != nil {
			jsonConf, err = json.Marshal(fileConfig)
			if err != nil {
				return err
			}

			stdlog.Printf("Storing file configuration: %s\n", jsonConf)
			config, err := fileConfig.BuildConfiguration()
			if err != nil {
				return err
			}

			stdlog.Print("Writing config to KV")
			err = kv.StoreConfig(config)
			if err != nil {
				return err
			}
		}

		if traefikConfiguration.GlobalConfiguration.ACME != nil {
			account := &acme.Account{}

			// Migrate ACME data from file to KV store if needed
			if len(traefikConfiguration.GlobalConfiguration.ACME.StorageFile) > 0 {
				account, err = migrateACMEData(traefikConfiguration.GlobalConfiguration.ACME.StorageFile)
				if err != nil {
					return err
				}
			}

			accountInitialized, err := keyExists(kv, traefikConfiguration.GlobalConfiguration.ACME.Storage)
			if err != nil && err != store.ErrKeyNotFound {
				return err
			}

			// Check to see if ACME account object is already in kv store
			if traefikConfiguration.GlobalConfiguration.ACME.OverrideCertificates || !accountInitialized {

				// Store the ACME Account into the KV Store
				// Certificates in KV Store will be overridden
				meta := cluster.NewMetadata(account)
				err = meta.Marshall()
				if err != nil {
					return err
				}

				source := staert.KvSource{
					Store:  kv,
					Prefix: traefikConfiguration.GlobalConfiguration.ACME.Storage,
				}

				err = source.StoreConfig(meta)
				if err != nil {
					return err
				}
			}

			// Force to delete storagefile
			return kv.Delete(kv.Prefix + "/acme/storagefile")
		}
		return nil
	}
}

func keyExists(source *staert.KvSource, key string) (bool, error) {
	list, err := source.List(key, nil)
	if err != nil {
		return false, err
	}

	return len(list) > 0, nil
}

// migrateACMEData allows migrating data from acme.json file to KV store in function of the file format
func migrateACMEData(fileName string) (*acme.Account, error) {

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	file, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// Check if the storage file is not empty before to get data
	account := &acme.Account{}
	if len(file) > 0 {
		accountFromNewFormat, err := acme.FromNewToOldFormat(fileName)
		if err != nil {
			return nil, err
		}

		if accountFromNewFormat == nil {
			// convert ACME json file to KV store (used for backward compatibility)
			localStore := acme.NewLocalStore(fileName)

			account, err = localStore.Get()
			if err != nil {
				return nil, err
			}

			err = account.RemoveAccountV1Values()
			if err != nil {
				return nil, err
			}
		} else {
			account = accountFromNewFormat
		}
	} else {
		log.Warnf("No data will be imported from the storageFile %q because it is empty.", fileName)
	}

	err = account.Init()
	return account, err
}

// CreateKvSource creates KvSource
// TLS support is enable for Consul and Etcd backends
func CreateKvSource(traefikConfiguration *cmd.TraefikConfiguration) (*staert.KvSource, error) {
	var kv *staert.KvSource
	var kvStore store.Store
	var err error

	switch {
	case traefikConfiguration.Consul != nil:
		kvStore, err = traefikConfiguration.Consul.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Consul.Prefix,
		}
	case traefikConfiguration.Etcd != nil:
		kvStore, err = traefikConfiguration.Etcd.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Etcd.Prefix,
		}
	case traefikConfiguration.Zookeeper != nil:
		kvStore, err = traefikConfiguration.Zookeeper.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Zookeeper.Prefix,
		}
	case traefikConfiguration.Boltdb != nil:
		kvStore, err = traefikConfiguration.Boltdb.CreateStore()
		kv = &staert.KvSource{
			Store:  kvStore,
			Prefix: traefikConfiguration.Boltdb.Prefix,
		}
	}
	return kv, err
}
