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

	"cloud.google.com/go/spanner"
	"go.mercari.io/yo/v2/models"
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

func (s *informationSchemaSource) TableList() ([]*models.Table, error) {
	ctx := context.Background()

	const sqlstr = `SELECT ` +
		`TABLE_NAME, PARENT_TABLE_NAME ` +
		`FROM INFORMATION_SCHEMA.TABLES ` +
		`WHERE TABLE_SCHEMA = "" ` +
		`ORDER BY TABLE_NAME`
	stmt := spanner.NewStatement(sqlstr)

	iter := s.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var res []*models.Table
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

		var parentTableName spanner.NullString
		if err := row.ColumnByName("PARENT_TABLE_NAME", &parentTableName); err != nil {
			return nil, err
		}
		t.ParentTableName = parentTableName.StringVal

		res = append(res, &t)
	}

	return res, nil
}

func (s *informationSchemaSource) ColumnList(table string) ([]*models.Column, error) {
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

	iter := s.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var res []*models.Column
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

func (s *informationSchemaSource) IndexList(table string) ([]*models.Index, error) {
	ctx := context.Background()

	// sql query
	const sqlstr = `SELECT ` +
		`INDEX_NAME, IS_UNIQUE ` +
		`FROM INFORMATION_SCHEMA.INDEXES ` +
		`WHERE TABLE_SCHEMA = "" AND INDEX_NAME != "PRIMARY_KEY" AND TABLE_NAME = @table `

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["table"] = table

	iter := s.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var res []*models.Index
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

func (s *informationSchemaSource) IndexColumnList(table string, index string) ([]*models.IndexColumn, error) {
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

	var res []*models.IndexColumn
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
