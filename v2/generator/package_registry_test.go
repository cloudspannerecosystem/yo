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

func TestNewTypePackageRegistry(t *testing.T) {
	t.Parallel()

	table := []struct {
		name            string
		typ             *models.Type
		expectedImports []string
		expectedErr     string
	}{
		{
			name: "no duplication with default imports",
			typ: &models.Type{
				Fields: []*models.Field{
					{Package: &models.Package{Path: "sort"}},
					{Package: &models.Package{Path: "go.mercari.io/test"}},
				},
			},
			expectedImports: []string{
				`"context"`,
				`"fmt"`,
				`"sort"`,
				`"strings"`,
				`"time"`,
				`"cloud.google.com/go/spanner"`,
				`"go.mercari.io/test"`,
				`"google.golang.org/api/iterator"`,
				`"google.golang.org/grpc/codes"`,
			},
		},
		{
			name: "duplicated package with default imports",
			typ: &models.Type{
				Fields: []*models.Field{
					{Package: &models.Package{Path: "context"}},
					{Package: &models.Package{Path: "cloud.google.com/go/spanner"}},
				},
			},
			expectedImports: []string{
				`"context"`,
				`"fmt"`,
				`"strings"`,
				`"time"`,
				`"cloud.google.com/go/spanner"`,
				`"google.golang.org/api/iterator"`,
				`"google.golang.org/grpc/codes"`,
			},
		},
		{
			name: "alias conflicts with default imports",
			typ: &models.Type{
				Fields: []*models.Field{
					{Package: &models.Package{Path: "cloud.google.com/go/spanner", Alias: "spn"}},
				},
			},
			expectedErr: `importing "cloud.google.com/go/spanner" package with different local names: "spn", "spanner"`,
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewTypePackageRegistry(tc.typ)
			if tc.expectedErr != "" {
				if err == nil {
					t.Fatal("expected to receive an error but got nil")
				}

				if diff := cmp.Diff(err.Error(), tc.expectedErr); diff != "" {
					t.Fatalf("%s (-got, +want)\n%s", tc.expectedErr, diff)
				}
			} else {
				if err != nil {
					t.Fatalf("received an unexpected error: %v", err)
				}

				if got == nil {
					t.Fatal("received a nil package registry")
				}

				if diff := cmp.Diff(got.GetImports(), tc.expectedImports); diff != "" {
					t.Fatalf("%s (-got, +want)\n%s", tc.expectedImports, diff)
				}
			}
		})
	}
}

