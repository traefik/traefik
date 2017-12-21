package xmlutils

import (
	"encoding/xml"
	"errors"
	"sort"
	"strings"

	"github.com/beevik/etree"
	"github.com/containous/traefik/middlewares/audittap/types"
)

// XMLFragment represents XML content as a []byte slice that can be populated by xml.Decoder
type XMLFragment struct {
	InnerXML []byte `xml:",innerxml"`
}

// XMLToDataMap transforms Etree elements into a DataMap
func XMLToDataMap(xml []*etree.Element, excludeTags []string) types.DataMap {
	m := types.DataMap{}
	sort.Strings(excludeTags)

	for _, el := range xml {
		if isExcluded(el.Tag, excludeTags) {
			continue
		}

		if len(el.ChildElements()) > 0 {
			submap := XMLToDataMap(el.ChildElements(), excludeTags)
			attrs := attributesToDataMap(el)
			submap.AddAll(attrs)
			m[el.Tag] = submap
		} else if len(el.Text()) > 0 && len(el.Attr) > 0 {
			submap := attributesToDataMap(el)
			submap[el.Tag] = strings.TrimSpace(el.Text())
			m[el.Tag] = submap
		} else if len(el.Text()) > 0 {
			m[el.Tag] = strings.TrimSpace(el.Text())
		}
	}

	return m
}

func attributesToDataMap(el *etree.Element) types.DataMap {
	m := types.DataMap{}
	exclusions := []string{"xmlns"} // Sort when more than 1 element
	for _, at := range el.Attr {
		if !isExcluded(at.Key, exclusions) && len(at.Value) > 0 {
			m[at.Key] = at.Value
		}
	}
	return m
}

// isExcluded checks if the value s is in slice (must be supplied sorted)
func isExcluded(s string, exclusions []string) bool {
	i := sort.SearchStrings(exclusions, s)
	return i < len(exclusions) && exclusions[i] == s
}

// ElementInnerToDocument transforms xml.StartElement contents to etree.Document. It ignores any attributes in src.
func ElementInnerToDocument(src *xml.StartElement, decoder *xml.Decoder) (*etree.Document, error) {
	var x XMLFragment
	err := decoder.DecodeElement(&x, src)
	if err != nil {
		return nil, err
	}
	doc := etree.NewDocument()
	// Missing the outer element so append it as bytes
	xml := []byte{}
	xml = append(xml, []byte("<"+src.Name.Local+">")...)
	xml = append(xml, x.InnerXML...)
	xml = append(xml, []byte("</"+src.Name.Local+">")...)
	err = doc.ReadFromBytes(xml)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// ScrollToNextElement iterates the decoder returning the first xml.StartElement found
func ScrollToNextElement(decoder *xml.Decoder) (*xml.StartElement, error) {
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			return &se, nil
		}
	}
	return &xml.StartElement{}, errors.New("No root XML element found")

}
