package kingpin

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTwoColumns(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	formatTwoColumns(buf, 2, 2, 20, [][2]string{
		{"--hello", "Hello world help with something that is cool."},
	})
	expected := `  --hello  Hello
           world
           help with
           something
           that is
           cool.
`
	assert.Equal(t, expected, buf.String())
}

func TestFormatTwoColumnsWide(t *testing.T) {
	samples := [][2]string{
		{strings.Repeat("x", 19), "19 chars"},
		{strings.Repeat("x", 20), "20 chars"}}
	buf := bytes.NewBuffer(nil)
	formatTwoColumns(buf, 0, 0, 200, samples)
	fmt.Println(buf.String())
	expected := `xxxxxxxxxxxxxxxxxxx19 chars
xxxxxxxxxxxxxxxxxxxx
                   20 chars
`
	assert.Equal(t, expected, buf.String())
}

func TestHiddenCommand(t *testing.T) {
	templates := []struct{ name, template string }{
		{"default", DefaultUsageTemplate},
		{"Compact", CompactUsageTemplate},
		{"Long", LongHelpTemplate},
		{"Man", ManPageTemplate},
	}

	var buf bytes.Buffer
	t.Log("1")

	a := New("test", "Test").Writer(&buf).Terminate(nil)
	a.Command("visible", "visible")
	a.Command("hidden", "hidden").Hidden()

	for _, tp := range templates {
		buf.Reset()
		a.UsageTemplate(tp.template)
		a.Parse(nil)
		// a.Parse([]string{"--help"})
		usage := buf.String()
		t.Logf("Usage for %s is:\n%s\n", tp.name, usage)

		assert.NotContains(t, usage, "hidden")
		assert.Contains(t, usage, "visible")
	}
}