func TestNewGlobalPackageRegistry(t *testing.T) {
	t.Parallel()

	got, err := NewGlobalPackageRegistry()
	if err != nil {
		t.Fatalf("received an unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("received a nil package registry")
	}

	expectedImports := []string{
		`"context"`,
		`"errors"`,
		`"fmt"`,
		`"strconv"`,
		`"cloud.google.com/go/spanner"`,
		`"github.com/googleapis/gax-go/v2/apierror"`,
		`"google.golang.org/grpc/codes"`,
		`"google.golang.org/grpc/status"`,
	}

	if diff := cmp.Diff(got.GetImports(), expectedImports); diff != "" {
		t.Fatalf("%s (-got, +want)\n%s", expectedImports, diff)
	}
}

func TestPackageRegistry_Register(t *testing.T) {
	t.Parallel()

	table := []struct {
		name            string
		initialPkgs     []*models.Package
		pkg             *models.Package
		expectedImports []string
		expectedErr     string
	}{
		{
			name: "plain package",
			initialPkgs: []*models.Package{
				{Path: "fmt"},
				{Path: "go.mercari.io/test"},
			},
			pkg: &models.Package{Path: "go.mercari.io/aaa"},
			expectedImports: []string{
				`"fmt"`,
				`"go.mercari.io/aaa"`,
				`"go.mercari.io/test"`,
			},
		},
		{
			name: "package with name",
			initialPkgs: []*models.Package{
				{Path: "fmt"},
				{Path: "go.mercari.io/test"},
			},
			pkg: &models.Package{Name: "testpkgname", Path: "go.mercari.io/aaa"},
			expectedImports: []string{
				`"fmt"`,
				`"go.mercari.io/aaa"`,
				`"go.mercari.io/test"`,
			},
		},
		{
			name: "package with alias",
			initialPkgs: []*models.Package{
				{Path: "fmt"},
				{Path: "go.mercari.io/test"},
			},
			pkg: &models.Package{Alias: "testpkgalias", Path: "go.mercari.io/aaa"},
			expectedImports: []string{
				`"fmt"`,
				`testpkgalias "go.mercari.io/aaa"`,
				`"go.mercari.io/test"`,
			},
		},
		{
			name: "duplicated packages",
			initialPkgs: []*models.Package{
				{Path: "fmt"},
				{Path: "go.mercari.io/test"},
			},
			pkg: &models.Package{Path: "go.mercari.io/test"},
			expectedImports: []string{
				`"fmt"`,
				`"go.mercari.io/test"`,
			},
		},
		{
			name: "the same package with different aliases",
			initialPkgs: []*models.Package{
				{Alias: "t1", Path: "go.mercari.io/test"},
			},
			pkg:         &models.Package{Alias: "t2", Path: "go.mercari.io/test"},
			expectedErr: `importing "go.mercari.io/test" package with different local names: "t2", "t1"`,
		},
		{
			name: "the same alias for different packages",
			initialPkgs: []*models.Package{
				{Alias: "testalias", Path: "go.mercari.io/test"},
			},
			pkg:         &models.Package{Alias: "testalias", Path: "test"},
			expectedErr: `using the same local name "testalias" for different packages: "test", "go.mercari.io/test"`,
		},
		{
			name: "the same package name for different packages",
			initialPkgs: []*models.Package{
				{Path: "go.mercari.io/test"},
			},
			pkg:         &models.Package{Path: "test"},
			expectedErr: `using the same local name "test" for different packages: "test", "go.mercari.io/test"`,
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			registry, err := newPackageRegistry(tc.initialPkgs)
			if err != nil {
				t.Fatalf("failed initializing pagkage registry: %v", err)
			}

			err = registry.Register(tc.pkg)

			if tc.expectedErr != "" {
				if err == nil {
					t.Fatal("expected to receive an error but got nil")
				}

				if diff := cmp.Diff(err.Error(), tc.expectedErr); diff != "" {
					t.Fatalf("%s (-got, +want)\n%s", tc.expectedErr, diff)
				}
			} else {
				if err != nil {
					t.Fatalf("received an unexpected error: %v", err)
				}

				if diff := cmp.Diff(registry.GetImports(), tc.expectedImports); diff != "" {
					t.Fatalf("%s (-got, +want)\n%s", tc.expectedImports, diff)
				}
			}
		})
	}
}

func TestPackageRegistry_GetImports(t *testing.T) {
	t.Parallel()

	table := []struct {
		name            string
		pkgs            []*models.Package
		expectedImports []string
	}{
		{
			name: "packages are unordered",
			pkgs: []*models.Package{
				{Path: "go.mercari.io/bbb"},
				{Path: "fmt"},
				{Path: "go.mercari.io/aaa"},
				{Path: "sort"},
			},
			expectedImports: []string{
				`"fmt"`,
				`"sort"`,
				`"go.mercari.io/aaa"`,
				`"go.mercari.io/bbb"`,
			},
		},
		{
			name: "packages with alias",
			pkgs: []*models.Package{
				{Path: "fmt"},
				{Alias: "srt", Path: "sort"},
				{Alias: "a", Path: "go.mercari.io/aaa"},
				{Alias: "b", Path: "go.mercari.io/bbb"},
			},
			expectedImports: []string{
				`"fmt"`,
				`srt "sort"`,
				`a "go.mercari.io/aaa"`,
				`b "go.mercari.io/bbb"`,
			},
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			registry, err := newPackageRegistry(tc.pkgs)
			if err != nil {
				t.Fatalf("failed initializing pagkage registry: %v", err)
			}

			if diff := cmp.Diff(registry.GetImports(), tc.expectedImports); diff != "" {
				t.Fatalf("%s (-got, +want)\n%s", tc.expectedImports, diff)
			}
		})
	}
}
