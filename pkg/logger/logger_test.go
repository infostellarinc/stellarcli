// Copyright © 2020 Infostellar, Inc.
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

package logger

// run with "go test -v" in this folder to see output

import (
	"fmt"
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	SetEmitRateMillis(500)
	for i := 0; i < 100; i++ {
		PrintlnThrottled("%d", i)
		time.Sleep(time.Duration(10) * time.Millisecond)
	}
	// manually make sure last message is printed
	fmt.Println("(99 should be printed below this line)")
	time.Sleep(time.Duration(600) * time.Millisecond)
}
