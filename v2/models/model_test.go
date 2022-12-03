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

package models

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPackage_LocalName(t *testing.T) {
	t.Parallel()

	table := []struct {
		name     string
		pkg      Package
		expected string
	}{
		{
			name:     "without package name or alias",
			pkg:      Package{Path: "go.mercari.io/test"},
			expected: "test",
		},
		{
			name:     "with package name",
			pkg:      Package{Name: "pkgname", Path: "go.mercari.io/test"},
			expected: "pkgname",
		},
		{
			name:     "with alias",
			pkg:      Package{Alias: "pkgalias", Path: "go.mercari.io/test"},
			expected: "pkgalias",
		},
		{
			name:     "with package name and alias",
			pkg:      Package{Alias: "pkgalias", Name: "pkgname", Path: "go.mercari.io/test"},
			expected: "pkgalias",
		},
		{
			name:     "standard library",
			pkg:      Package{Path: "fmt"},
			expected: "fmt",
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pkg.LocalName()
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Fatalf("%s (-got, +want)\n%s", got, diff)
			}
		})
	}
}

func TestPackage_Standard(t *testing.T) {
	t.Parallel()

	table := []struct {
		name     string
		pkg      Package
		expected bool
	}{
		{
			name:     "standard library",
			pkg:      Package{Path: "fmt"},
			expected: true,
		},
		{
			name:     "non-standard library",
			pkg:      Package{Path: "go.mercari.io/test"},
			expected: false,
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pkg.Standard()
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Fatalf("%v (-got, +want)\n%s", got, diff)
			}
		})
	}
}

func TestPackage_String(t *testing.T) {
	t.Parallel()

	table := []struct {
		name     string
		pkg      Package
		expected string
	}{
		{
			name:     "without package name or alias",
			pkg:      Package{Path: "go.mercari.io/test"},
			expected: `"go.mercari.io/test"`,
		},
		{
			name:     "with package name",
			pkg:      Package{Name: "pkgname", Path: "go.mercari.io/test"},
			expected: `"go.mercari.io/test"`,
		},
		{
			name:     "with alias",
			pkg:      Package{Alias: "pkgalias", Path: "go.mercari.io/test"},
			expected: `pkgalias "go.mercari.io/test"`,
		},
		{
			name:     "with package name and alias",
			pkg:      Package{Alias: "pkgalias", Name: "pkgname", Path: "go.mercari.io/test"},
			expected: `pkgalias "go.mercari.io/test"`,
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.pkg.String()
			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Fatalf("%s (-got, +want)\n%s", got, diff)
			}
		})
	}
}
