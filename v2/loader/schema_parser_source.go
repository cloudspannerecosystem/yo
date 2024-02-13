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
	"sort"
	"strings"

	"cloud.google.com/go/spanner/spansql"
)

func NewSchemaParserSource(fpath string) (SchemaSource, error) {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	tables := make(map[string]table)
	stmts := strings.Split(string(b), ";")
	for _, stmt := range stmts {
		stmt := strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		ddlstmt, err := spansql.ParseDDLStmt(stmt)
		if err != nil {
			return nil, err
		}

		switch val := ddlstmt.(type) {
		case *spansql.CreateTable:
			tableName := string(val.Name)
			v := tables[tableName]
			v.createTable = val
			tables[tableName] = v
		case *spansql.CreateIndex:
			tableName := string(val.Table)
			v := tables[tableName]
			v.createIndexes = append(v.createIndexes, val)
			tables[tableName] = v
		case *spansql.CreateChangeStream:
			// CreateChangeStream isn't supported yet
			continue
		case *spansql.AlterTable:
			if isAlterTableAddFK(val) {
				continue
			}
			return nil, fmt.Errorf("stmt should be CreateTable, CreateIndex, CreateChangeStream or AlterTableAddForeignKey, but got '%s'", ddlstmt.SQL())
		default:
			return nil, fmt.Errorf("stmt should be CreateTable, CreateIndex, CreateChangeStream or AlterTableAddForeignKey, but got '%s'", ddlstmt.SQL())
		}
	}

	return &schemaParserSource{tables: tables}, nil
}

func isAlterTableAddFK(at *spansql.AlterTable) bool {
	ac, ok := at.Alteration.(spansql.AddConstraint)
	if !ok {
		return false
	}
	_, ok = ac.Constraint.Constraint.(spansql.ForeignKey)
	return ok
}

type table struct {
	createTable   *spansql.CreateTable
	createIndexes []*spansql.CreateIndex
}

type schemaParserSource struct {
	tables map[string]table
}

func (s *schemaParserSource) TableList() ([]*SpannerTable, error) {
	var tables []*SpannerTable
	for _, t := range s.tables {
		var parent string
		if t.createTable.Interleave != nil {
			parent = string(t.createTable.Interleave.Parent)
		}

		tables = append(tables, &SpannerTable{
			TableName:       string(t.createTable.Name),
			ParentTableName: parent,
		})
	}

	sort.Slice(tables, func(i, j int) bool {
		return tables[i].TableName < tables[j].TableName
	})

	return tables, nil
}

func (s *schemaParserSource) ColumnList(name string) ([]*SpannerColumn, error) {
	var cols []*SpannerColumn
	table := s.tables[name].createTable

	check := make(map[string]struct{})
	for _, pk := range table.PrimaryKey {
		check[string(pk.Column)] = struct{}{}
	}

	for i, c := range table.Columns {
		_, pk := check[string(c.Name)]
		cols = append(cols, &SpannerColumn{
			FieldOrdinal: i + 1,
			ColumnName:   string(c.Name),
			DataType:     c.Type.SQL(),
			NotNull:      c.NotNull,
			IsPrimaryKey: pk,
			IsGenerated:  c.Generated != nil,
		})
	}

	return cols, nil
}

func (s *schemaParserSource) IndexList(name string) ([]*SpannerIndex, error) {
	var indexes []*SpannerIndex
	for _, index := range s.tables[name].createIndexes {
		indexes = append(indexes, &SpannerIndex{
			IndexName: string(index.Name),
			IsUnique:  index.Unique,
		})
	}

	return indexes, nil
}

func (s *schemaParserSource) IndexColumnList(table, index string) ([]*SpannerIndexColumn, error) {
	if index == "PRIMARY_KEY" {
		return s.primaryKeyColumnList(table)
	}

	var cols []*SpannerIndexColumn
	for _, ix := range s.tables[table].createIndexes {
		if string(ix.Name) != index {
			continue
		}

		// add storing columns first
		for _, storing := range ix.Storing {
			cols = append(cols, &SpannerIndexColumn{
				SeqNo:      0,
				Storing:    true,
				ColumnName: string(storing),
			})
		}

		for i, c := range ix.Columns {
			cols = append(cols, &SpannerIndexColumn{
				SeqNo:      i + 1,
				ColumnName: string(c.Column),
			})
		}
		break
	}

	return cols, nil
}

func (s *schemaParserSource) primaryKeyColumnList(table string) ([]*SpannerIndexColumn, error) {
	tbl, ok := s.tables[table]
	if !ok {
		return nil, nil
	}

	var cols []*SpannerIndexColumn
	for i, key := range tbl.createTable.PrimaryKey {
		cols = append(cols, &SpannerIndexColumn{
			SeqNo:      i + 1,
			ColumnName: string(key.Column),
		})
	}

	return cols, nil
}
