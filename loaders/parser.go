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

package loaders

import (
	"fmt"
	"io/ioutil"

	parser "github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/cloudspannerecosystem/memefish/token"
	"go.mercari.io/yo/models"
)

func NewSpannerLoaderFromDDL(fpath string) (*SpannerLoaderFromDDL, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	tables := make(map[string]table)
	ddls, err := (&parser.Parser{
		Lexer: &parser.Lexer{
			File: &token.File{FilePath: fpath, Buffer: string(b)},
		},
	}).ParseDDLs()
	if err != nil {
		return nil, err
	}
	for _, ddl := range ddls {
		switch val := ddl.(type) {
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
			v, ok := tables[tableName]
			if !ok {
				return nil, fmt.Errorf("table '%s' is undefined, but got '%s'", tableName, ddl.SQL())
			}
			v.createIndexes = append(v.createIndexes, val)
			tables[tableName] = v
		case *ast.AlterTable:
			if _, ok := val.TableAlteration.(*ast.AddTableConstraint); ok {
				continue
			}
			return nil, fmt.Errorf("stmt should be CreateTable, CreateIndex or AlterTableAddConstraint, but got '%s'", ddl.SQL())
		}
	}

	return &SpannerLoaderFromDDL{tables: tables}, nil
}

type table struct {
	createTable   *ast.CreateTable
	createIndexes []*ast.CreateIndex
}

type SpannerLoaderFromDDL struct {
	tables map[string]table
}

func (s *SpannerLoaderFromDDL) ParamN(n int) string {
	return fmt.Sprintf("@param%d", n)
}

func (s *SpannerLoaderFromDDL) MaskFunc() string {
	return "?"
}

func (s *SpannerLoaderFromDDL) ParseType(dt string, nullable bool) (int, string, string) {
	return SpanParseType(dt, nullable)
}

func (s *SpannerLoaderFromDDL) ValidCustomType(dataType string, customType string) bool {
	return SpanValidateCustomType(dataType, customType)
}

func (s *SpannerLoaderFromDDL) TableList() ([]*models.Table, error) {
	var tables []*models.Table
	for _, t := range s.tables {
		tableName, err := extractName(t.createTable.Name)
		if err != nil {
			return nil, err
		}
		tables = append(tables, &models.Table{
			TableName: tableName,
			ManualPk:  true,
		})
	}

	return tables, nil
}

func (s *SpannerLoaderFromDDL) ColumnList(name string) ([]*models.Column, error) {
	var cols []*models.Column
	table := s.tables[name].createTable

	check := make(map[string]struct{})
	for _, pk := range table.PrimaryKeys {
		check[pk.Name.Name] = struct{}{}
	}

	for i, c := range table.Columns {
		_, pk := check[c.Name.Name]
		isGenerated := false
		if _, ok := c.DefaultSemantics.(*ast.GeneratedColumnExpr); ok {
			isGenerated = true
		}

		allowCommitTimestamp := false
		if c.Options != nil {
			for _, r := range c.Options.Records {
				if r.Name.Name == "allow_commit_timestamp" {
					boolLiteral, ok := r.Value.(*ast.BoolLiteral)
					if !ok {
						return nil, fmt.Errorf("the type of 'allow_commit_timestamp' should be 'bool', but got '%T'", r.Value)
					}
					allowCommitTimestamp = boolLiteral.Value
					break
				}
			}
		}

		cols = append(cols, &models.Column{
			FieldOrdinal:         i + 1,
			ColumnName:           c.Name.Name,
			DataType:             c.Type.SQL(),
			NotNull:              c.NotNull,
			IsPrimaryKey:         pk,
			IsGenerated:          isGenerated,
			AllowCommitTimestamp: allowCommitTimestamp,
		})
	}

	return cols, nil
}

func (s *SpannerLoaderFromDDL) IndexList(name string) ([]*models.Index, error) {
	var indexes []*models.Index
	for _, index := range s.tables[name].createIndexes {
		idxName, err := extractName(index.Name)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, &models.Index{
			IndexName: idxName,
			IsUnique:  index.Unique,
		})
	}

	return indexes, nil
}

func (s *SpannerLoaderFromDDL) IndexColumnList(table, index string) ([]*models.IndexColumn, error) {
	if index == "PRIMARY_KEY" {
		return s.primaryKeyColumnList(table)
	}

	var cols []*models.IndexColumn
	for _, ix := range s.tables[table].createIndexes {
		idxName, err := extractName(ix.Name)
		if err != nil {
			return nil, err
		}
		if idxName != index {
			continue
		}

		// add storing columns first
		if ix.Storing != nil {
			for _, c := range ix.Storing.Columns {
				cols = append(cols, &models.IndexColumn{
					SeqNo:      0,
					Storing:    true,
					ColumnName: c.Name,
				})
			}
		}

		for i, c := range ix.Keys {
			cols = append(cols, &models.IndexColumn{
				SeqNo:      i + 1,
				ColumnName: c.Name.Name,
			})
		}
		break
	}

	return cols, nil
}

func (s *SpannerLoaderFromDDL) primaryKeyColumnList(table string) ([]*models.IndexColumn, error) {
	tbl, ok := s.tables[table]
	if !ok {
		return nil, nil
	}

	var cols []*models.IndexColumn
	for i, key := range tbl.createTable.PrimaryKeys {
		cols = append(cols, &models.IndexColumn{
			SeqNo:      i + 1,
			ColumnName: key.Name.Name,
		})
	}

	return cols, nil
}

func extractName(path *ast.Path) (string, error) {
	if len(path.Idents) != 1 {
		return "", fmt.Errorf("path isn't simple ident: %v", path.SQL())
	}
	return path.Idents[0].Name, nil
}
