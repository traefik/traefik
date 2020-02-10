package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddInContext(t *testing.T) {
	testCases := []struct {
		desc     string
		ctx      context.Context
		name     string
		expected string
	}{
		{
			desc:     "without provider information",
			ctx:      context.Background(),
			name:     "test",
			expected: "",
		},
		{
			desc:     "provider name embedded in element name",
			ctx:      context.Background(),
			name:     "test@foo",
			expected: "foo",
		},
		{
			desc:     "provider name in context",
			ctx:      context.WithValue(context.Background(), key, "foo"),
			name:     "test",
			expected: "foo",
		},
		{
			desc:     "provider name in context and different provider name embedded in element name",
			ctx:      context.WithValue(context.Background(), key, "foo"),
			name:     "test@fii",
			expected: "fii",
		},
		{
			desc:     "provider name in context and same provider name embedded in element name",
			ctx:      context.WithValue(context.Background(), key, "foo"),
			name:     "test@foo",
			expected: "foo",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			newCtx := AddInContext(test.ctx, test.name)

			var providerName string
			if name, ok := newCtx.Value(key).(string); ok {
				providerName = name
			}

			assert.Equal(t, test.expected, providerName)
		})
	}
}

func TestGetQualifiedName(t *testing.T) {
	testCases := []struct {
		desc     string
		ctx      context.Context
		name     string
		expected string
	}{
		{
			desc:     "empty name",
			ctx:      context.Background(),
			name:     "",
			expected: "",
		},
		{
			desc:     "without provider",
			ctx:      context.Background(),
			name:     "test",
			expected: "test",
		},
		{
			desc:     "with explicit provider",
			ctx:      context.Background(),
			name:     "test@foo",
			expected: "test@foo",
		},
		{
			desc:     "with provider in context",
			ctx:      context.WithValue(context.Background(), key, "foo"),
			name:     "test",
			expected: "test@foo",
		},
		{
			desc:     "with provider in context and explicit name",
			ctx:      context.WithValue(context.Background(), key, "foo"),
			name:     "test@fii",
			expected: "test@fii",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			qualifiedName := GetQualifiedName(test.ctx, test.name)

			assert.Equal(t, test.expected, qualifiedName)
		})
	}
}
