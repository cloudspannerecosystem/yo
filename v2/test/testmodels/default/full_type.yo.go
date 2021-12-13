// Code generated by yo. DO NOT EDIT.

// Package models contains the types.
package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
)

// FullType represents a row from 'FullTypes'.
type FullType struct {
	PKey                 string              `spanner:"PKey" json:"PKey"`                                 // PKey
	FTString             string              `spanner:"FTString" json:"FTString"`                         // FTString
	FTStringNull         spanner.NullString  `spanner:"FTStringNull" json:"FTStringNull"`                 // FTStringNull
	FTBool               bool                `spanner:"FTBool" json:"FTBool"`                             // FTBool
	FTBoolNull           spanner.NullBool    `spanner:"FTBoolNull" json:"FTBoolNull"`                     // FTBoolNull
	FTBytes              []byte              `spanner:"FTBytes" json:"FTBytes"`                           // FTBytes
	FTBytesNull          []byte              `spanner:"FTBytesNull" json:"FTBytesNull"`                   // FTBytesNull
	FTTimestamp          time.Time           `spanner:"FTTimestamp" json:"FTTimestamp"`                   // FTTimestamp
	FTTimestampNull      spanner.NullTime    `spanner:"FTTimestampNull" json:"FTTimestampNull"`           // FTTimestampNull
	FTInt                int64               `spanner:"FTInt" json:"FTInt"`                               // FTInt
	FTIntNull            spanner.NullInt64   `spanner:"FTIntNull" json:"FTIntNull"`                       // FTIntNull
	FTFloat              float64             `spanner:"FTFloat" json:"FTFloat"`                           // FTFloat
	FTFloatNull          spanner.NullFloat64 `spanner:"FTFloatNull" json:"FTFloatNull"`                   // FTFloatNull
	FTDate               civil.Date          `spanner:"FTDate" json:"FTDate"`                             // FTDate
	FTDateNull           spanner.NullDate    `spanner:"FTDateNull" json:"FTDateNull"`                     // FTDateNull
	FTArrayStringNull    []string            `spanner:"FTArrayStringNull" json:"FTArrayStringNull"`       // FTArrayStringNull
	FTArrayString        []string            `spanner:"FTArrayString" json:"FTArrayString"`               // FTArrayString
	FTArrayBoolNull      []bool              `spanner:"FTArrayBoolNull" json:"FTArrayBoolNull"`           // FTArrayBoolNull
	FTArrayBool          []bool              `spanner:"FTArrayBool" json:"FTArrayBool"`                   // FTArrayBool
	FTArrayBytesNull     [][]byte            `spanner:"FTArrayBytesNull" json:"FTArrayBytesNull"`         // FTArrayBytesNull
	FTArrayBytes         [][]byte            `spanner:"FTArrayBytes" json:"FTArrayBytes"`                 // FTArrayBytes
	FTArrayTimestampNull []time.Time         `spanner:"FTArrayTimestampNull" json:"FTArrayTimestampNull"` // FTArrayTimestampNull
	FTArrayTimestamp     []time.Time         `spanner:"FTArrayTimestamp" json:"FTArrayTimestamp"`         // FTArrayTimestamp
	FTArrayIntNull       []int64             `spanner:"FTArrayIntNull" json:"FTArrayIntNull"`             // FTArrayIntNull
	FTArrayInt           []int64             `spanner:"FTArrayInt" json:"FTArrayInt"`                     // FTArrayInt
	FTArrayFloatNull     []float64           `spanner:"FTArrayFloatNull" json:"FTArrayFloatNull"`         // FTArrayFloatNull
	FTArrayFloat         []float64           `spanner:"FTArrayFloat" json:"FTArrayFloat"`                 // FTArrayFloat
	FTArrayDateNull      []civil.Date        `spanner:"FTArrayDateNull" json:"FTArrayDateNull"`           // FTArrayDateNull
	FTArrayDate          []civil.Date        `spanner:"FTArrayDate" json:"FTArrayDate"`                   // FTArrayDate
}

func FullTypePrimaryKeys() []string {
	return []string{
		"PKey",
	}
}

