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
	"container/heap"
	"sync"
)

// An internal type of PriorityQueue that implements heap.Interface.
type pqInternal struct {
	items []interface{}
	less  func(i, j interface{}) bool

	mu sync.Mutex
}

// A PriorityQueue that returns items ordered by 'less' function.
// This type IS NOT thread safe.
type PriorityQueue struct {
	queue *pqInternal
}

func (pq *pqInternal) Len() int { return len(pq.items) }

func (pq *pqInternal) Less(i, j int) bool {
	return pq.less(pq.items[i], pq.items[j])
}

func (pq *pqInternal) Pop() interface{} {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[0 : n-1]
	return item
}

func (pq *pqInternal) Push(item interface{}) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if item == nil {
		panic("cannot push nil.")
	}

	pq.items = append(pq.items, item)
}

func (pq *pqInternal) Swap(i, j int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *PriorityQueue) Len() int { return pq.queue.Len() }

func (pq *PriorityQueue) Pop() interface{} {
	return heap.Pop(pq.queue)
}

func (pq *PriorityQueue) Push(item interface{}) {
	heap.Push(pq.queue, item)
}

func NewPriorityQueue(itemExample interface{}, less func(i, j interface{}) bool) *PriorityQueue {
	return &PriorityQueue{
		queue: &pqInternal{items: make([]interface{}, 0), less: less},
	}
}
