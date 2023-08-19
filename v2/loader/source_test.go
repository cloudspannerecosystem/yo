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
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.mercari.io/yo/v2/test/testutil"
)

var (
	testSchema1 = `
CREATE TABLE Simple (
  Id INT64 NOT NULL,
  Value STRING(32) NOT NULL,
) PRIMARY KEY(Id);
CREATE INDEX SimpleIndex ON Simple(Value);
CREATE UNIQUE INDEX SimpleIndex2 ON Simple(Id, Value);
`
	testSchema2 = `
CREATE TABLE MaxLengths (
  MaxString STRING(MAX) NOT NULL,
  MaxBytes BYTES(MAX) NOT NULL,
) PRIMARY KEY(MaxString);
`
	testSchema3 = `
CREATE TABLE FullTypes (
  PKey STRING(32) NOT NULL,
  FTString STRING(32) NOT NULL,
  FTStringNull STRING(32),
  FTBool BOOL NOT NULL,
  FTBoolNull BOOL,
  FTBytes BYTES(32) NOT NULL,
  FTBytesNull BYTES(32),
  FTTimestamp TIMESTAMP NOT NULL,
  FTTimestampNull TIMESTAMP,
  FTInt INT64 NOT NULL,
  FTIntNull INT64,
  FTFloat FLOAT64 NOT NULL,
  FTFloatNull FLOAT64,
  FTDate DATE NOT NULL,
  FTDateNull DATE,
  FTArrayStringNull ARRAY<STRING(32)>,
  FTArrayString ARRAY<STRING(32)> NOT NULL,
  FTArrayBoolNull ARRAY<BOOL>,
  FTArrayBool ARRAY<BOOL> NOT NULL,
  FTArrayBytesNull ARRAY<BYTES(32)>,
  FTArrayBytes ARRAY<BYTES(32)> NOT NULL,
  FTArrayTimestampNull ARRAY<TIMESTAMP>,
  FTArrayTimestamp ARRAY<TIMESTAMP> NOT NULL,
  FTArrayIntNull ARRAY<INT64>,
  FTArrayInt ARRAY<INT64> NOT NULL,
  FTArrayFloatNull ARRAY<FLOAT64>,
  FTArrayFloat ARRAY<FLOAT64> NOT NULL,
  FTArrayDateNull ARRAY<DATE>,
  FTArrayDate ARRAY<DATE> NOT NULL,
) PRIMARY KEY(PKey);
`

	testSchema4 = `
CREATE TABLE Items (
  ID INT64 NOT NULL,
) PRIMARY KEY (ID);

CREATE TABLE ForeignItems (
  ID INT64 NOT NULL,
  ItemID INT64 NOT NULL,
  CONSTRAINT FK_ItemID_ForeignItems FOREIGN KEY (ItemID) REFERENCES Items (ID)
) PRIMARY KEY (ID);
`

	testSchema5 = `
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
)

func TestSource(t *testing.T) {
	dir := t.TempDir()

	table := []struct {
		name                 string
		schema               string
		expectedTables       []*SpannerTable
		expectedColumns      map[string][]*SpannerColumn
		expectedIndex        map[string][]*SpannerIndex
		expectedIndexColumns map[string][]*SpannerIndexColumn
	}{
		{
			name:   "Simple",
			schema: testSchema1,
			expectedTables: []*SpannerTable{
				{
					TableName:       "Simple",
					ParentTableName: "",
				},
			},
			expectedColumns: map[string][]*SpannerColumn{
				"Simple": {
					{
						FieldOrdinal: 1,
						ColumnName:   "Id",
						DataType:     "INT64",
						NotNull:      true,
						IsPrimaryKey: true,
					},
					{
						FieldOrdinal: 2,
						ColumnName:   "Value",
						DataType:     "STRING(32)",
						NotNull:      true,
					},
				},
			},
			expectedIndex: map[string][]*SpannerIndex{
				"Simple": {
					{
						IndexName: "SimpleIndex",
						IsUnique:  false,
						IsPrimary: false,
					},
					{
						IndexName: "SimpleIndex2",
						IsUnique:  true,
						IsPrimary: false,
					},
				},
			},
			expectedIndexColumns: map[string][]*SpannerIndexColumn{
				"Simple/SimpleIndex": {
					{
						SeqNo:      1,
						ColumnName: "Value",
					},
				},
				"Simple/SimpleIndex2": {
					{
						SeqNo:      1,
						ColumnName: "Id",
					},
					{
						SeqNo:      2,
						ColumnName: "Value",
					},
				},
			},
		},
		{
			name:   "MaxLength",
			schema: testSchema2,
			expectedTables: []*SpannerTable{
				{
					TableName:       "MaxLengths",
					ParentTableName: "",
				},
			},
			expectedColumns: map[string][]*SpannerColumn{
				"MaxLengths": {
					{
						FieldOrdinal: 1,
						ColumnName:   "MaxString",
						DataType:     "STRING(MAX)",
						NotNull:      true,
						IsPrimaryKey: true,
					},
					{
						FieldOrdinal: 2,
						ColumnName:   "MaxBytes",
						DataType:     "BYTES(MAX)",
						NotNull:      true,
					},
				},
			},
			expectedIndex: map[string][]*SpannerIndex{
				"MaxLengths": nil,
			},
			expectedIndexColumns: map[string][]*SpannerIndexColumn{},
		},
		{
			name:   "FullTypes",
			schema: testSchema3,
			expectedTables: []*SpannerTable{
				{
					TableName:       "FullTypes",
					ParentTableName: "",
				},
			},
			expectedColumns: map[string][]*SpannerColumn{
				"FullTypes": {
					{FieldOrdinal: 1, ColumnName: "PKey", DataType: "STRING(32)", NotNull: true, IsPrimaryKey: true},
					{FieldOrdinal: 2, ColumnName: "FTString", DataType: "STRING(32)", NotNull: true},
					{FieldOrdinal: 3, ColumnName: "FTStringNull", DataType: "STRING(32)"},
					{FieldOrdinal: 4, ColumnName: "FTBool", DataType: "BOOL", NotNull: true},
					{FieldOrdinal: 5, ColumnName: "FTBoolNull", DataType: "BOOL"},
					{FieldOrdinal: 6, ColumnName: "FTBytes", DataType: "BYTES(32)", NotNull: true},
					{FieldOrdinal: 7, ColumnName: "FTBytesNull", DataType: "BYTES(32)"},
					{FieldOrdinal: 8, ColumnName: "FTTimestamp", DataType: "TIMESTAMP", NotNull: true},
					{FieldOrdinal: 9, ColumnName: "FTTimestampNull", DataType: "TIMESTAMP"},
					{FieldOrdinal: 10, ColumnName: "FTInt", DataType: "INT64", NotNull: true},
					{FieldOrdinal: 11, ColumnName: "FTIntNull", DataType: "INT64"},
					{FieldOrdinal: 12, ColumnName: "FTFloat", DataType: "FLOAT64", NotNull: true},
					{FieldOrdinal: 13, ColumnName: "FTFloatNull", DataType: "FLOAT64"},
					{FieldOrdinal: 14, ColumnName: "FTDate", DataType: "DATE", NotNull: true},
					{FieldOrdinal: 15, ColumnName: "FTDateNull", DataType: "DATE"},
					{FieldOrdinal: 16, ColumnName: "FTArrayStringNull", DataType: "ARRAY<STRING(32)>"},
					{FieldOrdinal: 17, ColumnName: "FTArrayString", DataType: "ARRAY<STRING(32)>", NotNull: true},
					{FieldOrdinal: 18, ColumnName: "FTArrayBoolNull", DataType: "ARRAY<BOOL>"},
					{FieldOrdinal: 19, ColumnName: "FTArrayBool", DataType: "ARRAY<BOOL>", NotNull: true},
					{FieldOrdinal: 20, ColumnName: "FTArrayBytesNull", DataType: "ARRAY<BYTES(32)>"},
					{FieldOrdinal: 21, ColumnName: "FTArrayBytes", DataType: "ARRAY<BYTES(32)>", NotNull: true}, {FieldOrdinal: 22, ColumnName: "FTArrayTimestampNull", DataType: "ARRAY<TIMESTAMP>"},
					{FieldOrdinal: 23, ColumnName: "FTArrayTimestamp", DataType: "ARRAY<TIMESTAMP>", NotNull: true},
					{FieldOrdinal: 24, ColumnName: "FTArrayIntNull", DataType: "ARRAY<INT64>"},
					{FieldOrdinal: 25, ColumnName: "FTArrayInt", DataType: "ARRAY<INT64>", NotNull: true},
					{FieldOrdinal: 26, ColumnName: "FTArrayFloatNull", DataType: "ARRAY<FLOAT64>"},
					{FieldOrdinal: 27, ColumnName: "FTArrayFloat", DataType: "ARRAY<FLOAT64>", NotNull: true},
					{FieldOrdinal: 28, ColumnName: "FTArrayDateNull", DataType: "ARRAY<DATE>"},
					{FieldOrdinal: 29, ColumnName: "FTArrayDate", DataType: "ARRAY<DATE>", NotNull: true},
				},
			},
			expectedIndex: map[string][]*SpannerIndex{
				"FullTypes": nil,
			},
			expectedIndexColumns: map[string][]*SpannerIndexColumn{},
		},
		{
			name:   "ForeignKey",
			schema: testSchema4,
			expectedTables: []*SpannerTable{
				{
					TableName: "ForeignItems",
				},
				{
					TableName: "Items",
				},
			},
			expectedColumns: map[string][]*SpannerColumn{
				"Items": {
					{FieldOrdinal: 1, ColumnName: "ID", DataType: "INT64", NotNull: true, IsPrimaryKey: true},
				},
				"ForeignItems": {
					{FieldOrdinal: 1, ColumnName: "ID", DataType: "INT64", NotNull: true, IsPrimaryKey: true},
					{FieldOrdinal: 2, ColumnName: "ItemID", DataType: "INT64", NotNull: true},
				},
			},
			expectedIndex: map[string][]*SpannerIndex{
				"Items":        nil,
				"ForeignItems": nil,
			},
			expectedIndexColumns: map[string][]*SpannerIndexColumn{},
		},
		{
			name:   "Interleave",
			schema: testSchema5,
			expectedTables: []*SpannerTable{
				{
					TableName:       "Interleaved",
					ParentTableName: "Parent",
				},
				{
					TableName: "Parent",
				},
			},
			expectedColumns: map[string][]*SpannerColumn{
				"Parent": {
					{FieldOrdinal: 1, ColumnName: "Id", DataType: "INT64", NotNull: true, IsPrimaryKey: true},
				},
				"Interleaved": {
					{FieldOrdinal: 1, ColumnName: "InterleavedId", DataType: "INT64", NotNull: true, IsPrimaryKey: true},
					{FieldOrdinal: 2, ColumnName: "Id", DataType: "INT64", NotNull: true, IsPrimaryKey: true},
					{FieldOrdinal: 3, ColumnName: "Value", DataType: "INT64", NotNull: true, IsPrimaryKey: false},
				},
			},
			expectedIndex: map[string][]*SpannerIndex{
				"Parent": nil,
				"Interleaved": {
					{IndexName: "InterleavedKey"},
				},
			},
			expectedIndexColumns: map[string][]*SpannerIndexColumn{
				"Interleaved/InterleavedKey": {
					{SeqNo: 1, ColumnName: "Id"},
					{SeqNo: 2, ColumnName: "Value"},
				},
			},
		},
	}

	for _, tc := range table {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			f, err := os.CreateTemp(dir, "")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			_, _ = f.Write([]byte(tc.schema))
			_ = f.Close()
			path := f.Name()

			parserSource, err := NewSchemaParserSource(path)
			if err != nil {
				t.Fatalf("failed to create schema parser source: %v", err)
			}

			if err := testutil.SetupDatabase(ctx, "yo-test", "yo-loader-test", "source-test", tc.schema); err != nil {
				t.Fatalf("failed to setup database: %v", err)
			}

			client, err := testutil.TestClient(ctx, "yo-test", "yo-loader-test", "source-test")
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}
			defer client.Close()

			informationSchemaSource, err := NewInformationSchemaSource(client)
			if err != nil {
				t.Fatalf("failed to create information schema source: %v", err)
			}

			sourceMap := map[string]SchemaSource{
				"Parser":            parserSource,
				"InformationSchema": informationSchemaSource,
			}

			for name, s := range sourceMap {
				t.Run(name, func(t *testing.T) {
					tbls, err := s.TableList()
					if err != nil {
						t.Fatalf("TableList failed: %v", err)
					}

					if diff := cmp.Diff(tc.expectedTables, tbls); diff != "" {
						t.Errorf("(-got, +want)\n%s", diff)
					}

					var tables []string
					for _, tbl := range tbls {
						tables = append(tables, tbl.TableName)
					}

					gotColumns := make(map[string][]*SpannerColumn)
					for _, tbl := range tables {
						columns, err := s.ColumnList(tbl)
						if err != nil {
							t.Fatalf("TableList failed: %v", err)
						}
						gotColumns[tbl] = columns
					}

					if diff := cmp.Diff(tc.expectedColumns, gotColumns); diff != "" {
						t.Errorf("(-got, +want)\n%s", diff)
					}

					gotIndex := make(map[string][]*SpannerIndex)
					for _, tbl := range tables {
						index, err := s.IndexList(tbl)
						if err != nil {
							t.Fatalf("IndexList failed: %v", err)
						}
						gotIndex[tbl] = index
					}

					if diff := cmp.Diff(tc.expectedIndex, gotIndex); diff != "" {
						t.Errorf("(-got, +want)\n%s", diff)
					}

					gotIndexColumns := make(map[string][]*SpannerIndexColumn)
					for tbl, indexes := range gotIndex {
						for _, index := range indexes {
							columns, err := s.IndexColumnList(tbl, index.IndexName)
							if err != nil {
								t.Fatalf("IndexColumnList failed: %v", err)
							}
							gotIndexColumns[fmt.Sprintf("%s/%s", tbl, index.IndexName)] = columns
						}
					}

					if diff := cmp.Diff(tc.expectedIndexColumns, gotIndexColumns); diff != "" {
						t.Errorf("(-got, +want)\n%s", diff)
					}
				})
			}
		})
	}
}
