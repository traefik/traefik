package acme

import (
	"reflect"
	"testing"
)

func TestDomainsSet(t *testing.T) {
	checkMap := map[string]Domains{
		"":                                   {},
		"foo.com":                            {Domain{Main: "foo.com", SANs: []string{}}},
		"foo.com,bar.net":                    {Domain{Main: "foo.com", SANs: []string{"bar.net"}}},
		"foo.com,bar1.net,bar2.net,bar3.net": {Domain{Main: "foo.com", SANs: []string{"bar1.net", "bar2.net", "bar3.net"}}},
	}
	for in, check := range checkMap {
		ds := Domains{}
		ds.Set(in)
		if !reflect.DeepEqual(check, ds) {
			t.Errorf("Expected %+v\nGo %+v", check, ds)
		}
	}
}

func TestDomainsSetAppend(t *testing.T) {
	inSlice := []string{
		"",
		"foo1.com",
		"foo2.com,bar.net",
		"foo3.com,bar1.net,bar2.net,bar3.net",
	}
	checkSlice := []Domains{
		{},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}}},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}},
			Domain{
				Main: "foo2.com",
				SANs: []string{"bar.net"}}},
		{
			Domain{
				Main: "foo1.com",
				SANs: []string{}},
			Domain{
				Main: "foo2.com",
				SANs: []string{"bar.net"}},
			Domain{Main: "foo3.com",
				SANs: []string{"bar1.net", "bar2.net", "bar3.net"}}},
	}
	ds := Domains{}
	for i, in := range inSlice {
		ds.Set(in)
		if !reflect.DeepEqual(checkSlice[i], ds) {
			t.Errorf("Expected  %s %+v\nGo %+v", in, checkSlice[i], ds)
		}
	}
}
