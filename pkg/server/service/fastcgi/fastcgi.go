package fastcgi

import "encoding/binary"

const (
	FastCgiHeaderSz     = 8
	FastCgiMaxContentSz = 65535

	FastCgiVersion  uint8 = 1
	FastCgiKeepConn uint8 = 1
)

const (
	FastCgiRoleResponder uint16 = iota + 1
	FastCgiRoleAuthorizer
	FastCgiRoleFilter
)

const (
	FastCgiBeginRecord uint8 = iota + 1
	FastCgiAbortRecord
	FastCgiEndRecord
	FastCgiParamsRecord
	FastCgiStdinRecord
	FastCgiStdoutRecord
	FastCgiStderrRecord
	FastCgiDataRecord
	FastCgiGetValuesRecord
	FastCgiGetValuesResultRecord
	FastCgiUnknownRecord
)

const (
	FastCgiRequestComplete uint8 = iota
	FastCgiCantMultiplexConn
	FastCgiOverloaded
	FastCgiUnknownRole
)

const (
	FastCgiMaxConnsKey   = "FCGI_MAX_CONNS"
	FastCgiMaxReqsKey    = "FCGI_MAX_REQS"
	FastCgiMpxEnabledKey = "FCGI_MPXS_CONNS"
)

type env map[string]string

func beginReqFlags(keepCon bool) byte {
	if keepCon {
		return FastCgiKeepConn
	}
	return 0
}

func calcPadding(contentLen uint16) uint8 {
	return uint8(-contentLen & 7)
}

func encodeParamLen(param string) (int, []byte) {
	length := len(param)
	// one byte for length
	if length < 128 {
		return 1, []byte{byte(length)}
	}

	// 4 bytes length with 0x80 flag
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], uint32(length))
	encoded[0] |= 0x80
	return 4, encoded[:]
}
