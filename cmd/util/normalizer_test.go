// Copyright © 2018 Infostellar, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test if leading and trailing spaces are removed by Normalize().
func TestTrimming(t *testing.T) {
	actual := Normalize(" abc ")

	expected := "abc"
	assert.Equal(t, expected, actual)
}

// Test if indentations in raw string literal are removed by Normalize().
func TestHeredoc(t *testing.T) {
	actual := Normalize(`sentence1
		sentence2
		sentence3
	`)
	expected := "sentence1\nsentence2\nsentence3"

	assert.Equal(t, expected, actual)
}

// Test if trim and heredoc is applied to the given string.
func TestNormalize(t *testing.T) {
	actual := Normalize(`  sentence1
		sentence2
		sentence3
    
	`)
	expected := "sentence1\nsentence2\nsentence3"

	assert.Equal(t, expected, actual)
}
