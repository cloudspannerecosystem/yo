package internal

import "go.mercari.io/yo/models"

// Field contains field information.
type Field struct {
	Name       string
	Type       string
	CustomType string
	NilType    string
	Len        int
	Col        *models.Column
}

// Type is a template item for a type.
type Type struct {
	Name             string
	Schema           string
	PrimaryKey       *Field
	PrimaryKeyFields []*Field
	Fields           []*Field
	Table            *models.Table
}

// Index is a template item for a index into a table.
type Index struct {
	FuncName string
	Schema   string
	Type     *Type
	Fields   []*Field
	Index    *models.Index
}
