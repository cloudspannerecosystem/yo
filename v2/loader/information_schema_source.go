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
	"os"
	"sort"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
)

func NewInformationSchemaSource(client *spanner.Client) (SchemaSource, error) {
	return &informationSchemaSource{
		client: client,
	}, nil
}

type informationSchemaSource struct {
	client *spanner.Client
}

func (s *informationSchemaSource) TableList() ([]*SpannerTable, error) {
	ctx := context.Background()

	const sqlstr = `SELECT ` +
		`TABLE_NAME, PARENT_TABLE_NAME ` +
		`FROM INFORMATION_SCHEMA.TABLES ` +
		`WHERE TABLE_SCHEMA = "" ` +
		`ORDER BY TABLE_NAME`
	stmt := spanner.NewStatement(sqlstr)

	iter := s.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var res []*SpannerTable
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var t SpannerTable
		if err := row.ColumnByName("TABLE_NAME", &t.TableName); err != nil {
			return nil, err
		}

		var parentTableName spanner.NullString
		if err := row.ColumnByName("PARENT_TABLE_NAME", &parentTableName); err != nil {
			return nil, err
		}
		t.ParentTableName = parentTableName.StringVal

		res = append(res, &t)
	}

	return res, nil
}

func (s *informationSchemaSource) ColumnList(table string) ([]*SpannerColumn, error) {
	ctx := context.Background()

	// sql query
	const sqlstr = `SELECT ` +
		`c.COLUMN_NAME, c.ORDINAL_POSITION, c.IS_NULLABLE, c.SPANNER_TYPE, ` +
		`EXISTS (` +
		`  SELECT 1 FROM INFORMATION_SCHEMA.INDEX_COLUMNS ic ` +
		`  WHERE ic.TABLE_SCHEMA = "" and ic.TABLE_NAME = c.TABLE_NAME ` +
		`  AND ic.COLUMN_NAME = c.COLUMN_NAME` +
		`  AND ic.INDEX_NAME = "PRIMARY_KEY" ` +
		`) IS_PRIMARY_KEY, ` +
		`IS_GENERATED = "ALWAYS" AS IS_GENERATED ` +
		`FROM INFORMATION_SCHEMA.COLUMNS c ` +
		`WHERE c.TABLE_SCHEMA = "" AND c.TABLE_NAME = @table ` +
		`ORDER BY c.ORDINAL_POSITION`

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["table"] = table

	iter := s.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var storing bool
	var res []*SpannerColumn
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var c SpannerColumn
		var ord spanner.NullInt64
		if err := row.ColumnByName("ORDINAL_POSITION", &ord); err != nil {
			return nil, err
		}
		if !ord.Valid {
			storing = true
		}
		c.FieldOrdinal = int(ord.Int64)
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
		if err := row.ColumnByName("IS_GENERATED", &c.IsGenerated); err != nil {
			return nil, err
		}

		res = append(res, &c)
	}

	// Since the value of ORDINAL_POSITION is NULL for the STORING column, the order is undetermined.
	// Spanner Instances are implicitly returned in the order of their definition, but the Spanner Emulator's specifications make the order random.
	// Currently, the Spanner Instance's Information Schema and DDL return the results in the order expected by the developer.
	// For this reason, only when using the Spanner Emulator are we sorted by column name to fix the order.
	// In reality, using the Spanner Emulator's Information Schema is not recommended, and this is a measure only for Unit Testing.
	// https://github.com/cloudspannerecosystem/yo/issues/154
	if os.Getenv("SPANNER_EMULATOR_HOST") != "" && storing {
		sort.Slice(res, func(i, j int) bool {
			return res[i].ColumnName < res[j].ColumnName
		})
	}

	return res, nil
}

func (s *informationSchemaSource) IndexList(table string) ([]*SpannerIndex, error) {
	ctx := context.Background()

	// sql query
	const sqlstr = `SELECT ` +
		`INDEX_NAME, IS_UNIQUE ` +
		`FROM INFORMATION_SCHEMA.INDEXES ` +
		`WHERE TABLE_SCHEMA = "" ` +
		`AND INDEX_NAME != "PRIMARY_KEY" ` +
		`AND TABLE_NAME = @table ` +
		`AND SPANNER_IS_MANAGED = FALSE `

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["table"] = table

	iter := s.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var res []*SpannerIndex
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var i SpannerIndex
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

func (s *informationSchemaSource) IndexColumnList(table string, index string) ([]*SpannerIndexColumn, error) {
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

	iter := s.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var res []*SpannerIndexColumn
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var i SpannerIndexColumn
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
