package util

import (
	"encoding/xml"
	"time"
)

// KeyValueXML resembles the complex struct for getting key/value pairs from XML
type KeyValueXML struct {
	Cont []struct {
		Item []struct {
			Key   string `xml:"key"`
			Value string `xml:"value"`
		} `xml:"item"`
	} `xml:"item"`
}

// XMLTime is a custom type to decode XML values to time.Time directly
type XMLTime struct {
	time.Time
}

// UnmarshalXML is implemented to be able act as custom XML type
// it tries to parse time from given elements value
func (x *XMLTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}

	if p, _ := time.Parse("2006-01-02 15:04:05", v); !p.IsZero() {
		*x = XMLTime{p}
	} else if p, _ := time.Parse("2006-01-02", v); !p.IsZero() {
		*x = XMLTime{p}
	}
	return nil
}
