// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2016 Canonical Ltd
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
	"sync"
)

type memoryKeypairManager struct {
	pairs map[string]PrivateKey
	mu    sync.RWMutex
}

// NewMemoryKeypairManager creates a new key pair manager with a memory backstore.
func NewMemoryKeypairManager() KeypairManager {
	return &memoryKeypairManager{
		pairs: make(map[string]PrivateKey),
	}
}

func (mkm *memoryKeypairManager) Put(privKey PrivateKey) error {
	mkm.mu.Lock()
	defer mkm.mu.Unlock()

	keyID := privKey.PublicKey().ID()
	if mkm.pairs[keyID] != nil {
		return errKeypairAlreadyExists
	}
	mkm.pairs[keyID] = privKey
	return nil
}

func (mkm *memoryKeypairManager) Get(keyID string) (PrivateKey, error) {
	mkm.mu.RLock()
	defer mkm.mu.RUnlock()

	privKey := mkm.pairs[keyID]
	if privKey == nil {
		return nil, errKeypairNotFound
	}
	return privKey, nil
}