func FullTypeColumns() []string {
	return []string{
		"PKey",
		"FTString",
		"FTStringNull",
		"FTBool",
		"FTBoolNull",
		"FTBytes",
		"FTBytesNull",
		"FTTimestamp",
		"FTTimestampNull",
		"FTInt",
		"FTIntNull",
		"FTFloat",
		"FTFloatNull",
		"FTDate",
		"FTDateNull",
		"FTArrayStringNull",
		"FTArrayString",
		"FTArrayBoolNull",
		"FTArrayBool",
		"FTArrayBytesNull",
		"FTArrayBytes",
		"FTArrayTimestampNull",
		"FTArrayTimestamp",
		"FTArrayIntNull",
		"FTArrayInt",
		"FTArrayFloatNull",
		"FTArrayFloat",
		"FTArrayDateNull",
		"FTArrayDate",
	}
}

func FullTypeWritableColumns() []string {
	return []string{
		"PKey",
		"FTString",
		"FTStringNull",
		"FTBool",
		"FTBoolNull",
		"FTBytes",
		"FTBytesNull",
		"FTTimestamp",
		"FTTimestampNull",
		"FTInt",
		"FTIntNull",
		"FTFloat",
		"FTFloatNull",
		"FTDate",
		"FTDateNull",
		"FTArrayStringNull",
		"FTArrayString",
		"FTArrayBoolNull",
		"FTArrayBool",
		"FTArrayBytesNull",
		"FTArrayBytes",
		"FTArrayTimestampNull",
		"FTArrayTimestamp",
		"FTArrayIntNull",
		"FTArrayInt",
		"FTArrayFloatNull",
		"FTArrayFloat",
		"FTArrayDateNull",
		"FTArrayDate",
	}
}

func (ft *FullType) columnsToPtrs(cols []string) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		switch col {
		case "PKey":
			ret = append(ret, yoDecode(&ft.PKey))
		case "FTString":
			ret = append(ret, yoDecode(&ft.FTString))
		case "FTStringNull":
			ret = append(ret, yoDecode(&ft.FTStringNull))
		case "FTBool":
			ret = append(ret, yoDecode(&ft.FTBool))
		case "FTBoolNull":
			ret = append(ret, yoDecode(&ft.FTBoolNull))
		case "FTBytes":
			ret = append(ret, yoDecode(&ft.FTBytes))
		case "FTBytesNull":
			ret = append(ret, yoDecode(&ft.FTBytesNull))
		case "FTTimestamp":
			ret = append(ret, yoDecode(&ft.FTTimestamp))
		case "FTTimestampNull":
			ret = append(ret, yoDecode(&ft.FTTimestampNull))
		case "FTInt":
			ret = append(ret, yoDecode(&ft.FTInt))
		case "FTIntNull":
			ret = append(ret, yoDecode(&ft.FTIntNull))
		case "FTFloat":
			ret = append(ret, yoDecode(&ft.FTFloat))
		case "FTFloatNull":
			ret = append(ret, yoDecode(&ft.FTFloatNull))
		case "FTDate":
			ret = append(ret, yoDecode(&ft.FTDate))
		case "FTDateNull":
			ret = append(ret, yoDecode(&ft.FTDateNull))
		case "FTArrayStringNull":
			ret = append(ret, yoDecode(&ft.FTArrayStringNull))
		case "FTArrayString":
			ret = append(ret, yoDecode(&ft.FTArrayString))
		case "FTArrayBoolNull":
			ret = append(ret, yoDecode(&ft.FTArrayBoolNull))
		case "FTArrayBool":
			ret = append(ret, yoDecode(&ft.FTArrayBool))
		case "FTArrayBytesNull":
			ret = append(ret, yoDecode(&ft.FTArrayBytesNull))
		case "FTArrayBytes":
			ret = append(ret, yoDecode(&ft.FTArrayBytes))
		case "FTArrayTimestampNull":
			ret = append(ret, yoDecode(&ft.FTArrayTimestampNull))
		case "FTArrayTimestamp":
			ret = append(ret, yoDecode(&ft.FTArrayTimestamp))
		case "FTArrayIntNull":
			ret = append(ret, yoDecode(&ft.FTArrayIntNull))
		case "FTArrayInt":
			ret = append(ret, yoDecode(&ft.FTArrayInt))
		case "FTArrayFloatNull":
			ret = append(ret, yoDecode(&ft.FTArrayFloatNull))
		case "FTArrayFloat":
			ret = append(ret, yoDecode(&ft.FTArrayFloat))
		case "FTArrayDateNull":
			ret = append(ret, yoDecode(&ft.FTArrayDateNull))
		case "FTArrayDate":
			ret = append(ret, yoDecode(&ft.FTArrayDate))
		default:
			return nil, fmt.Errorf("unknown column: %s", col)
		}
	}
	return ret, nil
}

