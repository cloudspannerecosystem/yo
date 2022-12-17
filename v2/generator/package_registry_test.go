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
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mercari.io/yo/v2/models"
)

func TestPackageRegistry_Use(t *testing.T) {
	t.Parallel()

	local := models.Package{Name: "v2", Path: "go.mercari.io/yo/v2"}
	initialPkgs := []models.Package{
		{Name: "p1", Path: "go.mercari.io/yo/v2/p1"},
		{Name: "p2", Path: "go.mercari.io/yo/v2/p2"},
		{Name: "p3", Path: "go.mercari.io/yo/v2/p3"},
	}

	table := []struct {
		desc     string
		pkg      models.Package
		name     string
		expected string
	}{
		{
			desc:     "local package",
			pkg:      local,
			name:     "Test{}",
			expected: "Test{}",
		},
		{
			desc:     "builtin package",
			pkg:      models.BuiltInPackage,
			name:     "int64",
			expected: "int64",
		},
		{
			desc:     "package is already used before",
			pkg:      models.Package{Name: "p1", Path: "go.mercari.io/yo/v2/p1"},
			name:     "Test{}",
			expected: "p1.Test{}",
		},
		{
			desc:     "package name conflicts with the go reserved names",
			pkg:      models.Package{Name: "var", Path: "go.mercari.io/yo/v2/var"},
			name:     "Test{}",
			expected: "_var.Test{}",
		},
		{
			desc:     "package name is already used",
			pkg:      models.Package{Name: "p1", Path: "github.com/mercari/repo/p1"},
			name:     "Test{}",
			expected: "p11.Test{}",
		},
	}

	for _, tc := range table {
		t.Run(tc.desc, func(t *testing.T) {
			registry := NewPackageRegistry(local)
			for _, pkg := range initialPkgs {
				registry.packageNames[pkg] = pkg.Name
				registry.usedPackageNames[pkg.Name] = struct{}{}
			}

			got := registry.Use(tc.pkg, tc.name)

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Fatalf("%s (-got, +want)\n%s", tc.expected, diff)
			}
		})
	}
}

func TestPackageRegistry_GetImports(t *testing.T) {
	t.Parallel()

	table := []struct {
		name         string
		packageNames map[models.Package]string
		expected     []string
	}{
		{
			name: "packages without alias",
			packageNames: map[models.Package]string{
				{Name: "p1", Path: "go.mercari.io/yo/v2/p1"}: "p1",
				{Name: "p2", Path: "go.mercari.io/yo/v2/p2"}: "p2",
				{Name: "p3", Path: "go.mercari.io/yo/v2/p3"}: "p3",
			},
			expected: []string{
				`"go.mercari.io/yo/v2/p1"`,
				`"go.mercari.io/yo/v2/p2"`,
				`"go.mercari.io/yo/v2/p3"`,
			},
		},
		{
			name: "packages with alias",
			packageNames: map[models.Package]string{
				{Name: "p1", Path: "go.mercari.io/yo/v2/p1"}:     "p1",
				{Name: "p2", Path: "go.mercari.io/yo/v2/p2"}:     "p2",
				{Name: "p1", Path: "github.com/mercari/repo/p1"}: "p11",
			},
			expected: []string{
				`"go.mercari.io/yo/v2/p1"`,
				`"go.mercari.io/yo/v2/p2"`,
				`p11 "github.com/mercari/repo/p1"`,
			},
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			registry := PackageRegistry{packageNames: tc.packageNames}

			if diff := cmp.Diff(registry.GetImports(), tc.expected); diff != "" {
				t.Fatalf("%s (-got, +want)\n%s", tc.expected, diff)
			}
		})
	}
}
