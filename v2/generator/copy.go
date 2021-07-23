// Copyright (c) 2020 Mercari, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"go.mercari.io/yo/v2/module/builtin"
)

// CopyDefaultTemplates copies default templete files to dir.
func CopyDefaultTemplates(dir string) error {
	for _, m := range builtin.All {
		if err := func() (err error) {
			filename := fmt.Sprintf("%s.go.tpl", m.Name())
			file, err := os.OpenFile(filepath.Join(dir, filename), os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				return err
			}
			defer func() {
				if cerr := file.Close(); err == nil {
					err = cerr
				}
			}()

			b, err := m.Load()
			if err != nil {
				return fmt.Errorf("failed to load builtin module %q: %v", m.Name(), err)
			}

			_, err = file.Write(b)
			return
		}(); err != nil {
			return err
		}
	}
	return nil
}
