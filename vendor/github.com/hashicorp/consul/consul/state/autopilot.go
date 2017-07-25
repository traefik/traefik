package state

import (
	"fmt"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/go-memdb"
)

// Autopilot is used to pull the autopilot config from the snapshot.
func (s *StateSnapshot) Autopilot() (*structs.AutopilotConfig, error) {
	c, err := s.tx.First("autopilot-config", "id")
	if err != nil {
		return nil, err
	}

	config, ok := c.(*structs.AutopilotConfig)
	if !ok {
		return nil, nil
	}

	return config, nil
}

// Autopilot is used when restoring from a snapshot.
func (s *StateRestore) Autopilot(config *structs.AutopilotConfig) error {
	if err := s.tx.Insert("autopilot-config", config); err != nil {
		return fmt.Errorf("failed restoring autopilot config: %s", err)
	}

	return nil
}

// AutopilotConfig is used to get the current Autopilot configuration.
func (s *StateStore) AutopilotConfig() (uint64, *structs.AutopilotConfig, error) {
	tx := s.db.Txn(false)
	defer tx.Abort()

	// Get the autopilot config
	c, err := tx.First("autopilot-config", "id")
	if err != nil {
		return 0, nil, fmt.Errorf("failed autopilot config lookup: %s", err)
	}

	config, ok := c.(*structs.AutopilotConfig)
	if !ok {
		return 0, nil, nil
	}

	return config.ModifyIndex, config, nil
}

// AutopilotSetConfig is used to set the current Autopilot configuration.
func (s *StateStore) AutopilotSetConfig(idx uint64, config *structs.AutopilotConfig) error {
	tx := s.db.Txn(true)
	defer tx.Abort()

	s.autopilotSetConfigTxn(idx, tx, config)

	tx.Commit()
	return nil
}

// AutopilotCASConfig is used to try updating the Autopilot configuration with a
// given Raft index. If the CAS index specified is not equal to the last observed index
// for the config, then the call is a noop,
func (s *StateStore) AutopilotCASConfig(idx, cidx uint64, config *structs.AutopilotConfig) (bool, error) {
	tx := s.db.Txn(true)
	defer tx.Abort()

	// Check for an existing config
	existing, err := tx.First("autopilot-config", "id")
	if err != nil {
		return false, fmt.Errorf("failed autopilot config lookup: %s", err)
	}

	// If the existing index does not match the provided CAS
	// index arg, then we shouldn't update anything and can safely
	// return early here.
	e, ok := existing.(*structs.AutopilotConfig)
	if !ok || e.ModifyIndex != cidx {
		return false, nil
	}

	s.autopilotSetConfigTxn(idx, tx, config)

	tx.Commit()
	return true, nil
}

func (s *StateStore) autopilotSetConfigTxn(idx uint64, tx *memdb.Txn, config *structs.AutopilotConfig) error {
	// Check for an existing config
	existing, err := tx.First("autopilot-config", "id")
	if err != nil {
		return fmt.Errorf("failed autopilot config lookup: %s", err)
	}

	// Set the indexes.
	if existing != nil {
		config.CreateIndex = existing.(*structs.AutopilotConfig).CreateIndex
	} else {
		config.CreateIndex = idx
	}
	config.ModifyIndex = idx

	if err := tx.Insert("autopilot-config", config); err != nil {
		return fmt.Errorf("failed updating autopilot config: %s", err)
	}
	return nil
}
