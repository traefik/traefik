package dynamic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func pointer[T any](v T) *T { return &v }

func TestMirrorService_SetDefaults(t *testing.T) {
	const dflt = MirroringServiceDefaultPercent // 100

	t.Run("unset ⇒ default", func(t *testing.T) {
		ms := MirrorService{}
		ms.SetDefaults()

		require.NotNil(t, ms.Percent)
		require.Equal(t, dflt, *ms.Percent)
	})

	t.Run("explicit 0 ⇒ keep 0", func(t *testing.T) {
		zero := 0
		ms := MirrorService{Percent: pointer(zero)}
		ms.SetDefaults()

		require.NotNil(t, ms.Percent)
		require.Equal(t, 0, *ms.Percent)
	})

	t.Run("explicit 42 ⇒ keep 42", func(t *testing.T) {
		v := 42
		ms := MirrorService{Percent: pointer(v)}
		ms.SetDefaults()

		require.NotNil(t, ms.Percent)
		require.Equal(t, 42, *ms.Percent)
	})
}
