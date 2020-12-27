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
	"sort"
	"strings"

	"go.mercari.io/yo/v2/config"
	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/models"
)

type Option struct {
	Config       *config.Config
	IgnoreFields []string
	IgnoreTables []string
}

type SchemaSource interface {
	TableList() ([]*models.Table, error)
	ColumnList(string) ([]*models.Column, error)
	IndexList(string) ([]*models.Index, error)
	IndexColumnList(string, string) ([]*models.IndexColumn, error)
}

func NewTypeLoader(source SchemaSource, inflector internal.Inflector, opt Option) *TypeLoader {
	return &TypeLoader{
		source:       source,
		inflector:    inflector,
		config:       opt.Config,
		ignoreFields: opt.IgnoreFields,
		ignoreTables: opt.IgnoreTables,
	}
}

// TypeLoader provides a common Loader implementation used by the built in
// schema/query loaders.
type TypeLoader struct {
	source    SchemaSource
	inflector internal.Inflector

	config       *config.Config
	ignoreFields []string
	ignoreTables []string
}

// NthParam satisifies Loader's NthParam.
func (tl *TypeLoader) NthParam(i int) string {
	return fmt.Sprintf("@param%d", i)
}

// Mask returns the parameter mask.
func (tl *TypeLoader) Mask() string {
	return "?"
}

func (tl *TypeLoader) validateCustomType(dataType string, customType string) bool {
	return true
}

// LoadSchema loads schema definitions.
func (tl *TypeLoader) LoadSchema() (*internal.Schema, error) {
	// load tables
	tableMap, err := tl.LoadTable()
	if err != nil {
		return nil, err
	}

	// load indexes
	ixMap, err := tl.LoadIndexes(tableMap)
	if err != nil {
		return nil, err
	}

	setIndexesToTables(tableMap, ixMap)

	tables := make([]*internal.Type, 0, len(tableMap))
	for _, tbl := range tableMap {
		tables = append(tables, tbl)
	}

	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	return &internal.Schema{
		Types: tables,
	}, nil
}

// LoadTable loads a schema table/view definition.
func (tl *TypeLoader) LoadTable() (map[string]*internal.Type, error) {
	var err error

	// load tables
	tableList, err := tl.source.TableList()
	if err != nil {
		return nil, err
	}

	// tables
	tableMap := make(map[string]*internal.Type)
	for _, ti := range tableList {
		ignore := false

		for _, ignoreTable := range tl.ignoreTables {
			if ignoreTable == ti.TableName {
				// Skip adding this table if user has specified they are not
				// interested.
				//
				// This could be useful for tables which are managed by the
				// database (e.g. SchemaMigrations) instead of
				// via Go code.
				ignore = true
			}
		}

		if ignore {
			continue
		}

		// create template
		typeTpl := &internal.Type{
			Name:   internal.SingularizeIdentifier(tl.inflector, ti.TableName),
			Schema: "",
			Fields: []*internal.Field{},
			Table:  ti,
		}

		// process columns
		err = tl.LoadColumns(typeTpl)
		if err != nil {
			return nil, err
		}

		if err := tl.loadPrimaryKeys(typeTpl); err != nil {
			return nil, err
		}

		tableMap[ti.TableName] = typeTpl
	}

	return tableMap, nil
}

// loadPrimaryKeys loads primary key fields
func (tl *TypeLoader) loadPrimaryKeys(typeTpl *internal.Type) error {
	// reorder primary keys
	indexCols, err := tl.source.IndexColumnList(typeTpl.Table.TableName, "PRIMARY_KEY")
	if err != nil {
		panic(err)
	}

	var fields []*internal.Field
	for _, idx := range indexCols {
		var field *internal.Field
		for _, f := range typeTpl.Fields {
			if f.Col.ColumnName == idx.ColumnName {
				field = f
				break
			}
		}

		if field == nil {
			return fmt.Errorf("primary key column is not found in column list: table=%v column=%v",
				typeTpl.Name, idx.ColumnName,
			)
		}
		fields = append(fields, field)
	}

	typeTpl.PrimaryKeyFields = fields
	return nil
}

// tableCustomTypes find custom type definitions of the table
func (tl *TypeLoader) tableCustomTypes(table string) map[string]string {
	columnTypes := make(map[string]string)
	for _, tbl := range tl.config.Tables {
		if tbl.Name != table {
			continue
		}

		for _, col := range tbl.Columns {
			columnTypes[col.Name] = col.CustomType
		}
		break
	}

	return columnTypes
}