func (ft *FullType) columnsToValues(cols []string) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		switch col {
		case "PKey":
			ret = append(ret, yoEncode(ft.PKey))
		case "FTString":
			ret = append(ret, yoEncode(ft.FTString))
		case "FTStringNull":
			ret = append(ret, yoEncode(ft.FTStringNull))
		case "FTBool":
			ret = append(ret, yoEncode(ft.FTBool))
		case "FTBoolNull":
			ret = append(ret, yoEncode(ft.FTBoolNull))
		case "FTBytes":
			ret = append(ret, yoEncode(ft.FTBytes))
		case "FTBytesNull":
			ret = append(ret, yoEncode(ft.FTBytesNull))
		case "FTTimestamp":
			ret = append(ret, yoEncode(ft.FTTimestamp))
		case "FTTimestampNull":
			ret = append(ret, yoEncode(ft.FTTimestampNull))
		case "FTInt":
			ret = append(ret, yoEncode(ft.FTInt))
		case "FTIntNull":
			ret = append(ret, yoEncode(ft.FTIntNull))
		case "FTFloat":
			ret = append(ret, yoEncode(ft.FTFloat))
		case "FTFloatNull":
			ret = append(ret, yoEncode(ft.FTFloatNull))
		case "FTDate":
			ret = append(ret, yoEncode(ft.FTDate))
		case "FTDateNull":
			ret = append(ret, yoEncode(ft.FTDateNull))
		case "FTArrayStringNull":
			ret = append(ret, yoEncode(ft.FTArrayStringNull))
		case "FTArrayString":
			ret = append(ret, yoEncode(ft.FTArrayString))
		case "FTArrayBoolNull":
			ret = append(ret, yoEncode(ft.FTArrayBoolNull))
		case "FTArrayBool":
			ret = append(ret, yoEncode(ft.FTArrayBool))
		case "FTArrayBytesNull":
			ret = append(ret, yoEncode(ft.FTArrayBytesNull))
		case "FTArrayBytes":
			ret = append(ret, yoEncode(ft.FTArrayBytes))
		case "FTArrayTimestampNull":
			ret = append(ret, yoEncode(ft.FTArrayTimestampNull))
		case "FTArrayTimestamp":
			ret = append(ret, yoEncode(ft.FTArrayTimestamp))
		case "FTArrayIntNull":
			ret = append(ret, yoEncode(ft.FTArrayIntNull))
		case "FTArrayInt":
			ret = append(ret, yoEncode(ft.FTArrayInt))
		case "FTArrayFloatNull":
			ret = append(ret, yoEncode(ft.FTArrayFloatNull))
		case "FTArrayFloat":
			ret = append(ret, yoEncode(ft.FTArrayFloat))
		case "FTArrayDateNull":
			ret = append(ret, yoEncode(ft.FTArrayDateNull))
		case "FTArrayDate":
			ret = append(ret, yoEncode(ft.FTArrayDate))
		default:
			return nil, fmt.Errorf("unknown column: %s", col)
		}
	}

	return ret, nil
}

// newFullType_Decoder returns a decoder which reads a row from *spanner.Row
// into FullType. The decoder is not goroutine-safe. Don't use it concurrently.
func newFullType_Decoder(cols []string) func(*spanner.Row) (*FullType, error) {
	return func(row *spanner.Row) (*FullType, error) {
		var ft FullType
		ptrs, err := ft.columnsToPtrs(cols)
		if err != nil {
			return nil, err
		}

		if err := row.Columns(ptrs...); err != nil {
			return nil, err
		}

		return &ft, nil
	}
}

// Insert returns a Mutation to insert a row into a table. If the row already
// exists, the write or transaction fails.
func (ft *FullType) Insert(ctx context.Context) *spanner.Mutation {
	values, _ := ft.columnsToValues(FullTypeWritableColumns())
	return spanner.Insert("FullTypes", FullTypeWritableColumns(), values)
}

