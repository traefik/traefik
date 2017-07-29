package zk

import "testing"

func TestFormatServers(t *testing.T) {
	t.Parallel()
	servers := []string{"127.0.0.1:2181", "127.0.0.42", "127.0.42.1:8811"}
	r := []string{"127.0.0.1:2181", "127.0.0.42:2181", "127.0.42.1:8811"}
	for i, s := range FormatServers(servers) {
		if s != r[i] {
			t.Errorf("%v should equal %v", s, r[i])
		}
	}
}
