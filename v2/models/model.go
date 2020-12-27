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

// Schema contains information of all Go types.
type Schema struct {
	Types []*Type
}

// Type is a Go type that represents a Spanner table.
type Type struct {
	Name             string // Go like (CamelCase) table name
	PrimaryKeyFields []*Field
	Fields           []*Field
	Indexes          []*Index
	TableName        string
	Parent           *Type
}

// Field is a field of Go type that represents a Spanner column.
type Field struct {
	Name            string // Go like (CamelCase) field name
	Type            string // Go type specified by custom type or same to OriginalType below
	OriginalType    string // Go type corresponding to Spanner type
	NullValue       string // NULL value for Type
	Len             int    // Length for STRING, BYTES. -1 for MAX or other types
	ColumnName      string // column_name
	SpannerDataType string // data_type
	IsNotNull       bool   // not_null
	IsPrimaryKey    bool   // is_primary_key
}

// Index is a template item for a index into a table.
type Index struct {
	Name           string // Go like (CamelCase) index name
	FuncName       string // `By` + Name
	LegacyFuncName string // `By` + Type name + Field names
	Type           *Type
	Fields         []*Field
	StoringFields  []*Field
	NullableFields []*Field
	IndexName      string // index name
	IsUnique       bool   // the index is unique ro not
	IsPrimary      bool   // the index is primary key or not
}