// Update returns a Mutation to update a row in a table. If the row does not
// already exist, the write or transaction fails.
func (ft *FullType) Update(ctx context.Context) *spanner.Mutation {
	values, _ := ft.columnsToValues(FullTypeWritableColumns())
	return spanner.Update("FullTypes", FullTypeWritableColumns(), values)
}

// InsertOrUpdate returns a Mutation to insert a row into a table. If the row
// already exists, it updates it instead. Any column values not explicitly
// written are preserved.
func (ft *FullType) InsertOrUpdate(ctx context.Context) *spanner.Mutation {
	values, _ := ft.columnsToValues(FullTypeWritableColumns())
	return spanner.InsertOrUpdate("FullTypes", FullTypeWritableColumns(), values)
}

// Replace returns a Mutation to insert a row into a table, deleting any
// existing row. Unlike InsertOrUpdate, this means any values not explicitly
// written become NULL.
func (ft *FullType) Replace(ctx context.Context) *spanner.Mutation {
	values, _ := ft.columnsToValues(FullTypeWritableColumns())
	return spanner.Replace("FullTypes", FullTypeWritableColumns(), values)
}

// UpdateColumns returns a Mutation to update specified columns of a row in a table.
func (ft *FullType) UpdateColumns(ctx context.Context, cols ...string) (*spanner.Mutation, error) {
	// add primary keys to columns to update by primary keys
	colsWithPKeys := append(cols, FullTypePrimaryKeys()...)

	values, err := ft.columnsToValues(colsWithPKeys)
	if err != nil {
		return nil, newErrorWithCode(codes.InvalidArgument, "FullType.UpdateColumns", "FullTypes", err)
	}

	return spanner.Update("FullTypes", colsWithPKeys, values), nil
}

// FindFullType gets a FullType by primary key
func FindFullType(ctx context.Context, db YODB, pKey string) (*FullType, error) {
	key := spanner.Key{yoEncode(pKey)}
	row, err := db.ReadRow(ctx, "FullTypes", key, FullTypeColumns())
	if err != nil {
		return nil, newError("FindFullType", "FullTypes", err)
	}

	decoder := newFullType_Decoder(FullTypeColumns())
	ft, err := decoder(row)
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "FindFullType", "FullTypes", err)
	}

	return ft, nil
}

// ReadFullType retrieves multiples rows from FullType by KeySet as a slice.
func ReadFullType(ctx context.Context, db YODB, keys spanner.KeySet) ([]*FullType, error) {
	var res []*FullType

	decoder := newFullType_Decoder(FullTypeColumns())

	rows := db.Read(ctx, "FullTypes", keys, FullTypeColumns())
	err := rows.Do(func(row *spanner.Row) error {
		ft, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, ft)

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "ReadFullType", "FullTypes", err)
	}

	return res, nil
}

// Delete deletes the FullType from the database.
func (ft *FullType) Delete(ctx context.Context) *spanner.Mutation {
	values, _ := ft.columnsToValues(FullTypePrimaryKeys())
	return spanner.Delete("FullTypes", spanner.Key(values))
}

// FindFullTypeByFullTypesByFTString retrieves a row from 'FullTypes' as a FullType.
//
// If no row is present with the given key, then ReadRow returns an error where
// spanner.ErrCode(err) is codes.NotFound.
//
// Generated from unique index 'FullTypesByFTString'.
func FindFullTypeByFullTypesByFTString(ctx context.Context, db YODB, fTString string) (*FullType, error) {
	const sqlstr = "SELECT " +
		"PKey, FTString, FTStringNull, FTBool, FTBoolNull, FTBytes, FTBytesNull, FTTimestamp, FTTimestampNull, FTInt, FTIntNull, FTFloat, FTFloatNull, FTDate, FTDateNull, FTArrayStringNull, FTArrayString, FTArrayBoolNull, FTArrayBool, FTArrayBytesNull, FTArrayBytes, FTArrayTimestampNull, FTArrayTimestamp, FTArrayIntNull, FTArrayInt, FTArrayFloatNull, FTArrayFloat, FTArrayDateNull, FTArrayDate " +
		"FROM FullTypes@{FORCE_INDEX=FullTypesByFTString} " +
		"WHERE FTString = @param0"

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["param0"] = yoEncode(fTString)

	decoder := newFullType_Decoder(FullTypeColumns())

	// run query
	YOLog(ctx, sqlstr, fTString)
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, newErrorWithCode(codes.NotFound, "FindFullTypeByFullTypesByFTString", "FullTypes", err)
		}
		return nil, newError("FindFullTypeByFullTypesByFTString", "FullTypes", err)
	}

	ft, err := decoder(row)
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "FindFullTypeByFullTypesByFTString", "FullTypes", err)
	}

	return ft, nil
}

