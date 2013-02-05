// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

// ErrorEqual compares Go errors for equality.
func ErrorEqual(left, right error) bool {
	if left == right {
		return true
	}

	if left != nil && right != nil {
		if left.Error() == right.Error() {
			return true
		}

		return false
	}

	return false
}