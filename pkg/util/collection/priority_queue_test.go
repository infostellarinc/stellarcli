//
// Copyright © 2019 Infostellar, Inc.
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

package collection

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Item struct {
	t time.Time
}

var _ = Describe("PriorityQueue", func() {
	Context("when called without any items", func() {
		pq := NewPriorityQueue((*int)(nil), func(i, j interface{}) bool {
			int1 := i.(*int)
			int2 := i.(*int)
			return *int1 < *int2
		})

		It("Len() should return", func() {
			Expect(pq.Len()).Should(Equal(0))
		})
		It("Pop() should panic", func() {
			Expect(func() {
				pq.Pop()
			}).To(Panic())
		})
	})

	Context("when passed items", func() {
		pq := NewPriorityQueue((*Item)(nil), func(i, j interface{}) bool {
			item1 := i.(*Item)
			item2 := j.(*Item)

			return item1.t.Before(item2.t)
		})

		It("should succeed", func() {
			rand.Seed(time.Now().UnixNano())

			n := 50
			for i := 0; i < n; i++ {
				nsec := rand.Intn(1000)
				seconds := rand.Intn(30)
				item := &Item{t: time.Date(2019, 10, 01, 3, 16, seconds, nsec, time.UTC)}
				pq.Push(item)
			}

			items := make([]*Item, n)
			for i := 0; i < n; i++ {
				Expect(pq.Len()).Should(Equal(n - i))
				item := pq.Pop().(*Item)
				items[i] = item
			}

			for i := 1; i < n; i++ {
				// Test if items[i-1].t is equal or before items[i]
				Expect(!items[i-1].t.After(items[i].t)).Should(BeTrue())
			}
		})
	})

	Context("when nil is passed to Push()", func() {
		pq := NewPriorityQueue((*Item)(nil), func(i, j interface{}) bool {
			item1 := i.(*Item)
			item2 := j.(*Item)

			return item1.t.Before(item2.t)
		})

		It("should panic", func() {
			Expect(func() {
				pq.Push(nil)
			}).To(Panic())
		})
	})
})