// ReadFullTypeByFullTypesByFTString retrieves multiples rows from 'FullTypes' by KeySet as a slice.
//
// This does not retrives all columns of 'FullTypes' because an index has only columns
// used for primary key, index key and storing columns. If you need more columns, add storing
// columns or Read by primary key or Query with join.
//
// Generated from unique index 'FullTypesByFTString'.
func ReadFullTypeByFullTypesByFTString(ctx context.Context, db YODB, keys spanner.KeySet) ([]*FullType, error) {
	var res []*FullType
	columns := []string{
		"PKey",
		"FTString",
	}

	decoder := newFullType_Decoder(columns)

	rows := db.ReadUsingIndex(ctx, "FullTypes", "FullTypesByFTString", keys, columns)
	err := rows.Do(func(row *spanner.Row) error {
		ft, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, ft)

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "ReadFullTypeByFullTypesByFTString", "FullTypes", err)
	}

	return res, nil
}

// FindFullTypesByFullTypesByInTimestampNull retrieves multiple rows from 'FullTypes' as a slice of FullType.
//
// Generated from index 'FullTypesByInTimestampNull'.
func FindFullTypesByFullTypesByInTimestampNull(ctx context.Context, db YODB, fTInt int64, fTTimestampNull spanner.NullTime) ([]*FullType, error) {
	var sqlstr = "SELECT " +
		"PKey, FTString, FTStringNull, FTBool, FTBoolNull, FTBytes, FTBytesNull, FTTimestamp, FTTimestampNull, FTInt, FTIntNull, FTFloat, FTFloatNull, FTDate, FTDateNull, FTArrayStringNull, FTArrayString, FTArrayBoolNull, FTArrayBool, FTArrayBytesNull, FTArrayBytes, FTArrayTimestampNull, FTArrayTimestamp, FTArrayIntNull, FTArrayInt, FTArrayFloatNull, FTArrayFloat, FTArrayDateNull, FTArrayDate " +
		"FROM FullTypes@{FORCE_INDEX=FullTypesByInTimestampNull} "

	conds := make([]string, 2)
	conds[0] = "FTInt = @param0"
	if fTTimestampNull.IsNull() {
		conds[1] = "FTTimestampNull IS NULL"
	} else {
		conds[1] = "FTTimestampNull = @param1"
	}
	sqlstr += "WHERE " + strings.Join(conds, " AND ")

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["param0"] = yoEncode(fTInt)
	stmt.Params["param1"] = yoEncode(fTTimestampNull)

	decoder := newFullType_Decoder(FullTypeColumns())

	// run query
	YOLog(ctx, sqlstr, fTInt, fTTimestampNull)
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	// load results
	res := []*FullType{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, newError("FindFullTypesByFullTypesByInTimestampNull", "FullTypes", err)
		}

		ft, err := decoder(row)
		if err != nil {
			return nil, newErrorWithCode(codes.Internal, "FindFullTypesByFullTypesByInTimestampNull", "FullTypes", err)
		}

		res = append(res, ft)
	}

	return res, nil
}

// ReadFullTypesByFullTypesByInTimestampNull retrieves multiples rows from 'FullTypes' by KeySet as a slice.
//
// This does not retrives all columns of 'FullTypes' because an index has only columns
// used for primary key, index key and storing columns. If you need more columns, add storing
// columns or Read by primary key or Query with join.
//
// Generated from unique index 'FullTypesByInTimestampNull'.
func ReadFullTypesByFullTypesByInTimestampNull(ctx context.Context, db YODB, keys spanner.KeySet) ([]*FullType, error) {
	var res []*FullType
	columns := []string{
		"PKey",
		"FTInt",
		"FTTimestampNull",
	}

	decoder := newFullType_Decoder(columns)

	rows := db.ReadUsingIndex(ctx, "FullTypes", "FullTypesByInTimestampNull", keys, columns)
	err := rows.Do(func(row *spanner.Row) error {
		ft, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, ft)

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "ReadFullTypesByFullTypesByInTimestampNull", "FullTypes", err)
	}

	return res, nil
}

