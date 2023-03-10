/*
Copyright 2019 The Vitess Authors.

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

package trace

import (
	"io"

	"github.com/stackql/stackql-parser/go/vt/log"
)

// LogErrorsWhenClosing will close the provided Closer, and log any errors it generates
func LogErrorsWhenClosing(in io.Closer) func() {
	return func() {
		err := in.Close()
		if err != nil {
			log.Error(err)
		}
	}
}
