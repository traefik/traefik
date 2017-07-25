package lib_test

import (
	"testing"

	"github.com/hashicorp/consul/lib"
)

func TestMathAbsInt(t *testing.T) {
	tests := [][3]int{{1, 1}, {-1, 1}, {0, 0}}
	for _, test := range tests {
		if test[1] != lib.AbsInt(test[0]) {
			t.Fatalf("expected %d, got %d", test[1], test[0])
		}
	}
}

func TestMathMaxInt(t *testing.T) {
	tests := [][3]int{{1, 2, 2}, {-1, 1, 1}, {2, 0, 2}}
	for _, test := range tests {
		expected := test[2]
		actual := lib.MaxInt(test[0], test[1])
		if expected != actual {
			t.Fatalf("expected %d, got %d", expected, actual)
		}
	}
}

func TestMathMinInt(t *testing.T) {
	tests := [][3]int{{1, 2, 1}, {-1, 1, -1}, {2, 0, 0}}
	for _, test := range tests {
		expected := test[2]
		actual := lib.MinInt(test[0], test[1])
		if expected != actual {
			t.Fatalf("expected %d, got %d", expected, actual)
		}
	}
}
