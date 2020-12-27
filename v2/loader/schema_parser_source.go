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
	"sort"
	"strings"

	"github.com/MakeNowJust/memefish/pkg/ast"
	"github.com/MakeNowJust/memefish/pkg/parser"
	"github.com/MakeNowJust/memefish/pkg/token"
)

func NewSchemaParserSource(fpath string) (SchemaSource, error) {
	b, err := ioutil.ReadFile(fpath)
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
		ddl, err := (&parser.Parser{
			Lexer: &parser.Lexer{
				File: &token.File{FilePath: fpath, Buffer: stmt},
			},
		}).ParseDDL()
		if err != nil {
			return nil, err
		}

		switch val := ddl.(type) {
		case *ast.CreateTable:
			v := tables[val.Name.Name]
			v.createTable = val
			tables[val.Name.Name] = v
		case *ast.CreateIndex:
			v := tables[val.TableName.Name]
			v.createIndexes = append(v.createIndexes, val)
			tables[val.TableName.Name] = v
		default:
			return nil, fmt.Errorf("stmt should be CreateTable or CreateIndex, but got '%s'", ddl.SQL())
		}
	}

	return &schemaParserSource{tables: tables}, nil
}

type table struct {
	createTable   *ast.CreateTable
	createIndexes []*ast.CreateIndex
}

type schemaParserSource struct {
	tables map[string]table
}

func (s *schemaParserSource) TableList() ([]*SpannerTable, error) {
	var tables []*SpannerTable
	for _, t := range s.tables {
		var parent string
		if t.createTable.Cluster != nil {
			parent = t.createTable.Cluster.TableName.Name
		}

		tables = append(tables, &SpannerTable{
			TableName:       t.createTable.Name.Name,
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
	for _, pk := range table.PrimaryKeys {
		check[pk.Name.Name] = struct{}{}
	}

	for i, c := range table.Columns {
		_, pk := check[c.Name.Name]
		cols = append(cols, &SpannerColumn{
			FieldOrdinal: i + 1,
			ColumnName:   c.Name.Name,
			DataType:     c.Type.SQL(),
			NotNull:      c.NotNull,
			IsPrimaryKey: pk,
		})
	}

	return cols, nil
}

func (s *schemaParserSource) IndexList(name string) ([]*SpannerIndex, error) {
	var indexes []*SpannerIndex
	for _, index := range s.tables[name].createIndexes {
		indexes = append(indexes, &SpannerIndex{
			IndexName: index.Name.Name,
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
		if ix.Name.Name != index {
			continue
		}

		// add storing columns first
		if ix.Storing != nil {
			for _, c := range ix.Storing.Columns {
				cols = append(cols, &SpannerIndexColumn{
					SeqNo:      0,
					Storing:    true,
					ColumnName: c.Name,
				})
			}
		}

		for i, c := range ix.Keys {
			cols = append(cols, &SpannerIndexColumn{
				SeqNo:      i + 1,
				ColumnName: c.Name.Name,
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
	for i, key := range tbl.createTable.PrimaryKeys {
		cols = append(cols, &SpannerIndexColumn{
			SeqNo:      i + 1,
			ColumnName: key.Name.Name,
		})
	}

	return cols, nil
}