// FindFullTypesByFullTypesByIntDate retrieves multiple rows from 'FullTypes' as a slice of FullType.
//
// Generated from index 'FullTypesByIntDate'.
func FindFullTypesByFullTypesByIntDate(ctx context.Context, db YODB, fTInt int64, fTDate civil.Date) ([]*FullType, error) {
	const sqlstr = "SELECT " +
		"PKey, FTString, FTStringNull, FTBool, FTBoolNull, FTBytes, FTBytesNull, FTTimestamp, FTTimestampNull, FTInt, FTIntNull, FTFloat, FTFloatNull, FTDate, FTDateNull, FTArrayStringNull, FTArrayString, FTArrayBoolNull, FTArrayBool, FTArrayBytesNull, FTArrayBytes, FTArrayTimestampNull, FTArrayTimestamp, FTArrayIntNull, FTArrayInt, FTArrayFloatNull, FTArrayFloat, FTArrayDateNull, FTArrayDate " +
		"FROM FullTypes@{FORCE_INDEX=FullTypesByIntDate} " +
		"WHERE FTInt = @param0 AND FTDate = @param1"

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["param0"] = yoEncode(fTInt)
	stmt.Params["param1"] = yoEncode(fTDate)

	decoder := newFullType_Decoder(FullTypeColumns())

	// run query
	YOLog(ctx, sqlstr, fTInt, fTDate)
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	// load results
	res := []*FullType{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, newError("FindFullTypesByFullTypesByIntDate", "FullTypes", err)
		}

		ft, err := decoder(row)
		if err != nil {
			return nil, newErrorWithCode(codes.Internal, "FindFullTypesByFullTypesByIntDate", "FullTypes", err)
		}

		res = append(res, ft)
	}

	return res, nil
}

// ReadFullTypesByFullTypesByIntDate retrieves multiples rows from 'FullTypes' by KeySet as a slice.
//
// This does not retrives all columns of 'FullTypes' because an index has only columns
// used for primary key, index key and storing columns. If you need more columns, add storing
// columns or Read by primary key or Query with join.
//
// Generated from unique index 'FullTypesByIntDate'.
func ReadFullTypesByFullTypesByIntDate(ctx context.Context, db YODB, keys spanner.KeySet) ([]*FullType, error) {
	var res []*FullType
	columns := []string{
		"PKey",
		"FTInt",
		"FTDate",
	}

	decoder := newFullType_Decoder(columns)

	rows := db.ReadUsingIndex(ctx, "FullTypes", "FullTypesByIntDate", keys, columns)
	err := rows.Do(func(row *spanner.Row) error {
		ft, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, ft)

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "ReadFullTypesByFullTypesByIntDate", "FullTypes", err)
	}

	return res, nil
}

// FindFullTypesByFullTypesByIntTimestamp retrieves multiple rows from 'FullTypes' as a slice of FullType.
//
// Generated from index 'FullTypesByIntTimestamp'.
func FindFullTypesByFullTypesByIntTimestamp(ctx context.Context, db YODB, fTInt int64, fTTimestamp time.Time) ([]*FullType, error) {
	const sqlstr = "SELECT " +
		"PKey, FTString, FTStringNull, FTBool, FTBoolNull, FTBytes, FTBytesNull, FTTimestamp, FTTimestampNull, FTInt, FTIntNull, FTFloat, FTFloatNull, FTDate, FTDateNull, FTArrayStringNull, FTArrayString, FTArrayBoolNull, FTArrayBool, FTArrayBytesNull, FTArrayBytes, FTArrayTimestampNull, FTArrayTimestamp, FTArrayIntNull, FTArrayInt, FTArrayFloatNull, FTArrayFloat, FTArrayDateNull, FTArrayDate " +
		"FROM FullTypes@{FORCE_INDEX=FullTypesByIntTimestamp} " +
		"WHERE FTInt = @param0 AND FTTimestamp = @param1"

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["param0"] = yoEncode(fTInt)
	stmt.Params["param1"] = yoEncode(fTTimestamp)

	decoder := newFullType_Decoder(FullTypeColumns())

	// run query
	YOLog(ctx, sqlstr, fTInt, fTTimestamp)
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	// load results
	res := []*FullType{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, newError("FindFullTypesByFullTypesByIntTimestamp", "FullTypes", err)
		}

		ft, err := decoder(row)
		if err != nil {
			return nil, newErrorWithCode(codes.Internal, "FindFullTypesByFullTypesByIntTimestamp", "FullTypes", err)
		}

		res = append(res, ft)
	}

	return res, nil
}

