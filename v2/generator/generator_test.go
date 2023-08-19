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
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/models"
)

type fakeLoader struct{}

func (*fakeLoader) NthParam(int) string {
	return "@"
}

func newTestGenerator(t *testing.T) *Generator {
	t.Helper()

	inflector, err := internal.NewInflector(nil)
	if err != nil {
		t.Fatalf("failed to create inflector: %v", err)
	}

	return NewGenerator(&fakeLoader{}, inflector, GeneratorOption{
		PackageName:    "yotest",
		Tags:           "",
		FilenameSuffix: ".yo.go",
		BaseDir:        t.TempDir(),
	})
}

func TestGenerator(t *testing.T) {
	table := []struct {
		name             string
		schema           *models.Schema
		expectedFilesDir string
		compareBaseFile  bool
	}{
		{
			name:             "BaseOnly",
			schema:           &models.Schema{},
			expectedFilesDir: "testdata/empty",
			compareBaseFile:  true,
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestGenerator(t)
			if err := g.Generate(tc.schema); err != nil {
				t.Fatalf("failed to generate: %v", err)
			}

			if err := filepath.Walk(g.baseDir, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}

				if !tc.compareBaseFile && info.Name() == "yo_db.yo.go" {
					return nil
				}

				t.Logf("generated file path: %v\n", path)

				actualContent, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}

				expectedFilePath := filepath.Join(tc.expectedFilesDir, info.Name())
				expectedContent, err := os.ReadFile(expectedFilePath)
				if os.IsNotExist(err) {
					err = os.MkdirAll(filepath.Join(tc.expectedFilesDir), 0766)
					if err != nil {
						t.Fatal(err)
					}
					err = os.WriteFile(expectedFilePath, actualContent, 0444)
					if err != nil {
						t.Fatal(err)
					}
					return nil
				} else if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(actualContent, expectedContent); diff != "" {
					t.Errorf("%s (-got, +want)\n%s", expectedFilePath, diff)
				}

				return nil
			}); err != nil {
				t.Fatalf("filepath.Walk failed: %v", err)
			}
		})
	}
}
