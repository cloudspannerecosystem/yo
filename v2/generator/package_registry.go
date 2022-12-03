package generator

import (
	"fmt"
	"sort"

	"go.mercari.io/yo/v2/models"
)

var defaultTypeModulePackages = []*models.Package{
	{Path: "context"},
	{Path: "fmt"},
	{Path: "strings"},
	{Path: "time"},

	{Path: "cloud.google.com/go/spanner"},
	{Path: "google.golang.org/api/iterator"},
	{Path: "google.golang.org/grpc/codes"},
}

var defaultGlobalModulePackages = []*models.Package{
	{Path: "context"},
	{Path: "errors"},
	{Path: "fmt"},
	{Path: "strconv"},

	{Path: "cloud.google.com/go/spanner"},
	{Path: "github.com/googleapis/gax-go/v2/apierror"},
	{Path: "google.golang.org/grpc/codes"},
	{Path: "google.golang.org/grpc/status"},
}

// NewTypePackageRegistry creates a new instance of PackageRegistry for a specific type module
func NewTypePackageRegistry(t *models.Type) (*PackageRegistry, error) {
	r, err := newPackageRegistry(defaultTypeModulePackages)
	if err != nil {
		return nil, err
	}

	for _, f := range t.Fields {
		if f.Package == nil {
			continue
		}

		if err := r.Register(f.Package); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// NewGlobalPackageRegistry creates a new instance of PackageRegistry for a global module
func NewGlobalPackageRegistry() (*PackageRegistry, error) {
	r, err := newPackageRegistry(defaultGlobalModulePackages)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func newPackageRegistry(defaults []*models.Package) (*PackageRegistry, error) {
	r := &PackageRegistry{
		pathToLocalName: map[string]string{},
		localNameToPath: map[string]string{},
	}
	for _, p := range defaults {
		if err := r.Register(p); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// PackageRegistry manages Go packages imported in a generated file
type PackageRegistry struct {
	packages []*models.Package
	// pathToLocalName maps a package import path to its local name
	pathToLocalName map[string]string
	// localNameToPath maps a package local name to its import path
	localNameToPath map[string]string
}

// Register registers Go packages imported in a given file
func (r *PackageRegistry) Register(p *models.Package) error {
	path, localNameRegistered := r.localNameToPath[p.LocalName()]
	if localNameRegistered && path != p.Path {
		return fmt.Errorf("using the same local name %q for different packages: %q, %q", p.LocalName(), p.Path, path)
	}

	localName, pathRegistered := r.pathToLocalName[p.Path]
	if pathRegistered && localName != p.LocalName() {
		return fmt.Errorf("importing %q package with different local names: %q, %q", p.Path, p.LocalName(), localName)
	}

	if !pathRegistered {
		r.packages = append(r.packages, p)
		r.pathToLocalName[p.Path] = p.LocalName()
		r.localNameToPath[p.LocalName()] = p.Path
	}

	return nil
}

// GetImports returns a list of import statements
func (r *PackageRegistry) GetImports() []string {
	sort.Slice(r.packages, func(i, j int) bool {
		// Standard libraries come first, sorted lexicographically
		if r.packages[i].Standard() == r.packages[j].Standard() {
			return r.packages[i].Path < r.packages[j].Path
		}

		return r.packages[i].Standard()
	})

	var imports []string
	for _, p := range r.packages {
		imports = append(imports, p.String())
	}

	return imports
}
