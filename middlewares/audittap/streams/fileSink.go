package streams

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/containous/traefik/middlewares/audittap/types"
)

var opener = []byte{'['}
var closer = []byte{']', '\n'}
var commaNewline = []byte{',', '\n'}
var newline = []byte{'\n'}

type fileSink struct {
	w       io.WriteCloser
	lineEnd []byte
}

// NewFileSink creates a new file sink
func NewFileSink(file, backend string) (AuditSink, error) {
	flag := os.O_RDWR | os.O_CREATE
	if strings.HasPrefix(file, ">>") {
		file = strings.TrimSpace(file[2:])
	} else {
		flag |= os.O_TRUNC
	}
	name := determineFilename(file, backend)
	f, err := os.OpenFile(name, flag, 0644)
	if err != nil {
		return nil, err
	}
	f.Write(opener)
	return &fileSink{f, newline}, nil
}

func determineFilename(file, backend string) string {
	name := file
	if backend != "" {
		if strings.HasSuffix(name, ".json") {
			name = name[:len(name)-5]
		}
		name = fmt.Sprintf("%s-%s.json", name, backend)
	} else if !strings.HasSuffix(name, ".json") {
		name = name + ".json"
	}
	return name
}

func (fas *fileSink) Audit(encoded types.Encoded) error {
	fas.w.Write(fas.lineEnd)
	_, err := fas.w.Write(encoded.Bytes)
	fas.lineEnd = commaNewline
	return err
}

func (fas *fileSink) Close() error {
	fas.w.Write(newline)
	fas.w.Write(closer)
	return fas.w.Close()
}