// ReadFullTypesByFullTypesByIntTimestamp retrieves multiples rows from 'FullTypes' by KeySet as a slice.
//
// This does not retrives all columns of 'FullTypes' because an index has only columns
// used for primary key, index key and storing columns. If you need more columns, add storing
// columns or Read by primary key or Query with join.
//
// Generated from unique index 'FullTypesByIntTimestamp'.
func ReadFullTypesByFullTypesByIntTimestamp(ctx context.Context, db YODB, keys spanner.KeySet) ([]*FullType, error) {
	var res []*FullType
	columns := []string{
		"PKey",
		"FTInt",
		"FTTimestamp",
	}

	decoder := newFullType_Decoder(columns)

	rows := db.ReadUsingIndex(ctx, "FullTypes", "FullTypesByIntTimestamp", keys, columns)
	err := rows.Do(func(row *spanner.Row) error {
		ft, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, ft)

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "ReadFullTypesByFullTypesByIntTimestamp", "FullTypes", err)
	}

	return res, nil
}

// FindFullTypesByFullTypesByTimestamp retrieves multiple rows from 'FullTypes' as a slice of FullType.
//
// Generated from index 'FullTypesByTimestamp'.
func FindFullTypesByFullTypesByTimestamp(ctx context.Context, db YODB, fTTimestamp time.Time) ([]*FullType, error) {
	const sqlstr = "SELECT " +
		"PKey, FTString, FTStringNull, FTBool, FTBoolNull, FTBytes, FTBytesNull, FTTimestamp, FTTimestampNull, FTInt, FTIntNull, FTFloat, FTFloatNull, FTDate, FTDateNull, FTArrayStringNull, FTArrayString, FTArrayBoolNull, FTArrayBool, FTArrayBytesNull, FTArrayBytes, FTArrayTimestampNull, FTArrayTimestamp, FTArrayIntNull, FTArrayInt, FTArrayFloatNull, FTArrayFloat, FTArrayDateNull, FTArrayDate " +
		"FROM FullTypes@{FORCE_INDEX=FullTypesByTimestamp} " +
		"WHERE FTTimestamp = @param0"

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["param0"] = yoEncode(fTTimestamp)

	decoder := newFullType_Decoder(FullTypeColumns())

	// run query
	YOLog(ctx, sqlstr, fTTimestamp)
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	// load results
	res := []*FullType{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, newError("FindFullTypesByFullTypesByTimestamp", "FullTypes", err)
		}

		ft, err := decoder(row)
		if err != nil {
			return nil, newErrorWithCode(codes.Internal, "FindFullTypesByFullTypesByTimestamp", "FullTypes", err)
		}

		res = append(res, ft)
	}

	return res, nil
}

// ReadFullTypesByFullTypesByTimestamp retrieves multiples rows from 'FullTypes' by KeySet as a slice.
//
// This does not retrives all columns of 'FullTypes' because an index has only columns
// used for primary key, index key and storing columns. If you need more columns, add storing
// columns or Read by primary key or Query with join.
//
// Generated from unique index 'FullTypesByTimestamp'.
func ReadFullTypesByFullTypesByTimestamp(ctx context.Context, db YODB, keys spanner.KeySet) ([]*FullType, error) {
	var res []*FullType
	columns := []string{
		"PKey",
		"FTTimestamp",
	}

	decoder := newFullType_Decoder(columns)

	rows := db.ReadUsingIndex(ctx, "FullTypes", "FullTypesByTimestamp", keys, columns)
	err := rows.Do(func(row *spanner.Row) error {
		ft, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, ft)

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "ReadFullTypesByFullTypesByTimestamp", "FullTypes", err)
	}

	return res, nil
}