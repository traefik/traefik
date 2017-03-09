package project

import (
	"bytes"
	"io"
	"text/tabwriter"
)

// InfoPart holds key/value strings.
type InfoPart struct {
	Key, Value string
}

// InfoSet holds a list of Info.
type InfoSet []Info

// Info holds a list of InfoPart.
type Info []InfoPart

func (infos InfoSet) String(titleFlag bool) string {
	//no error checking, none of this should fail
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	tabwriter := tabwriter.NewWriter(buffer, 4, 4, 2, ' ', 0)

	first := true
	for _, info := range infos {
		if first && titleFlag {
			writeLine(tabwriter, true, info)
		}
		first = false
		writeLine(tabwriter, false, info)
	}

	tabwriter.Flush()
	return buffer.String()
}

func writeLine(writer io.Writer, key bool, info Info) {
	first := true
	for _, part := range info {
		if !first {
			writer.Write([]byte{'\t'})
		}
		first = false
		if key {
			writer.Write([]byte(part.Key))
		} else {
			writer.Write([]byte(part.Value))
		}
	}

	writer.Write([]byte{'\n'})
}
