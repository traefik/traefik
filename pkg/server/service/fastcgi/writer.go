package fastcgi

import (
	"bytes"
	"encoding/binary"
	"io"
)

var padding [255]byte

type fastcgiWriter struct {
	reqID     uint16
	reqWriter io.Writer
	buff      *bytes.Buffer
}

func (w *fastcgiWriter) writeBeginReq(role uint16, keepCon bool) error {
	var (
		contentLen uint16 = 8
		paddingLen        = calcPadding(contentLen)
	)

	content := [8]byte{0, 0, beginReqFlags(keepCon)}
	binary.BigEndian.PutUint16(content[:], role)

	w.buff.Reset()
	w.buff.Write(padding[:FastCgiHeaderSz])
	if err := w.writeHeader(FastCgiBeginRecord, contentLen, paddingLen); err != nil {
		return err
	}

	w.buff.Write(content[:])
	w.buff.Write(padding[:paddingLen])
	_, err := w.buff.WriteTo(w.reqWriter)

	return err
}

func (w *fastcgiWriter) writeParamsReq(params map[string]string) error {
	if err := w.writePairs(FastCgiParamsRecord, params); err != nil {
		return err
	}

	return w.terminateStream(FastCgiParamsRecord)
}

func (w *fastcgiWriter) writeStdinReq(stdin io.Reader) error {
	w.buff.Reset()

	for {
		// limit read up to FastCgiMaxContentSz
		lr := io.LimitReader(stdin, FastCgiMaxContentSz)
		// alloc space for header
		w.buff.Write(padding[:FastCgiHeaderSz])
		n, err := w.buff.ReadFrom(lr)
		if err != nil {
			return err
		}
		err = w.writeRecordFromBuff(FastCgiStdinRecord)
		if err != nil {
			return err
		}

		if n < FastCgiMaxContentSz {
			break
		}
	}

	return w.terminateStream(FastCgiStdinRecord)
}

func (w *fastcgiWriter) writePairs(recordType uint8, pairs map[string]string) error {
	w.buff.Reset()
	// space for header
	w.buff.Write(padding[:FastCgiHeaderSz])

	for k, v := range pairs {
		n, keyLenBin := encodeParamLen(k)
		m, valLenBin := encodeParamLen(v)

		pairSz := n + m + len(k) + len(v)
		if w.contentLen()+pairSz > FastCgiMaxContentSz {
			if err := w.writeRecordFromBuff(recordType); err != nil {
				return err
			}
			// space for header
			w.buff.Write(padding[:FastCgiHeaderSz])
		}
		w.buff.Write(keyLenBin)
		w.buff.Write(valLenBin)
		w.buff.WriteString(k)
		w.buff.WriteString(v)
	}

	return w.writeRecordFromBuff(recordType)
}

func (w *fastcgiWriter) writeRecordFromBuff(recordType uint8) error {
	var (
		contentLen = uint16(w.contentLen())
		paddingLen = calcPadding(contentLen)
	)

	if err := w.writeHeader(recordType, contentLen, paddingLen); err != nil {
		return err
	}
	w.buff.Write(padding[:paddingLen])
	_, err := w.buff.WriteTo(w.reqWriter)

	return err
}

func (w *fastcgiWriter) writeHeader(recordType uint8, contentLen uint16, paddingLen uint8) error {
	// header is always first 8 bytes
	dst := w.buff.Bytes()

	h := header{
		Version:       FastCgiVersion,
		Type:          recordType,
		ContentLength: contentLen,
		PaddingLength: paddingLen,
		Reserved:      0,
	}
	return h.encode(dst[:FastCgiHeaderSz])
}

func (w *fastcgiWriter) terminateStream(recType uint8) error {
	// write an empty record to terminate stream
	w.buff.Reset()
	w.buff.Write(padding[:FastCgiHeaderSz])
	if err := w.writeHeader(recType, 0, 0); err != nil {
		return err
	}
	_, err := w.buff.WriteTo(w.reqWriter)
	return err
}

func (w *fastcgiWriter) contentLen() int {
	return w.buff.Len() - FastCgiHeaderSz
}
