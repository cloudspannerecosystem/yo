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

package loader

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/models"
)

var (
	simpleSchema = `
CREATE TABLE Simple (
  Id INT64 NOT NULL,
  Value STRING(32) NOT NULL,
) PRIMARY KEY(Id);
CREATE INDEX SimpleIndex ON Simple(Value);
CREATE UNIQUE INDEX SimpleIndex2 ON Simple(Id, Value)`
	interleaveSchema = `
CREATE TABLE Parent (
  Id INT64 NOT NULL,
) PRIMARY KEY(Id);

CREATE TABLE Interleaved (
  InterleavedId INT64 NOT NULL,
  Id INT64 NOT NULL,
  Value INT64 NOT NULL,
) PRIMARY KEY(Id, InterleavedId),
INTERLEAVE IN PARENT Parent;
CREATE INDEX InterleavedKey ON Interleaved(Id, Value), INTERLEAVE IN Parent
`

	oooSchema = `
	CREATE TABLE OutOfOrderPrimaryKeys (
  PKey1 STRING(32) NOT NULL,
  PKey2 STRING(32) NOT NULL,
  PKey3 STRING(32) NOT NULL,
) PRIMARY KEY(PKey2, PKey1, PKey3);
`

	maxLengthSchema = `
CREATE TABLE MaxLengths (
  MaxString STRING(MAX) NOT NULL,
  MaxBytes BYTES(MAX) NOT NULL,
) PRIMARY KEY(MaxString);
`
)

