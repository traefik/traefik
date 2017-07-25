/*
Copyright 2017 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package spanner

import (
	"reflect"
	"testing"

	proto3 "github.com/golang/protobuf/ptypes/struct"

	sppb "google.golang.org/genproto/googleapis/spanner/v1"
)

// Test Statement.bindParams.
func TestBindParams(t *testing.T) {
	// Verify Statement.bindParams generates correct values and types.
	want := sppb.ExecuteSqlRequest{
		Params: &proto3.Struct{
			Fields: map[string]*proto3.Value{
				"var1": stringProto("abc"),
				"var2": intProto(1),
			},
		},
		ParamTypes: map[string]*sppb.Type{
			"var1": stringType(),
			"var2": intType(),
		},
	}
	st := Statement{
		SQL:    "SELECT id from t_foo WHERE col1 = @var1 AND col2 = @var2",
		Params: map[string]interface{}{"var1": "abc", "var2": int64(1)},
	}
	got := sppb.ExecuteSqlRequest{}
	if err := st.bindParams(&got); err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("bind result: \n(%v, %v)\nwant\n(%v, %v)\n", got, err, want, nil)
	}
	// Verify type error reporting.
	st.Params["var2"] = struct{}{}
	wantErr := errBindParam("var2", struct{}{}, errEncoderUnsupportedType(struct{}{}))
	if err := st.bindParams(&got); !reflect.DeepEqual(err, wantErr) {
		t.Errorf("got unexpected error: %v, want: %v", err, wantErr)
	}
}

func TestNewStatement(t *testing.T) {
	s := NewStatement("query")
	if got, want := s.SQL, "query"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
