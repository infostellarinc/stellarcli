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

/*
Copyright 2016 The Kubernetes Authors.

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

package util

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
)

// Normalizer normalize a string by removing indentation in it and trim it.
type normalizer struct {
	string
}

// Remove indentation of the string, which may be added when defining it with raw string literals
func (s normalizer) heredoc() normalizer {
	s.string = heredoc.Doc(s.string)
	return s
}

// Remove leading and trailing spaces.
func (s normalizer) trim() normalizer {
	s.string = strings.TrimSpace(s.string)
	return s
}

// Normalize the string by removing indentation and trimming.
func Normalize(s string) string {
	return normalizer{s}.heredoc().trim().string
}
