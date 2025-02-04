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

// Table represents table info.
type Table struct {
	Type      string // type
	TableName string // table_name
	ManualPk  bool   // manual_pk
}

// Column represents column info.
type Column struct {
	FieldOrdinal         int    // field_ordinal
	ColumnName           string // column_name
	DataType             string // data_type
	NotNull              bool   // not_null
	IsPrimaryKey         bool   // is_primary_key
	IsGenerated          bool   // is_generated
	AllowCommitTimestamp bool   // allow_commit_timestamp
}

// Index represents an index.
type Index struct {
	IndexName string // index_name
	IsUnique  bool   // is_unique
	IsPrimary bool   // is_primary
	SeqNo     int    // seq_no
	Origin    string // origin
	IsPartial bool   // is_partial
}

// IndexColumn represents index column info.
type IndexColumn struct {
	SeqNo      int    // seq_no. If is'a Storing Column, this value is 0.
	ColumnName string // column_name
	Storing    bool   // storing column or not
}

// CustomTypes represents custom type definitions
type CustomTypes struct {
	Tables []struct {
		Name    string            `yaml:"name"`
		Columns map[string]string `yaml:"columns"`
	}
}
