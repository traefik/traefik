package crd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func Test_buildStickyFromCRD(t *testing.T) {
	t.Parallel()

	t.Run("cookie", func(t *testing.T) {
		t.Parallel()

		sticky := &dynamic.Sticky{
			Cookie: &dynamic.Cookie{
				Name:     "crd-cookie",
				HTTPOnly: true,
			},
		}

		result := buildStickyFromCRD(sticky)
		if assert.NotNil(t, result) {
			if assert.NotNil(t, result.Cookie) {
				assert.Equal(t, "crd-cookie", result.Cookie.Name)
				assert.True(t, result.Cookie.HTTPOnly)
			}
			assert.Nil(t, result.Header)
		}
	})

	t.Run("header", func(t *testing.T) {
		t.Parallel()

		sticky := &dynamic.Sticky{
			Header: &dynamic.Header{Name: "X-CRD-Session"},
		}

		result := buildStickyFromCRD(sticky)
		if assert.NotNil(t, result) {
			if assert.NotNil(t, result.Header) {
				assert.Equal(t, "X-CRD-Session", result.Header.Name)
			}
			assert.Nil(t, result.Cookie)
		}
	})
}
