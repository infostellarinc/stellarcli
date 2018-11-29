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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Normalizer", func() {

	It("trims", func() {
		Expect(Normalize(" ")).To(Equal(""))
		Expect(Normalize(" a")).To(Equal("a"))
		Expect(Normalize("b ")).To(Equal("b"))
		Expect(Normalize(" abc ")).To(Equal("abc"))
	})

	It("heredoc", func() {
		Expect(Normalize(Normalize(`sentence1
			sentence2
			sentence3
		`))).To(Equal("sentence1\nsentence2\nsentence3"))

		Expect(Normalize(Normalize(`
			sentence1
			sentence2
			sentence3
		`))).To(Equal("sentence1\nsentence2\nsentence3"))
	})

	It("Normalize", func() {
		actual := Normalize(`  sentence1
			sentence2
			sentence3
		
		`)

		Expect(Normalize(actual)).To(Equal("sentence1\nsentence2\nsentence3"))
	})

	It("ToLower", func() {
		actual := Normalize("  HeLLO  ")

		Expect(ToLower(actual)).To(Equal("hello"))
	})
})
