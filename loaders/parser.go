package loaders

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/MakeNowJust/memefish/pkg/parser"
	"go.mercari.io/yo/models"
)

func NewSpannerLoaderFromDDL(fpath string) (*SpannerLoaderFromDDL, error) {
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
				File: &parser.File{FilePath: fpath, Buffer: stmt},
			},
		}).ParseDDL()
		if err != nil {
			return nil, err
		}

		switch val := ddl.(type) {
		case *parser.CreateTable:
			v := tables[val.Name.Name]
			v.createTable = val
			tables[val.Name.Name] = v
		case *parser.CreateIndex:
			v := tables[val.TableName.Name]
			v.createIndexes = append(v.createIndexes, val)
			tables[val.TableName.Name] = v
		default:
			return nil, fmt.Errorf("stmt should be CreateTable or CreateIndex, but got '%s'", ddl.SQL())
		}
	}

	return &SpannerLoaderFromDDL{tables: tables}, nil
}

type table struct {
	createTable   *parser.CreateTable
	createIndexes []*parser.CreateIndex
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
		tables = append(tables, &models.Table{
			TableName: t.createTable.Name.Name,
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
		cols = append(cols, &models.Column{
			FieldOrdinal: i + 1,
			ColumnName:   c.Name.Name,
			DataType:     c.Type.SQL(),
			NotNull:      c.NotNull,
			IsPrimaryKey: pk,
		})
	}

	return cols, nil
}

func (s *SpannerLoaderFromDDL) IndexList(name string) ([]*models.Index, error) {
	var indexes []*models.Index
	for _, index := range s.tables[name].createIndexes {
		indexes = append(indexes, &models.Index{
			IndexName: index.Name.Name,
			IsUnique:  index.Unique,
		})
	}

	return indexes, nil
}

func (s *SpannerLoaderFromDDL) IndexColumnList(table, index string) ([]*models.IndexColumn, error) {
	var cols []*models.IndexColumn
	for _, ix := range s.tables[table].createIndexes {
		if ix.Name.Name != index {
			continue
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
