package options

import (
	"testing"
)

func TestImageType(t *testing.T) {
	cases := []struct {
		imageType string
		valid     bool
	}{
		{
			imageType: "",
			valid:     true,
		},
		{
			imageType: " ",
			valid:     false,
		},
		{
			imageType: "hello",
			valid:     false,
		},
		{
			imageType: "local",
			valid:     true,
		},
		{
			imageType: "all",
			valid:     true,
		},
	}
	for _, c := range cases {
		i := ImageType(c.imageType)
		if i.Valid() != c.valid {
			t.Errorf("expected %v, got %v, for %v", c.valid, i.Valid(), c)
		}
	}
}
