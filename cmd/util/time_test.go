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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Time", func() {
	Context("when passed a date", func() {
		It("should succeed", func() {
			expected := time.Date(2018, 11, 30, 0, 0, 0, 0, time.UTC)
			Expect(ParseDateTime("20181130")).Should(Equal(expected))
			Expect(ParseDateTime("2018-11-30")).Should(Equal(expected))
			Expect(ParseDateTime("2018/11/30")).Should(Equal(expected))
		})

		It("should fail", func() {
			var err error

			_, err = ParseDateTime("2018113")
			Expect(err).Should(HaveOccurred())
			_, err = ParseDateTime("2018-11-30T")
			Expect(err).Should(HaveOccurred())
			_, err = ParseDateTime("2018@11@30")
			Expect(err).Should(HaveOccurred())
			_, err = ParseDateTime("")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when passed a date and time", func() {
		It("should succeed", func() {
			expected := time.Date(2018, 11, 30, 15, 07, 30, 0, time.UTC)
			Expect(ParseDateTime("20181130150730")).Should(Equal(expected))
			Expect(ParseDateTime("2018-11-30T15:07:30")).Should(Equal(expected))
			Expect(ParseDateTime("2018-11-30 15:07:30")).Should(Equal(expected))
			Expect(ParseDateTime("2018/11/30T15:07:30")).Should(Equal(expected))
			Expect(ParseDateTime("2018/11/30 15:07:30")).Should(Equal(expected))
		})

		It("should fail", func() {
			var err error

			_, err = ParseDateTime("201811301507")
			Expect(err).Should(HaveOccurred())
			_, err = ParseDateTime("2018-11-30T15:07:30ZZ")
			Expect(err).Should(HaveOccurred())
			_, err = ParseDateTime("2018@11@3015:07:30")
			Expect(err).Should(HaveOccurred())
			_, err = ParseDateTime("2018-11-3015:07:30.324")
			Expect(err).Should(HaveOccurred())
			_, err = ParseDateTime("")
			Expect(err).Should(HaveOccurred())
		})
	})
})