// LoadColumns loads schema table/view columns.
func (tl *TypeLoader) LoadColumns(typeTpl *internal.Type) error {
	var err error

	// load columns
	columnList, err := tl.source.ColumnList(typeTpl.Table.TableName)
	if err != nil {
		return err
	}

	columnTypes := tl.tableCustomTypes(typeTpl.Table.TableName)
	// process columns
	for _, c := range columnList {
		ignore := false

		for _, ignoreField := range tl.ignoreFields {
			if ignoreField == c.ColumnName {
				// Skip adding this field if user has specified they are not
				// interested.
				//
				// This could be useful for fields which are managed by the
				// database (e.g. automatically updated timestamps) instead of
				// via Go code.
				ignore = true
			}
		}

		if ignore {
			continue
		}

		len, nilType, typ := parseSpannerType(c.DataType, !c.NotNull)

		// set col info
		f := &internal.Field{
			Name:    internal.SnakeToCamel(c.ColumnName),
			Col:     c,
			Len:     len,
			NilType: nilType,
			Type:    typ,
		}

		// set custom type
		customType, ok := columnTypes[c.ColumnName]
		if ok && tl.validateCustomType(c.DataType, customType) {
			f.CustomType = customType
		}

		// append col to template fields
		typeTpl.Fields = append(typeTpl.Fields, f)
	}

	return nil
}

// LoadIndexes loads schema index definitions.
func (tl *TypeLoader) LoadIndexes(tableMap map[string]*internal.Type) (map[string]*internal.Index, error) {
	var err error

	ixMap := map[string]*internal.Index{}
	for _, t := range tableMap {
		// load table indexes
		err = tl.LoadTableIndexes(t, ixMap)
		if err != nil {
			return nil, err
		}
	}

	return ixMap, nil
}

// LoadTableIndexes loads schema index definitions per table.
func (tl *TypeLoader) LoadTableIndexes(typeTpl *internal.Type, ixMap map[string]*internal.Index) error {
	var err error
	var priIxLoaded bool

	// load indexes
	indexList, err := tl.source.IndexList(typeTpl.Table.TableName)
	if err != nil {
		return err
	}

	// process indexes
	for _, ix := range indexList {
		// save whether or not the primary key index was processed
		priIxLoaded = priIxLoaded || ix.IsPrimary

		// create index template
		ixTpl := &internal.Index{
			Name:   internal.SnakeToCamel(ix.IndexName),
			Schema: "",
			Type:   typeTpl,
			Fields: []*internal.Field{},
			Index:  ix,
		}

		// load index columns
		err = tl.LoadIndexColumns(ixTpl)
		if err != nil {
			return err
		}

		// build func name
		ixTpl.FuncName = tl.buildIndexFuncName(ixTpl)
		ixTpl.LegacyFuncName = tl.buildLegacyIndexFuncName(ixTpl)

		ixMap[typeTpl.Table.TableName+"_"+ix.IndexName] = ixTpl
	}

	return nil
}

func (tl *TypeLoader) buildLegacyIndexFuncName(ixTpl *internal.Index) string {
	// build func name
	funcName := ixTpl.Type.Name
	if !ixTpl.Index.IsUnique {
		funcName = tl.inflector.Pluralize(ixTpl.Type.Name)
	}
	funcName = funcName + "By"

	// add param names
	paramNames := make([]string, 0, len(ixTpl.Fields))
	for _, f := range ixTpl.StoringFields {
		paramNames = append(paramNames, f.Name)
	}
	for _, f := range ixTpl.Fields {
		paramNames = append(paramNames, f.Name)
	}

	return funcName + strings.Join(paramNames, "")
}

func (tl *TypeLoader) buildIndexFuncName(ixTpl *internal.Index) string {
	// build func name
	funcName := ixTpl.Type.Name
	if !ixTpl.Index.IsUnique {
		funcName = tl.inflector.Pluralize(ixTpl.Type.Name)
	}
	return funcName + "By" + internal.SnakeToCamel(ixTpl.Index.IndexName)
}

// LoadIndexColumns loads the index column information.
func (tl *TypeLoader) LoadIndexColumns(ixTpl *internal.Index) error {
	var err error

	// load index columns
	indexCols, err := tl.source.IndexColumnList(ixTpl.Type.Table.TableName, ixTpl.Index.IndexName)
	if err != nil {
		return err
	}

	// process index columns
	for _, ic := range indexCols {
		var field *internal.Field

	fieldLoop:
		// find field
		for _, f := range ixTpl.Type.Fields {
			if f.Col.ColumnName == ic.ColumnName {
				field = f
				break fieldLoop
			}
		}

		if field == nil {
			continue
		}

		if ic.Storing {
			// Storing column is added to StoringFields
			ixTpl.StoringFields = append(ixTpl.StoringFields, field)
		} else {
			ixTpl.Fields = append(ixTpl.Fields, field)
		}
		if !field.Col.NotNull {
			ixTpl.NullableFields = append(ixTpl.NullableFields, field)
		}
	}

	return nil
}

func setIndexesToTables(tableMap map[string]*internal.Type, ixMap map[string]*internal.Index) {
	indexes := make([]*internal.Index, 0, len(ixMap))
	for _, ix := range ixMap {
		indexes = append(indexes, ix)
	}
	// sort by index name
	sort.Slice(indexes, func(i, j int) bool {
		return indexes[i].Index.IndexName < indexes[j].Index.IndexName
	})
	for tbl, t := range tableMap {
		for _, ix := range indexes {
			if ix.Type.Table.TableName == tbl {
				t.Indexes = append(t.Indexes, ix)
			}
		}
	}
}
