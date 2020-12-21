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
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/kenshaw/snaker"
	"go.mercari.io/yo/v2/models"
	"google.golang.org/api/iterator"
)

func NewSpannerLoader(client *spanner.Client) *SpannerLoader {
	return &SpannerLoader{
		client: client,
	}
}

type SpannerLoader struct {
	client *spanner.Client
}

func (s *SpannerLoader) ParamN(n int) string {
	return fmt.Sprintf("@param%d", n)
}

func (s *SpannerLoader) MaskFunc() string {
	return "?"
}

func (s *SpannerLoader) ParseType(dt string, nullable bool) (int, string, string) {
	return SpanParseType(dt, nullable)
}

func (s *SpannerLoader) ValidCustomType(dataType string, customType string) bool {
	return SpanValidateCustomType(dataType, customType)
}

func (s *SpannerLoader) TableList() ([]*models.Table, error) {
	var err error

	// get the tables
	rows, err := spanTables(s.client)
	if err != nil {
		return nil, err
	}

	// Add information about manual FK.
	var tables []*models.Table
	for _, row := range rows {
		tables = append(tables, &models.Table{
			TableName: row.TableName,
			Type:      row.Type,
			ManualPk:  true,
		})
	}

	return tables, nil
}

func (s *SpannerLoader) ColumnList(table string) ([]*models.Column, error) {
	return SpanTableColumns(s.client, table)
}

func (s *SpannerLoader) IndexList(table string) ([]*models.Index, error) {
	return SpanTableIndexes(s.client, table)
}

func (s *SpannerLoader) IndexColumnList(table string, index string) ([]*models.IndexColumn, error) {
	return SpanIndexColumns(s.client, table, index)
}

var lengthRegexp = regexp.MustCompile(`\(([0-9]+|MAX)\)$`)

// SpanParseType parse a mysql type into a Go type based on the column
// definition.
func SpanParseType(dt string, nullable bool) (int, string, string) {
	nilVal := "nil"
	length := -1

	// separate type and length from dt with length such as STRING(32) or BYTES(256)
	m := lengthRegexp.FindStringSubmatchIndex(dt)
	if m != nil {
		lengthStr := dt[m[2]:m[3]]
		if lengthStr == "MAX" {
			length = -1
		} else {
			l, err := strconv.Atoi(lengthStr)
			if err != nil {
				panic("could not convert precision")
			}
			length = l
		}

		// trim length from dt
		dt = dt[:m[0]] + dt[m[1]:]
	}

	var typ string
	switch dt {
	case "BOOL":
		nilVal = "false"
		typ = "bool"
		if nullable {
			nilVal = "spanner.NullBool{}"
			typ = "spanner.NullBool"
		}

	case "STRING":
		nilVal = `""`
		typ = "string"
		if nullable {
			nilVal = "spanner.NullString{}"
			typ = "spanner.NullString"
		}

	case "INT64":
		nilVal = "0"
		typ = "int64"
		if nullable {
			nilVal = "spanner.NullInt64{}"
			typ = "spanner.NullInt64"
		}

	case "FLOAT64":
		nilVal = "0.0"
		typ = "float64"
		if nullable {
			nilVal = "spanner.NullFloat64{}"
			typ = "spanner.NullFloat64"
		}

	case "BYTES":
		typ = "[]byte"

	case "TIMESTAMP":
		nilVal = "time.Time{}"
		typ = "time.Time"
		if nullable {
			nilVal = "spanner.NullTime{}"
			typ = "spanner.NullTime"
		}

	case "DATE":
		nilVal = "civil.Date{}"
		typ = "civil.Date"
		if nullable {
			nilVal = "spanner.NullDate{}"
			typ = "spanner.NullDate"
		}

	default:
		if strings.HasPrefix(dt, "ARRAY<") {
			eleDataType := strings.TrimSuffix(strings.TrimPrefix(dt, "ARRAY<"), ">")
			_, _, eleTyp := SpanParseType(eleDataType, false)
			typ, nilVal = "[]"+eleTyp, "nil"
			if !nullable {
				nilVal = typ + "{}"
			}
			break
		}

		typ = snaker.SnakeToCamelIdentifier(dt)
		nilVal = typ + "{}"
	}

	return length, nilVal, typ
}

// spanTables runs a custom query, returning results as Table.
func spanTables(client *spanner.Client) ([]*models.Table, error) {
	ctx := context.Background()

	const sqlstr = `SELECT ` +
		`TABLE_NAME ` +
		`FROM INFORMATION_SCHEMA.TABLES ` +
		`WHERE TABLE_SCHEMA = ""`
	stmt := spanner.NewStatement(sqlstr)
	iter := client.Single().Query(ctx, stmt)
	defer iter.Stop()

	res := []*models.Table{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var t models.Table
		if err := row.ColumnByName("TABLE_NAME", &t.TableName); err != nil {
			return nil, err
		}

		res = append(res, &t)
	}

	return res, nil
}

