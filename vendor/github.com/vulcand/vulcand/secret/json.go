package secret

import (
	"encoding/json"
	"fmt"
	"github.com/vulcand/vulcand/engine"
)

type sealedValue struct {
	Encryption string
	Value      SealedBytes
}

func SealedValueToJSON(b *SealedBytes) ([]byte, error) {
	data := &sealedValue{
		Encryption: encryptionSecretBox,
		Value:      *b,
	}
	return json.Marshal(&data)
}

func SealedValueFromJSON(bytes []byte) (*SealedBytes, error) {
	var v *sealedValue
	if err := json.Unmarshal(bytes, &v); err != nil {
		return nil, err
	}
	if v.Encryption != encryptionSecretBox {
		return nil, fmt.Errorf("unsupported encryption type: '%s'", v.Encryption)
	}
	return &v.Value, nil
}

func SealKeyPairToJSON(box *Box, keyPair *engine.KeyPair) ([]byte, error) {
	bytes, err := json.Marshal(keyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to JSON encode certificate: %s", bytes)
	}

	sealed, err := box.Seal(bytes)
	if err != nil {
		return nil, err
	}

	return SealedValueToJSON(sealed)
}

const encryptionSecretBox = "secretbox.v1"
