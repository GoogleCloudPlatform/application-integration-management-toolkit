// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiclient

import (
	"fmt"
	"runtime"
)

func newError(message string, err error) error {
	pc := make([]uintptr, 1) // Program counter values
	runtime.Callers(2, pc)   // Skip runtime.Callers and newError itself
	frame, _ := runtime.CallersFrames(pc).Next()

	// Construct the error message with package, function, and message.
	// frame.Function gives the full path, including package.
	if err != nil {
		return fmt.Errorf("%s:%d: %s: %w", frame.Function, frame.Line, message, err)
	} else {
		return fmt.Errorf("%s:%d: %s", frame.Function, frame.Line, message)
	}
}
