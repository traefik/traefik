package labels

import (
	"testing"
)

func TestLabelEq(t *testing.T) {
	label := Label("labelName")
	m := label.Eq("value")
	values, ok := m["label"]
	if !ok {
		t.Fatalf("expected a label key, got %v", m)
	}
	if len(values) != 1 {
		t.Fatalf("expected only one value, got %v", values)
	}
	if values[0] != "labelName=value" {
		t.Fatalf("expected 'labelName=value', got %s", values)
	}
}

func TestLabelEqString(t *testing.T) {
	label := Label("labelName")
	value := label.EqString("value")
	if value != `{"label":["labelName=value"]}` {
		t.Fatalf("expected '{labelName=value}', got %s", value)
	}
}

func TestLabelFilter(t *testing.T) {
	filters := []struct {
		key      string
		value    string
		expected string
	}{
		{
			"key", "value", `{"label":["key=value"]}`,
		}, {
			"key", "", `{"label":["key="]}`,
		}, {
			"", "", `{"label":["="]}`,
		},
	}
	for _, filter := range filters {
		actual := LabelFilterString(filter.key, filter.value)
		if actual != filter.expected {
			t.Fatalf("Expected '%s for key=%s and value=%s, got %s", filter.expected, filter.key, filter.value, actual)
		}
	}
}
