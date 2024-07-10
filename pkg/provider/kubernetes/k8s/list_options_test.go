package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_concatenateSelector(t *testing.T) {
	cases := map[string]struct {
		A        string
		B        string
		Expected string
	}{
		"both empty": {},
		"a empty is b": {
			B:        "teapot",
			Expected: "teapot",
		},
		"b empty is a": {
			A:        "teapot",
			Expected: "teapot",
		},
		"both non-empty": {
			A:        "hotdog",
			B:        "not-a-hotdog",
			Expected: "hotdog,not-a-hotdog",
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			actual := concatenateSelector(test.A, test.B)
			assert.Equal(t, test.Expected, actual)
		})
	}
}

func Test_mergeListOptions(t *testing.T) {
	cases := map[string]struct {
		In        *metav1.ListOptions
		Overrides *metav1.ListOptions
		Expected  *metav1.ListOptions
	}{
		"nil override is fine": {
			In:       &metav1.ListOptions{},
			Expected: &metav1.ListOptions{},
		},
		"label selector is concatenated": {
			In: &metav1.ListOptions{
				LabelSelector: "a=b",
			},
			Overrides: &metav1.ListOptions{
				LabelSelector: "c=d",
			},
			Expected: &metav1.ListOptions{
				LabelSelector: "a=b,c=d",
			},
		},
		"field selector is concatenated": {
			In: &metav1.ListOptions{
				FieldSelector: "a=b",
			},
			Overrides: &metav1.ListOptions{
				FieldSelector: "c=d",
			},
			Expected: &metav1.ListOptions{
				FieldSelector: "a=b,c=d",
			},
		},
		"other fields are copied": {
			In: &metav1.ListOptions{},
			Overrides: &metav1.ListOptions{
				Watch: true,
			},
			Expected: &metav1.ListOptions{
				Watch: true,
			},
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			mergeListOptions(test.In, test.Overrides)
			assert.Equal(t, test.Expected, test.In)
		})
	}
}
