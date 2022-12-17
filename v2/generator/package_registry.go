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
	"strconv"

	"go.mercari.io/yo/v2/models"
)

func NewPackageRegistry(local models.Package) *PackageRegistry {
	return &PackageRegistry{
		localPackage:     local,
		packageNames:     map[models.Package]string{},
		usedPackageNames: map[string]struct{}{},
		manualImports:    map[models.Package]struct{}{},
	}
}

// PackageRegistry manages Go packages imported in a generated file
type PackageRegistry struct {
	localPackage     models.Package
	packageNames     map[models.Package]string
	usedPackageNames map[string]struct{}
	manualImports    map[models.Package]struct{}
}

// Use registers Go packages imported in a given file
func (r *PackageRegistry) Use(pkg models.Package, name string) string {
	if pkg == r.localPackage || pkg == models.BuiltInPackage {
		return name
	}
	if packageName, ok := r.packageNames[pkg]; ok {
		return fmt.Sprintf("%s.%s", packageName, name)
	}

	packageName := pkg.Name
	if _, reserved := goReservedNames[packageName]; reserved {
		// Prepend '_' if the package name conflict with a Go keyword
		packageName = "_" + packageName
	}

	i := 1
	original := packageName
	for {
		if _, used := r.usedPackageNames[packageName]; !used {
			break
		}
		packageName = original + strconv.Itoa(i)
		i += 1
	}
	r.packageNames[pkg] = packageName
	r.usedPackageNames[packageName] = struct{}{}
	return fmt.Sprintf("%s.%s", packageName, name)
}

// GetImports returns a list of import statements
func (r *PackageRegistry) GetImports() []string {
	var imports []string
	for pkg, name := range r.packageNames {
		if name != pkg.Name {
			// Need to set an alias since the name is not equal to the package name
			imports = append(imports, fmt.Sprintf("%s %q", name, pkg.Path))
			continue
		}

		imports = append(imports, strconv.Quote(pkg.Path))
	}

	return imports
}
