package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	testCases := []struct {
		desc     string
		name     string
		expected string
	}{
		{
			desc:     "leaves alphanumerics and single dashes",
			name:     "prod-host-rewrite-portal-abc",
			expected: "prod-host-rewrite-portal-abc",
		},
		{
			desc:     "collapses consecutive dashes",
			name:     "prod-host-rewrite--portal--abc",
			expected: "prod-host-rewrite-portal-abc",
		},
		{
			desc:     "collapses mixed separators",
			name:     "foo..bar__baz",
			expected: "foo-bar-baz",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, Normalize(test.name))
		})
	}
}

func TestHasCollapsibleSeparators(t *testing.T) {
	testCases := []struct {
		desc     string
		name     string
		expected bool
	}{
		{desc: "single dashes", name: "a-b-c", expected: false},
		{desc: "double dashes", name: "a--b", expected: true},
		{desc: "dots", name: "a..b", expected: true},
		{desc: "empty", name: "", expected: false},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, HasCollapsibleSeparators(test.name))
		})
	}
}

func TestMiddlewareNameHint(t *testing.T) {
	assert.Empty(t, MiddlewareNameHint("auth@file"))
	assert.Empty(t, MiddlewareNameHint("prod-host-rewrite-portal-abc@kubernetescrd"))

	hint := MiddlewareNameHint("prod-host-rewrite--portal--abc@kubernetescrd")
	assert.Contains(t, hint, "prod-host-rewrite--portal--abc")
	assert.Contains(t, hint, "prod-host-rewrite-portal-abc")
}
