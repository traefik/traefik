package gettext

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ojii/gettext.go/pluralforms"
	"log"
	"os"
	"strings"
)

const le_magic = 0x950412de
const be_magic = 0xde120495

type header struct {
	Version           uint32
	NumStrings        uint32
	MasterIndex       uint32
	TranslationsIndex uint32
}

func (header header) get_major_version() uint32 {
	return header.Version >> 16
}

func (header header) get_minor_version() uint32 {
	return header.Version & 0xffff
}

// Catalog of translations for a given locale.
type Catalog interface {
	Gettext(msgid string) string
	NGettext(msgid string, msgid_plural string, n uint32) string
}

type mocatalog struct {
	header      header
	language    string
	messages    map[string][]string
	pluralforms pluralforms.Expression
	info        map[string]string
	charset     string
}

type nullcatalog struct{}

func (catalog nullcatalog) Gettext(msgid string) string {
	return msgid
}

func (catalog nullcatalog) NGettext(msgid string, msgid_plural string, n uint32) string {
	if n == 1 {
		return msgid
	} else {
		return msgid_plural
	}
}

func (catalog mocatalog) Gettext(msgid string) string {
	msgstrs, ok := catalog.messages[msgid]
	if !ok {
		return msgid
	}
	return msgstrs[0]
}

func (catalog mocatalog) NGettext(msgid string, msgid_plural string, n uint32) string {
	msgstrs, ok := catalog.messages[msgid]
	if !ok {
		if n == 1 {
			return msgid
		} else {
			return msgid_plural
		}
	} else {
		/* Bogus/missing pluralforms in mo */
		if catalog.pluralforms == nil {
			/* Use the Germanic plural rule.  */
			if n == 1 {
				return msgstrs[0]
			} else {
				return msgstrs[1]
			}
		}

		index := catalog.pluralforms.Eval(n)
		if index > len(msgstrs) {
			if n == 1 {
				return msgid
			} else {
				return msgid_plural
			}
		}
		return msgstrs[index]
	}
}

type len_offset struct {
	Len uint32
	Off uint32
}

func read_len_off(index uint32, file *os.File, order binary.ByteOrder) (len_offset, error) {
	lenoff := len_offset{}
	buf := make([]byte, 8)
	_, err := file.Seek(int64(index), os.SEEK_SET)
	if err != nil {
		return lenoff, err
	}
	_, err = file.Read(buf)
	if err != nil {
		return lenoff, err
	}
	buffer := bytes.NewBuffer(buf)
	err = binary.Read(buffer, order, &lenoff)
	if err != nil {
		return lenoff, err
	}
	return lenoff, nil
}

func read_message(file *os.File, lenoff len_offset) (string, error) {
	_, err := file.Seek(int64(lenoff.Off), os.SEEK_SET)
	if err != nil {
		return "", err
	}
	buf := make([]byte, lenoff.Len)
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (catalog *mocatalog) read_info(info string) error {
	lastk := ""
	for _, line := range strings.Split(info, "\n") {
		item := strings.TrimSpace(line)
		if len(item) == 0 {
			continue
		}
		var k string
		var v string
		if strings.Contains(item, ":") {
			tmp := strings.SplitN(item, ":", 2)
			k = strings.ToLower(strings.TrimSpace(tmp[0]))
			v = strings.TrimSpace(tmp[1])
			catalog.info[k] = v
			lastk = k
		} else if len(lastk) != 0 {
			catalog.info[lastk] += "\n" + item
		}
		if k == "content-type" {
			catalog.charset = strings.Split(v, "charset=")[1]
		} else if k == "plural-forms" {
			p := strings.Split(v, ";")[1]
			s := strings.Split(p, "plural=")[1]
			expr, err := pluralforms.Compile(s)
			if err != nil {
				return err
			}
			catalog.pluralforms = expr
		}
	}
	return nil
}

// ParseMO parses a mo file into a Catalog if possible.
func ParseMO(file *os.File) (Catalog, error) {
	var order binary.ByteOrder
	header := header{}
	catalog := mocatalog{
		header:   header,
		info:     make(map[string]string),
		messages: make(map[string][]string),
	}
	magic := make([]byte, 4)
	_, err := file.Read(magic)
	if err != nil {
		return catalog, err
	}
	magic_number := binary.LittleEndian.Uint32(magic)
	switch magic_number {
	case le_magic:
		order = binary.LittleEndian
	case be_magic:
		order = binary.BigEndian
	default:
		return catalog, fmt.Errorf("Wrong magic %d", magic_number)
	}
	raw_headers := make([]byte, 32)
	_, err = file.Read(raw_headers)
	if err != nil {
		return catalog, err
	}
	buffer := bytes.NewBuffer(raw_headers)
	err = binary.Read(buffer, order, &header)
	if err != nil {
		return catalog, err
	}
	if (header.get_major_version() != 0) && (header.get_major_version() != 1) {
		log.Printf("major %d minor %d", header.get_major_version(), header.get_minor_version())
		return catalog, fmt.Errorf("Unsupported version: %d.%d", header.get_major_version(), header.get_minor_version())
	}
	current_master_index := header.MasterIndex
	current_transl_index := header.TranslationsIndex
	var index uint32 = 0
	for ; index < header.NumStrings; index++ {
		mlenoff, err := read_len_off(current_master_index, file, order)
		if err != nil {
			return catalog, err
		}
		tlenoff, err := read_len_off(current_transl_index, file, order)
		if err != nil {
			return catalog, err
		}
		msgid, err := read_message(file, mlenoff)
		if err != nil {
			return catalog, nil
		}
		msgstr, err := read_message(file, tlenoff)
		if err != nil {
			return catalog, err
		}
		if mlenoff.Len == 0 {
			err = catalog.read_info(msgstr)
			if err != nil {
				return catalog, err
			}
		}
		if strings.Contains(msgid, "\x00") {
			// Plural!
			msgidsingular := strings.Split(msgid, "\x00")[0]
			translations := strings.Split(msgstr, "\x00")
			catalog.messages[msgidsingular] = translations
		} else {
			catalog.messages[msgid] = []string{msgstr}
		}

		current_master_index += 8
		current_transl_index += 8
	}
	return catalog, nil
}