// spanTableColumns runs a custom query, returning results as Column.
func spanTableColumns(client *spanner.Client, table string) ([]*models.Column, error) {
	ctx := context.Background()

	// sql query
	const sqlstr = `SELECT ` +
		`c.COLUMN_NAME, c.ORDINAL_POSITION, c.IS_NULLABLE, c.SPANNER_TYPE, ` +
		`EXISTS (` +
		`  SELECT 1 FROM INFORMATION_SCHEMA.INDEX_COLUMNS ic ` +
		`  WHERE ic.TABLE_SCHEMA = "" and ic.TABLE_NAME = c.TABLE_NAME ` +
		`  AND ic.COLUMN_NAME = c.COLUMN_NAME` +
		`  AND ic.INDEX_NAME = "PRIMARY_KEY" ` +
		`) IS_PRIMARY_KEY ` +
		`FROM INFORMATION_SCHEMA.COLUMNS c ` +
		`WHERE c.TABLE_SCHEMA = "" AND c.TABLE_NAME = @table ` +
		`ORDER BY c.ORDINAL_POSITION`

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["table"] = table
	iter := client.Single().Query(ctx, stmt)

	defer iter.Stop()

	res := []*models.Column{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var c models.Column
		var ord int64
		if err := row.ColumnByName("ORDINAL_POSITION", &ord); err != nil {
			return nil, err
		}
		c.FieldOrdinal = int(ord)
		if err := row.ColumnByName("COLUMN_NAME", &c.ColumnName); err != nil {
			return nil, err
		}
		var isNullable string
		if err := row.ColumnByName("IS_NULLABLE", &isNullable); err != nil {
			return nil, err
		}
		if isNullable == "NO" {
			c.NotNull = true
		}
		if err := row.ColumnByName("SPANNER_TYPE", &c.DataType); err != nil {
			return nil, err
		}
		if err := row.ColumnByName("IS_PRIMARY_KEY", &c.IsPrimaryKey); err != nil {
			return nil, err
		}

		res = append(res, &c)
	}

	return res, nil
}

// SpanTableColumns parses the query and generates a type for it.
func SpanTableColumns(client *spanner.Client, table string) ([]*models.Column, error) {
	return spanTableColumns(client, table)
}

func SpanTableIndexes(client *spanner.Client, table string) ([]*models.Index, error) {
	ctx := context.Background()

	// sql query
	const sqlstr = `SELECT ` +
		`INDEX_NAME, IS_UNIQUE ` +
		`FROM INFORMATION_SCHEMA.INDEXES ` +
		`WHERE TABLE_SCHEMA = "" AND INDEX_NAME != "PRIMARY_KEY" AND TABLE_NAME = @table `

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["table"] = table
	iter := client.Single().Query(ctx, stmt)

	defer iter.Stop()

	res := []*models.Index{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var i models.Index
		if err := row.ColumnByName("INDEX_NAME", &i.IndexName); err != nil {
			return nil, err
		}
		if err := row.ColumnByName("IS_UNIQUE", &i.IsUnique); err != nil {
			return nil, err
		}

		res = append(res, &i)
	}

	return res, nil
}

// SpanIndexColumns runs a custom query, returning results as IndexColumn.
func SpanIndexColumns(client *spanner.Client, table string, index string) ([]*models.IndexColumn, error) {
	ctx := context.Background()

	// sql query
	const sqlstr = `SELECT ` +
		`ORDINAL_POSITION, COLUMN_NAME ` +
		`FROM INFORMATION_SCHEMA.INDEX_COLUMNS ` +
		`WHERE TABLE_SCHEMA = "" AND INDEX_NAME = @index AND TABLE_NAME = @table ` +
		`ORDER BY ORDINAL_POSITION`

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["table"] = table
	stmt.Params["index"] = index
	iter := client.Single().Query(ctx, stmt)

	defer iter.Stop()

	res := []*models.IndexColumn{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var i models.IndexColumn
		var ord spanner.NullInt64
		if err := row.ColumnByName("ORDINAL_POSITION", &ord); err != nil {
			return nil, err
		}
		i.SeqNo = int(ord.Int64)
		if !ord.Valid {
			i.Storing = true
		}
		if err := row.ColumnByName("COLUMN_NAME", &i.ColumnName); err != nil {
			return nil, err
		}

		res = append(res, &i)
	}

	return res, nil
}

func SpanValidateCustomType(dataType string, customType string) bool {
	// No custom type validation now
	return true
}
