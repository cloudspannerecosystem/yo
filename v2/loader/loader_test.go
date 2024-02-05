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
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.mercari.io/yo/v2/config"
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

	alterTableAddFKSchema = `
ALTER TABLE Foo ADD FOREIGN KEY (CustomerID) REFERENCES Customers (CustomerID)

`

	alterTableAddConstraintFKSchema = `
ALTER TABLE Foo ADD CONSTRAINT FK_CustomerOrder FOREIGN KEY (CustomerID) REFERENCES Customers (CustomerID)
`
)

func TestLoader(t *testing.T) {
	table := []struct {
		name           string
		opt            Option
		schema         string
		expectedSchema *models.Schema
	}{
		{
			name:   "Simple",
			opt:    Option{},
			schema: simpleSchema,
			expectedSchema: &models.Schema{
				Types: []*models.Type{
					{
						Name: "Simple",
						PrimaryKeyFields: []*models.Field{
							{ColumnName: "Id"},
						},
						Fields: []*models.Field{
							{
								Name:            "ID",
								Type:            "int64",
								OriginalType:    "int64",
								NullValue:       "0",
								Len:             -1,
								ColumnName:      "Id",
								SpannerDataType: "INT64",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
							{
								Name:            "Value",
								Type:            "string",
								OriginalType:    "string",
								NullValue:       `""`,
								Len:             32,
								ColumnName:      "Value",
								SpannerDataType: "STRING(32)",
								IsNotNull:       true,
								IsPrimaryKey:    false,
							},
						},
						TableName: "Simple",
						Indexes: []*models.Index{
							{
								Name:           "SimpleIndex",
								FuncName:       "SimplesBySimpleIndex",
								LegacyFuncName: "SimplesByValue",
								Fields: []*models.Field{
									{ColumnName: "Value"},
								},
								IndexName: "SimpleIndex",
							},
							{
								Name:           "SimpleIndex2",
								FuncName:       "SimpleBySimpleIndex2",
								LegacyFuncName: "SimpleByIDValue",
								Fields: []*models.Field{
									{ColumnName: "Id"},
									{ColumnName: "Value"},
								},
								IndexName: "SimpleIndex2",
								IsUnique:  true,
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
			expectedSchema: &models.Schema{
				Types: []*models.Type{
					{
						Name: "Interleaved",
						PrimaryKeyFields: []*models.Field{
							{ColumnName: "Id"},
							{ColumnName: "InterleavedId"},
						},
						Fields: []*models.Field{
							{
								Name:            "InterleavedID",
								Type:            "int64",
								OriginalType:    "int64",
								NullValue:       "0",
								Len:             -1,
								ColumnName:      "InterleavedId",
								SpannerDataType: "INT64",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
							{
								Name:            "ID",
								Type:            "int64",
								OriginalType:    "int64",
								NullValue:       "0",
								Len:             -1,
								ColumnName:      "Id",
								SpannerDataType: "INT64",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
							{
								Name:            "Value",
								Type:            "int64",
								OriginalType:    "int64",
								NullValue:       "0",
								Len:             -1,
								ColumnName:      "Value",
								SpannerDataType: "INT64",
								IsNotNull:       true,
								IsPrimaryKey:    false,
							},
						},
						TableName: "Interleaved",
						Indexes: []*models.Index{
							{
								Name:           "InterleavedKey",
								FuncName:       "InterleavedsByInterleavedKey",
								LegacyFuncName: "InterleavedsByIDValue",
								Fields: []*models.Field{
									{ColumnName: "Id"},
									{ColumnName: "Value"},
								},
								IndexName: "InterleavedKey",
							},
						},
					},
					{
						Name: "Parent",
						PrimaryKeyFields: []*models.Field{
							{ColumnName: "Id"},
						},
						Fields: []*models.Field{
							{
								Name:            "ID",
								Type:            "int64",
								OriginalType:    "int64",
								NullValue:       "0",
								Len:             -1,
								ColumnName:      "Id",
								SpannerDataType: "INT64",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
						},
						TableName: "Parent",
					},
				},
			},
		},
		{
			name:   "OutOfOrderPrimaryKey",
			opt:    Option{},
			schema: oooSchema,
			expectedSchema: &models.Schema{
				Types: []*models.Type{
					{
						Name: "OutOfOrderPrimaryKey",
						PrimaryKeyFields: []*models.Field{
							{ColumnName: "PKey2"},
							{ColumnName: "PKey1"},
							{ColumnName: "PKey3"},
						},
						Fields: []*models.Field{
							{
								Name:            "PKey1",
								Type:            "string",
								OriginalType:    "string",
								NullValue:       `""`,
								Len:             32,
								ColumnName:      "PKey1",
								SpannerDataType: "STRING(32)",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
							{
								Name:            "PKey2",
								Type:            "string",
								OriginalType:    "string",
								NullValue:       `""`,
								Len:             32,
								ColumnName:      "PKey2",
								SpannerDataType: "STRING(32)",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
							{
								Name:            "PKey3",
								Type:            "string",
								OriginalType:    "string",
								NullValue:       `""`,
								Len:             32,
								ColumnName:      "PKey3",
								SpannerDataType: "STRING(32)",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
						},
						TableName: "OutOfOrderPrimaryKeys",
					},
				},
			},
		},
		{
			name:   "MaxLength",
			opt:    Option{},
			schema: maxLengthSchema,
			expectedSchema: &models.Schema{
				Types: []*models.Type{
					{
						Name: "MaxLength",
						PrimaryKeyFields: []*models.Field{
							{ColumnName: "MaxString"},
						},
						Fields: []*models.Field{
							{
								Name:            "MaxString",
								Type:            "string",
								OriginalType:    "string",
								NullValue:       `""`,
								Len:             -1,
								ColumnName:      "MaxString",
								SpannerDataType: "STRING(MAX)",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
							{
								Name:            "MaxBytes",
								Type:            "[]byte",
								OriginalType:    "[]byte",
								NullValue:       "nil",
								Len:             -1,
								ColumnName:      "MaxBytes",
								SpannerDataType: "BYTES(MAX)",
								IsNotNull:       true,
								IsPrimaryKey:    false,
							},
						},
						TableName: "MaxLengths",
					},
				},
			},
		},
		{
			name:   "AlterTableAddFK",
			opt:    Option{},
			schema: alterTableAddFKSchema,
			expectedSchema: &models.Schema{
				Types: []*models.Type{},
			},
		},
		{
			name:   "AlterTableAddConstraintFK",
			opt:    Option{},
			schema: alterTableAddConstraintFKSchema,
			expectedSchema: &models.Schema{
				Types: []*models.Type{},
			},
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			l := setUpTypeLoader(t, tc.schema, tc.opt)

			schema, err := l.LoadSchema()
			if err != nil {
				t.Fatalf("failed to load schema: %v", err)
			}

			compareSchemas(t, schema, tc.expectedSchema)
		})
	}
}

func TestLoader_CustomTypes(t *testing.T) {
	table := []struct {
		name           string
		opt            Option
		schema         string
		expectedSchema *models.Schema
		expectedErr    string
	}{
		{
			name: "Custom type table does not exist",
			opt: Option{
				Config: &config.Config{
					Tables: []config.Table{
						{
							Name: "UnknownTable",
						},
					},
				},
			},
			schema:      simpleSchema,
			expectedErr: "unknown custom type table UnknownTable",
		},
		{
			name: "Custom type column does not exist",
			opt: Option{
				Config: &config.Config{
					Tables: []config.Table{
						{
							Name: "Simple",
							Columns: []config.Column{
								{
									Name:       "UnknownColumn",
									CustomType: "UnknownCustomColumn",
								},
							},
						},
					},
				},
			},
			schema:      simpleSchema,
			expectedErr: "unknown custom type column UnknownColumn in the table Simple",
		},
		{
			name: "Success",
			opt: Option{
				Config: &config.Config{
					Tables: []config.Table{
						{
							Name: "Simple",
							Columns: []config.Column{
								{
									Name:       "Value",
									CustomType: "Value",
								},
							},
						},
					},
				},
			},
			schema: simpleSchema,
			expectedSchema: &models.Schema{
				Types: []*models.Type{
					{
						Name: "Simple",
						PrimaryKeyFields: []*models.Field{
							{ColumnName: "Id"},
						},
						Fields: []*models.Field{
							{
								Name:            "ID",
								Type:            "int64",
								OriginalType:    "int64",
								NullValue:       "0",
								Len:             -1,
								ColumnName:      "Id",
								SpannerDataType: "INT64",
								IsNotNull:       true,
								IsPrimaryKey:    true,
							},
							{
								Name:            "Value",
								Type:            "Value",
								OriginalType:    "string",
								NullValue:       `""`,
								Len:             32,
								ColumnName:      "Value",
								SpannerDataType: "STRING(32)",
								IsNotNull:       true,
								IsPrimaryKey:    false,
							},
						},
						TableName: "Simple",
						Indexes: []*models.Index{
							{
								Name:           "SimpleIndex",
								FuncName:       "SimplesBySimpleIndex",
								LegacyFuncName: "SimplesByValue",
								Fields: []*models.Field{
									{ColumnName: "Value"},
								},
								IndexName: "SimpleIndex",
							},
							{
								Name:           "SimpleIndex2",
								FuncName:       "SimpleBySimpleIndex2",
								LegacyFuncName: "SimpleByIDValue",
								Fields: []*models.Field{
									{ColumnName: "Id"},
									{ColumnName: "Value"},
								},
								IndexName: "SimpleIndex2",
								IsUnique:  true,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			l := setUpTypeLoader(t, tc.schema, tc.opt)

			schema, err := l.LoadSchema()

			if tc.expectedErr != "" {
				if err == nil {
					t.Fatal("expected to load schema failure")
				}

				if err.Error() != tc.expectedErr {
					t.Fatalf("unexpected error: expected: %s, actual: %s", tc.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("failed to load schema: %v", err)
				}

				compareSchemas(t, schema, tc.expectedSchema)
			}
		})
	}
}

func Test_setIndexesToTables(t *testing.T) {
	tests := []struct {
		table  map[string]*models.Type
		ix     map[string]*models.Index
		result map[string]int
	}{
		{
			table: map[string]*models.Type{
				"TableA": &models.Type{
					Indexes: []*models.Index{},
				},
			},
			ix: map[string]*models.Index{
				"TableA_Index1": &models.Index{
					Type: &models.Type{
						TableName: "TableA",
					},
					IndexName: "Index1",
				},
				"TableA_Index2": &models.Index{
					Type: &models.Type{
						TableName: "TableA",
					},
					IndexName: "Index2",
				},
			},
			result: map[string]int{
				"TableA": 2,
			},
		},
		{
			table: map[string]*models.Type{
				"TableA": &models.Type{
					Indexes: []*models.Index{},
				},
				"TableB": &models.Type{
					Indexes: []*models.Index{},
				},
			},
			ix: map[string]*models.Index{
				"TableA_Index1": &models.Index{
					Type: &models.Type{
						TableName: "TableA",
					},
					IndexName: "Index1",
				},
				"TableA_Index2": &models.Index{
					Type: &models.Type{
						TableName: "TableA",
					},
					IndexName: "Index2",
				},
			},
			result: map[string]int{
				"TableA": 2,
				"TableB": 0,
			},
		},
		{
			table: map[string]*models.Type{
				"TableA": &models.Type{
					Indexes: []*models.Index{},
				},
				"TableB": &models.Type{
					Indexes: []*models.Index{},
				},
			},
			ix: map[string]*models.Index{
				"TableA_Index1": &models.Index{
					Type: &models.Type{
						TableName: "TableA",
					},
					IndexName: "Index1",
				},
				"TableA_Index2": &models.Index{
					Type: &models.Type{
						TableName: "TableA",
					},
					IndexName: "Index2",
				},
				"TableB_Index1": &models.Index{
					Type: &models.Type{
						TableName: "TableB",
					},
					IndexName: "Index1",
				},
				"TableB_Index2forTableA_Hoge": &models.Index{
					Type: &models.Type{
						TableName: "TableB",
					},
					IndexName: "Index2",
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

func setUpTypeLoader(t *testing.T, schema string, opt Option) *TypeLoader {
	t.Helper()

	dir := t.TempDir()

	f, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_, _ = f.Write([]byte(schema))
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

	return NewTypeLoader(source, inflector, opt)
}

func compareSchemas(t *testing.T, actual *models.Schema, expected *models.Schema) {
	t.Helper()

	if diff := cmp.Diff(actual, expected,
		cmp.Transformer("FilterInTypePrimaryKeyFields", func(in *models.Type) *models.Type {
			if in == nil {
				return in
			}
			for i := range in.PrimaryKeyFields {
				f := in.PrimaryKeyFields[i]
				in.PrimaryKeyFields[i] = &models.Field{ColumnName: f.ColumnName}
			}
			return in
		}),
		cmp.Transformer("FilterInIndexFields", func(in *models.Index) *models.Index {
			for i := range in.Fields {
				f := in.Fields[i]
				in.Fields[i] = &models.Field{ColumnName: f.ColumnName}
			}
			return in
		}),
		cmpopts.IgnoreFields(models.Index{}, "Type"),
	); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}
