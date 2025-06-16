package fastcgi

import (
	"encoding/binary"
	"errors"
)

type header struct {
	Version       uint8
	Type          uint8
	ID            uint16
	ContentLength uint16
	PaddingLength uint8
	Reserved      uint8
}

func (h *header) encode(dst []byte) error {
	if len(dst) < FastCgiHeaderSz {
		return errors.New("dst too small for fastcgi record header")
	}
	dst[0] = h.Version
	dst[1] = h.Type
	binary.BigEndian.PutUint16(dst[2:], h.ID)
	binary.BigEndian.PutUint16(dst[4:], h.ContentLength)
	dst[6] = h.PaddingLength
	dst[7] = h.Reserved

	return nil
}

func (h *header) decode(src []byte) error {
	if len(src) < FastCgiHeaderSz {
		return errors.New("src too small for fastcgi record header")
	}
	h.Version = src[0]
	h.Type = src[1]
	h.ID = binary.BigEndian.Uint16(src[2:])
	h.ContentLength = binary.BigEndian.Uint16(src[4:])
	h.PaddingLength = src[6]
	h.Reserved = src[7]

	return nil
}

type endRequestBody struct {
	appStatus      string
	protocolStatus uint8
	reserved       []byte
}

func (r *endRequestBody) decode(src []byte) error {
	if len(src) < 8 {
		return errors.New("src too small for end request body")
	}
	r.appStatus = string(src[:4])
	r.protocolStatus = src[4]
	r.reserved = src[5:8]
	return nil
}
