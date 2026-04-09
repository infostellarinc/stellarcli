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

package printer

import (
	"fmt"
	"strings"
)

type TemplateItem struct {
	Label string
	Path  string
}

// Flatten an object using the template.
func Flatten(obj map[string]interface{}, t []TemplateItem) []interface{} {
	flattened := make([]interface{}, len(t))
	for i, item := range t {
		v, err := GetValue(obj, item.Path)
		if err != nil {
			flattened[i] = ""
		} else {
			flattened[i] = v
		}
	}
	return flattened
}

// Converts TemplateItems to an array of label.
func GetLabels(items []TemplateItem) []interface{} {
	result := make([]interface{}, len(items))
	for i, item := range items {
		result[i] = item.Label
	}
	return result
}

// Get a value from the map by a path separated by dots.
// Example: gsInfo.gsLat
func GetValue(m map[string]interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")

	var current = m
	for _, key := range keys {
		v, ok := current[key]
		if !ok {
			return nil, fmt.Errorf("cannot find a value for the path, %s", path)
		}

		current, ok = v.(map[string]interface{})
		if !ok {
			return v, nil
		}
	}
	return nil, fmt.Errorf("cannot find a value for the path, %s", path)
}
