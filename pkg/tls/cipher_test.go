package tls

import (
	"testing"
)

func TestCiphersMapsSync(t *testing.T) {
	for k, v := range CipherSuites {
		if rv, ok := CipherSuitesReversed[v]; !ok {
			t.Errorf("Maps not in sync: `%v` key is missing in tls.CipherSuitesReversed", v)
		} else if k != rv {
			t.Errorf("Maps not in sync: tls.CipherSuites[%v] = `%v` AND tls.CipherSuitesReversed[`%v`] = `%v`", k, v, v, rv)
		}
	}

	for k, v := range CipherSuitesReversed {
		if rv, ok := CipherSuites[v]; !ok {
			t.Errorf("Maps not in sync: `%v` key is missing in tls.CipherSuites", v)
		} else if k != rv {
			t.Errorf("Maps not in sync: tls.CipherSuitesReversed[`%v`] = `%v` AND tls.CipherSuites[`%v`] = `%v`", k, v, v, rv)
		}
	}
}
