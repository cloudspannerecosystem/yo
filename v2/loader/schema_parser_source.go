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
	"github.com/cloudspannerecosystem/memefish/ast"
	"os"
	"sort"
	"strings"

	"github.com/cloudspannerecosystem/memefish"
)

func extractName(path *ast.Path) (string, error) {
	if len(path.Idents) != 1 {
		return "", fmt.Errorf("path isn't simple ident: %v", path.SQL())
	}
	return path.Idents[0].Name, nil
}

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

		ddlstmt, err := memefish.ParseDDL("", stmt)
		if err != nil {
			return nil, err
		}

		switch val := ddlstmt.(type) {
		case *ast.CreateTable:
			tableName, err := extractName(val.Name)
			if err != nil {
				return nil, err
			}

			v := tables[tableName]
			v.createTable = val
			tables[tableName] = v
		case *ast.CreateIndex:
			tableName, err := extractName(val.TableName)
			if err != nil {
				return nil, err
			}

			v := tables[tableName]
			v.createIndexes = append(v.createIndexes, val)
			tables[tableName] = v
		case *ast.AlterTable:
			if isAlterTableAddFK(val) {
				continue
			}
			return nil, fmt.Errorf("unknown statement is specified: %s", ddlstmt.SQL())
		}
	}

	return &schemaParserSource{tables: tables}, nil
}

func isAlterTableAddFK(at *ast.AlterTable) bool {
	ac, ok := at.TableAlteration.(*ast.AddTableConstraint)
	if !ok {
		return false
	}
	_, ok = ac.TableConstraint.Constraint.(*ast.ForeignKey)
	return ok
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
			var err error
			parent, err = extractName(t.createTable.Cluster.TableName)
			if err != nil {
				return nil, err
			}
		}
		tableName, err := extractName(t.createTable.Name)
		if err != nil {
			return nil, err
		}

		tables = append(tables, &SpannerTable{
			TableName:       tableName,
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
			IsGenerated:  c.GeneratedExpr != nil,
		})
	}

	return cols, nil
}

func (s *schemaParserSource) IndexList(name string) ([]*SpannerIndex, error) {
	var indexes []*SpannerIndex
	for _, index := range s.tables[name].createIndexes {
		indexName, err := extractName(index.Name)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, &SpannerIndex{
			IndexName: indexName,
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
		ixName, err := extractName(ix.Name)
		if err != nil {
			return nil, err
		}
		if ixName != index {
			continue
		}

		if ix.Storing != nil {
			// add storing columns first
			for _, storing := range ix.Storing.Columns {
				cols = append(cols, &SpannerIndexColumn{
					SeqNo:      0,
					Storing:    true,
					ColumnName: storing.Name,
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
