package quiclb

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	serverID := [serverIDLen]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	testCases := []struct {
		desc             string
		serverIDBase64   string
		expectedServerID *[serverIDLen]byte
		expectedErr      string
	}{
		{
			desc: "empty",
		},
		{
			desc:             "valid",
			serverIDBase64:   base64.StdEncoding.EncodeToString(serverID[:]),
			expectedServerID: &serverID,
		},
		{
			desc:           "invalid base64",
			serverIDBase64: "not-valid-base64!!!",
			expectedErr:    "decoding server ID",
		},
		{
			desc:           "too short",
			serverIDBase64: base64.StdEncoding.EncodeToString([]byte{0x01, 0x02, 0x03, 0x04}),
			expectedErr:    "server ID must be 8 bytes, got 4",
		},
		{
			desc:           "too long",
			serverIDBase64: base64.StdEncoding.EncodeToString([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C}),
			expectedErr:    "server ID must be 8 bytes, got 12",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gen, err := NewGenerator(test.serverIDBase64)
			if test.expectedErr != "" {
				require.ErrorContains(t, err, test.expectedErr)
				assert.Nil(t, gen)
				return
			}
			require.NoError(t, err)

			if test.expectedServerID == nil {
				assert.Nil(t, gen)
				return
			}
			require.NotNil(t, gen)
			assert.Equal(t, *test.expectedServerID, gen.serverID)
		})
	}
}

func TestGenerateConnectionID(t *testing.T) {
	serverID := [serverIDLen]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22}

	gen, err := NewGenerator(base64.StdEncoding.EncodeToString(serverID[:]))
	require.NoError(t, err)

	cid1, err := gen.GenerateConnectionID()
	require.NoError(t, err)

	cid2, err := gen.GenerateConnectionID()
	require.NoError(t, err)

	for _, cid := range [][]byte{cid1.Bytes(), cid2.Bytes()} {
		require.Len(t, cid, 17)
		assert.Len(t, cid, gen.ConnectionIDLen())
		assert.Equal(t, byte(0x0F), cid[0])
		assert.Equal(t, serverID[:], cid[1:1+serverIDLen])
	}

	assert.NotEqual(t, cid1.Bytes()[1+serverIDLen:], cid2.Bytes()[1+serverIDLen:])
}
