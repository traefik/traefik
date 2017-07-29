// Copyright 2015 xeipuuv ( https://github.com/xeipuuv )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author           janmentzel
// author-github    https://github.com/janmentzel
// author-mail      ? ( forward to xeipuuv@gmail.com )
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description     (Unit) Tests for utils ( Float / Integer conversion ).
//
// created          08-08-2013

package gojsonschema

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestResultErrorFormatNumber(t *testing.T) {

	assert.Equal(t, "1", resultErrorFormatNumber(1))
	assert.Equal(t, "-1", resultErrorFormatNumber(-1))
	assert.Equal(t, "0", resultErrorFormatNumber(0))
	// unfortunately, can not be recognized as float
	assert.Equal(t, "0", resultErrorFormatNumber(0.0))

	assert.Equal(t, "1.001", resultErrorFormatNumber(1.001))
	assert.Equal(t, "-1.001", resultErrorFormatNumber(-1.001))
	assert.Equal(t, "0.0001", resultErrorFormatNumber(0.0001))

	// casting math.MaxInt64 (1<<63 -1) to float back to int64
	// becomes negative. obviousely because of bit missinterpretation.
	// so simply test a slightly smaller "large" integer here
	assert.Equal(t, "4.611686018427388e+18", resultErrorFormatNumber(1<<62))
	// with negative int64 max works
	assert.Equal(t, "-9.223372036854776e+18", resultErrorFormatNumber(math.MinInt64))
	assert.Equal(t, "-4.611686018427388e+18", resultErrorFormatNumber(-1<<62))

	assert.Equal(t, "10000000000", resultErrorFormatNumber(1e10))
	assert.Equal(t, "-10000000000", resultErrorFormatNumber(-1e10))

	assert.Equal(t, "1.000000000001", resultErrorFormatNumber(1.000000000001))
	assert.Equal(t, "-1.000000000001", resultErrorFormatNumber(-1.000000000001))
	assert.Equal(t, "1e-10", resultErrorFormatNumber(1e-10))
	assert.Equal(t, "-1e-10", resultErrorFormatNumber(-1e-10))
	assert.Equal(t, "4.6116860184273876e+07", resultErrorFormatNumber(4.611686018427387904e7))
	assert.Equal(t, "-4.6116860184273876e+07", resultErrorFormatNumber(-4.611686018427387904e7))

}
