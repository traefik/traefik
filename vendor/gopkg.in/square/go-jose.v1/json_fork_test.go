// +build !std_json

/*-
 * Copyright 2014 Square Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jose

import (
	"testing"

	"gopkg.in/square/go-jose.v1/json"
)

type CaseSensitive struct {
	A int `json:"Test"`
	B int `json:"test"`
	C int `json:"TEST"`
}

type UnicodeTest struct {
	Sig string `json:"sig"`
}

func TestUnicodeComparison(t *testing.T) {
	// Some tests from RFC 7515, Section 10.13
	raw := []byte(`{"\u0073ig":"foo"}`)
	var ut UnicodeTest
	err := json.Unmarshal(raw, &ut)
	if err != nil {
		t.Error(err)
	}

	if ut.Sig != "foo" {
		t.Error("strings 'sig' and '\\u0073ig' should be equal")
	}

	raw = []byte(`{"si\u0047":"bar"}`)
	var ut2 UnicodeTest
	err = json.Unmarshal(raw, &ut2)
	if err != nil {
		t.Error(err)
	}

	if ut2.Sig != "" {
		t.Error("strings 'sig' and 'si\\u0047' should not be equal")
	}
}

func TestCaseSensitiveJSON(t *testing.T) {
	raw := []byte(`{"test":42}`)
	var cs CaseSensitive
	err := json.Unmarshal(raw, &cs)
	if err != nil {
		t.Error(err)
	}

	if cs.A != 0 || cs.B != 42 || cs.C != 0 {
		t.Errorf("parsing JSON should be case-sensitive (got %v)", cs)
	}
}

func TestErrorOnTrailingCharacters(t *testing.T) {
	raw := []byte(`{"test":42}asdf`)
	var m map[string]interface{}
	err := json.Unmarshal(raw, &m)
	if err == nil {
		t.Error("json.Unmarshal should fail if string has trailing chars")
	}
}

func TestRejectDuplicateKeysObject(t *testing.T) {
	raw := []byte(`{"test":42,"test":43}`)
	var cs CaseSensitive
	err := json.Unmarshal(raw, &cs)
	if err == nil {
		t.Error("should reject JSON with duplicate keys, but didn't")
	}
}

func TestRejectDuplicateKeysInterface(t *testing.T) {
	raw := []byte(`{"test":42,"test":43}`)
	var m interface{}
	err := json.Unmarshal(raw, &m)
	if err == nil {
		t.Error("should reject JSON with duplicate keys, but didn't")
	}
}

func TestParseCaseSensitiveJWE(t *testing.T) {
	invalidJWE := `{"protected":"eyJlbmMiOiJYWVoiLCJBTEciOiJYWVoifQo","encrypted_key":"QUJD","iv":"QUJD","ciphertext":"QUJD","tag":"QUJD"}`
	_, err := ParseEncrypted(invalidJWE)
	if err == nil {
		t.Error("Able to parse message with case-invalid headers", invalidJWE)
	}
}

func TestParseCaseSensitiveJWS(t *testing.T) {
	invalidJWS := `{"PAYLOAD":"CUJD","signatures":[{"protected":"e30","signature":"CUJD"}]}`
	_, err := ParseSigned(invalidJWS)
	if err == nil {
		t.Error("Able to parse message with case-invalid headers", invalidJWS)
	}
}
