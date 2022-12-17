package models

import "fmt"

const GoNil = "nil"

type PackageResolver interface {
	Use(pkg Package, name string) string
}

type FieldType interface {
	GetType(resolver PackageResolver) string
	GetOriginalType(resolver PackageResolver) string
	GetNullValue(resolver PackageResolver) string
}

type PlainFieldType struct {
	Pkg        Package
	Type       string
	CustomType string
	NullValue  string
}

func (t PlainFieldType) GetType(resolver PackageResolver) string {
	if t.CustomType != "" {
		return t.CustomType
	}

	return t.GetOriginalType(resolver)
}

func (t PlainFieldType) GetOriginalType(resolver PackageResolver) string {
	return resolver.Use(t.Pkg, t.Type)
}

func (t PlainFieldType) GetNullValue(resolver PackageResolver) string {
	return resolver.Use(t.Pkg, t.NullValue)
}

type ArrayFieldType struct {
	Nullable   bool
	CustomType string
	Element    FieldType
}

func (t ArrayFieldType) GetType(resolver PackageResolver) string {
	if t.CustomType != "" {
		return t.CustomType
	}

	return t.GetOriginalType(resolver)
}

func (t ArrayFieldType) GetOriginalType(resolver PackageResolver) string {
	return fmt.Sprintf("[]%s", t.Element.GetType(resolver))
}

func (t ArrayFieldType) GetNullValue(resolver PackageResolver) string {
	if t.Nullable {
		return GoNil
	}

	return fmt.Sprintf("%s{}", t.GetType(resolver))
}