func TestLoader(t *testing.T) {
	dir := t.TempDir()

	table := []struct {
		name           string
		opt            Option
		schema         string
		expectedSchema *internal.Schema
	}{
		{
			name:   "Simple",
			opt:    Option{},
			schema: simpleSchema,
			expectedSchema: &internal.Schema{
				Types: []*internal.Type{
					{
						Name: "Simple",
						PrimaryKeyFields: []*internal.Field{
							{
								Name:    "ID",
								Type:    "int64",
								NilType: "0",
								Len:     -1,
							},
						},
						Fields: []*internal.Field{
							{
								Name:    "ID",
								Type:    "int64",
								NilType: "0",
								Len:     -1,
							},
							{
								Name:    "Value",
								Type:    "string",
								NilType: `""`,
								Len:     32,
							},
						},
						Table: &models.Table{TableName: "Simple"},
						Indexes: []*internal.Index{
							{
								Name:           "SimpleIndex",
								FuncName:       "SimplesBySimpleIndex",
								LegacyFuncName: "SimplesByValue",
								Fields: []*internal.Field{
									{Name: "Value", Type: "string", NilType: `""`, Len: 32},
								},
								Index: &models.Index{IndexName: "SimpleIndex"},
							},
							{
								Name:           "SimpleIndex2",
								FuncName:       "SimpleBySimpleIndex2",
								LegacyFuncName: "SimpleByIDValue",
								Fields: []*internal.Field{
									{Name: "ID", Type: "int64", NilType: "0", Len: -1},
									{Name: "Value", Type: "string", NilType: `""`, Len: 32},
								},
								Index: &models.Index{IndexName: "SimpleIndex2", IsUnique: true},
							},
						},
					},
				},
			},
		},
		{
			name:   "Interleave",
			opt:    Option{},
			schema: interleaveSchema,
			expectedSchema: &internal.Schema{
				Types: []*internal.Type{
					{
						Name: "Interleaved",
						PrimaryKeyFields: []*internal.Field{
							{Name: "ID", Type: "int64", NilType: "0", Len: -1},
							{Name: "InterleavedID", Type: "int64", NilType: "0", Len: -1},
						},
						Fields: []*internal.Field{
							{Name: "InterleavedID", Type: "int64", NilType: "0", Len: -1},
							{Name: "ID", Type: "int64", NilType: "0", Len: -1},
							{Name: "Value", Type: "int64", NilType: "0", Len: -1},
						},
						Table: &models.Table{TableName: "Interleaved", ParentTableName: "Parent"},
						Indexes: []*internal.Index{
							{
								Name:           "InterleavedKey",
								FuncName:       "InterleavedsByInterleavedKey",
								LegacyFuncName: "InterleavedsByIDValue",
								Fields: []*internal.Field{
									{Name: "ID", Type: "int64", NilType: "0", Len: -1},
									{Name: "Value", Type: "int64", NilType: "0", Len: -1},
								},
								Index: &models.Index{IndexName: "InterleavedKey"},
							},
						},
					},
					{
						Name: "Parent",
						PrimaryKeyFields: []*internal.Field{
							{Name: "ID", Type: "int64", NilType: "0", Len: -1},
						},
						Fields: []*internal.Field{
							{Name: "ID", Type: "int64", NilType: "0", Len: -1},
						},
						Table: &models.Table{TableName: "Parent"},
					},
				},
			},
		},
		{
			name:   "OutOfOrderPrimaryKey",
			opt:    Option{},
			schema: oooSchema,
			expectedSchema: &internal.Schema{
				Types: []*internal.Type{
					{
						Name: "OutOfOrderPrimaryKey",
						PrimaryKeyFields: []*internal.Field{
							{Name: "PKey2", Type: "string", NilType: `""`, Len: 32},
							{Name: "PKey1", Type: "string", NilType: `""`, Len: 32},
							{Name: "PKey3", Type: "string", NilType: `""`, Len: 32},
						},
						Fields: []*internal.Field{
							{Name: "PKey1", Type: "string", NilType: `""`, Len: 32},
							{Name: "PKey2", Type: "string", NilType: `""`, Len: 32},
							{Name: "PKey3", Type: "string", NilType: `""`, Len: 32},
						},
						Table: &models.Table{TableName: "OutOfOrderPrimaryKeys"},
					},
				},
			},
		},
		{
			name:   "MaxLength",
			opt:    Option{},
			schema: maxLengthSchema,
			expectedSchema: &internal.Schema{
				Types: []*internal.Type{
					{
						Name: "MaxLength",
						PrimaryKeyFields: []*internal.Field{
							{Name: "MaxString", Type: "string", NilType: `""`, Len: -1},
						},
						Fields: []*internal.Field{
							{Name: "MaxString", Type: "string", NilType: `""`, Len: -1},
							{Name: "MaxBytes", Type: "[]byte", NilType: "nil", Len: -1},
						},
						Table: &models.Table{TableName: "MaxLengths"},
					},
				},
			},
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			f, err := ioutil.TempFile(dir, "")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			_, _ = f.Write([]byte(tc.schema))
			_ = f.Close()
			path := f.Name()

			source, err := NewSchemaParserSource(path)
			if err != nil {
				t.Fatalf("failed to create schema parser source: %v", err)
			}

			inflector, err := internal.NewInflector(nil)
			if err != nil {
				t.Fatalf("failed to create inflector: %v", err)
			}

			l := NewTypeLoader(source, inflector, tc.opt)
			schema, err := l.LoadSchema()
			if err != nil {
				t.Fatalf("failed to load schema: %v", err)
			}

			if diff := cmp.Diff(schema, tc.expectedSchema,
				cmpopts.IgnoreFields(internal.Field{}, "Col"),
				cmpopts.IgnoreFields(internal.Index{}, "Type"),
			); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}
}

func Test_setIndexesToTables(t *testing.T) {
	tests := []struct {
		table  map[string]*internal.Type
		ix     map[string]*internal.Index
		result map[string]int
	}{
		{
			table: map[string]*internal.Type{
				"TableA": &internal.Type{
					Indexes: []*internal.Index{},
				},
			},
			ix: map[string]*internal.Index{
				"TableA_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableA_Index2": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
			},
			result: map[string]int{
				"TableA": 2,
			},
		},
		{
			table: map[string]*internal.Type{
				"TableA": &internal.Type{
					Indexes: []*internal.Index{},
				},
				"TableB": &internal.Type{
					Indexes: []*internal.Index{},
				},
			},
			ix: map[string]*internal.Index{
				"TableA_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableA_Index2": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
			},
			result: map[string]int{
				"TableA": 2,
				"TableB": 0,
			},
		},
		{
			table: map[string]*internal.Type{
				"TableA": &internal.Type{
					Indexes: []*internal.Index{},
				},
				"TableB": &internal.Type{
					Indexes: []*internal.Index{},
				},
			},
			ix: map[string]*internal.Index{
				"TableA_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableA_Index2": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
				"TableB_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableB"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableB_Index2forTableA_Hoge": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableB"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
			},
			result: map[string]int{
				"TableA": 2,
				"TableB": 2,
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case:%d", i), func(t *testing.T) {
			setIndexesToTables(tt.table, tt.ix)
			for k, v := range tt.table {
				if len(v.Indexes) != tt.result[k] {
					t.Errorf("error. want:%d got:%d", tt.result[k], len(v.Indexes))
				}
			}
		})
	}
}
