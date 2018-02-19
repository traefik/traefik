package main

import (
	"encoding/json"
	"fmt"
	stdlog "log"

	"github.com/abronan/valkeyrie/store"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/cluster"
)

func newStoreConfigCmd(traefikConfiguration *TraefikConfiguration, traefikPointersConfiguration *TraefikConfiguration) *flaeg.Command {
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

func runStoreConfig(kv *staert.KvSource, traefikConfiguration *TraefikConfiguration) func() error {
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
			var object cluster.Object
			if len(traefikConfiguration.GlobalConfiguration.ACME.StorageFile) > 0 {
				// convert ACME json file to KV store
				localStore := acme.NewLocalStore(traefikConfiguration.GlobalConfiguration.ACME.StorageFile)
				object, err = localStore.Load()
				if err != nil {
					return err
				}
			} else {
				// Create an empty account to create all the keys into the KV store
				account := &acme.Account{}
				err = account.Init()
				if err != nil {
					return err
				}

				object = account
			}

			meta := cluster.NewMetadata(object)
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
			// Force to delete storagefile
			err = kv.Delete(kv.Prefix + "/acme/storagefile")
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// createKvSource creates KvSource
// TLS support is enable for Consul and Etcd backends
func createKvSource(traefikConfiguration *TraefikConfiguration) (*staert.KvSource, error) {
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
