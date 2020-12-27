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

package generator

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/kenshaw/snaker"

	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/models"
)

// newTemplateFuncs returns a set of template funcs bound to the supplied args.
func (a *Generator) newTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"colcount":          a.colcount,
		"columncount":       a.columncount,
		"colnames":          a.colnames,
		"escapedcolnames":   a.escapedcolnames,
		"colnamesquery":     a.colnamesquery,
		"colprefixnames":    a.colprefixnames,
		"colvals":           a.colvals,
		"fieldnames":        a.fieldnames,
		"goparamlist":       a.goparamlist,
		"goEncodedParams":   a.goEncodedParams,
		"gocustomparamlist": a.gocustomparamlist,
		"reniltype":         a.reniltype,
		"retype":            a.retype,
		"shortname":         a.shortname,
		"goconvert":         a.goconvert,
		"escapedcolname":    a.escapedcolname,
		"hascolumn":         a.hascolumn,
		"hasfield":          a.hasfield,
		"getstartcount":     a.getstartcount,
		"customfieldcount":  a.customfieldcount,
		"goparamname":       a.goparamname,
		"customtypeparam":   a.customtypeparam,
		"tolower":           a.tolower,
		"nullcheck":         a.nullcheck,
		"pluralize":         a.pluralize,
	}
}

func ignoreFromMultiTypes(ignoreNames []interface{}) map[string]bool {
	ignore := map[string]bool{}
	for _, f := range ignoreNames {
		switch k := f.(type) {
		case string:
			ignore[k] = true

		case []*models.Field:
			for _, f := range k {
				ignore[f.Name] = true
			}
		}
	}

	return ignore
}

// retype checks typ against known types, and prefixing
// customTypePackage (if applicable).
func (a *Generator) retype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	prefix := ""
	for strings.HasPrefix(typ, "[]") {
		typ = typ[2:]
		prefix = prefix + "[]"
	}

	if _, ok := KnownTypeMap[typ]; !ok {
		pkg := a.customTypePackage
		if pkg != "" {
			pkg = pkg + "."
		}

		return prefix + pkg + typ
	}

	return prefix + typ
}

// reniltype checks typ against known nil types (similar to retype), prefixing
// customTypePackage (if applicable).
func (a *Generator) reniltype(typ string) string {
	if strings.Contains(typ, ".") {
		return typ
	}

	if strings.HasSuffix(typ, "{}") {
		if _, ok := KnownTypeMap[typ[:len(typ)-2]]; ok {
			return typ
		}

		pkg := a.customTypePackage
		if pkg != "" {
			pkg = pkg + "."
		}

		return pkg + typ
	}

	return typ
}

// shortname generates a safe Go identifier for typ. typ is first checked
// against ShortNameTypeMap, and if not found, then the value is
// calculated and stored in the ShortNameTypeMap for future use.
//
// A shortname is the concatentation of the lowercase of the first character in
// the words comprising the name. For example, "MyCustomName" will have have
// the shortname of "mcn".
//
// If a generated shortname conflicts with a Go reserved name, then the
// corresponding value in goReservedNames map will be used.
//
// Generated shortnames that have conflicts with any scopeConflicts member will
// have nameConflictSuffix appended.
//
// Note: recognized types for scopeConflicts are string, []*models.Field.
func (a *Generator) shortname(typ string, scopeConflicts ...interface{}) string {
	var v string
	var ok bool

	// check short name map
	if v, ok = ShortNameTypeMap[typ]; !ok {
		// calc the short name
		u := []string{}
		for _, s := range strings.Split(strings.ToLower(snaker.CamelToSnake(typ)), "_") {
			if len(s) > 0 && s != "id" {
				u = append(u, s[:1])
			}
		}
		v = strings.Join(u, "")

		// check go reserved names
		if n, ok := goReservedNames[v]; ok {
			v = n
		}

		// store back to short name map
		ShortNameTypeMap[typ] = v
	}

	// add scopeConflicts to conflicts
	for _, c := range scopeConflicts {
		switch k := c.(type) {
		case string:
			if k == v {
				v = v + a.nameConflictSuffix
			}

		case []*models.Field:
			for _, f := range k {
				if f.Name == v {
					v = v + a.nameConflictSuffix
				}
			}

		default:
			panic("shortname: supported type")
		}
	}

	// append suffix if conflict exists
	if _, ok := ConflictedShortNames[v]; ok {
		v = v + a.nameConflictSuffix
	}

	return v
}

// colnames creates a list of the column names found in fields, excluding any
// models.Field with Name contained in ignoreNames.
//
// Used to present a comma separated list of column names, that can be used in
// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
// (ie, "field_1, field_2, field_3, ...").
func (a *Generator) colnames(fields []*models.Field, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + f.ColumnName
		i++
	}
	return str
}

// escapedcolnames creates a list of the column names found in the field, excluding any
// models.Field with Name contained in ignoreNames.
// If use reserved word as a column name, it is escaped by surrounding it with backquotes.
//
// Used to present a comma separated list of espcaped column names, that can be used in
// a SELECT, or UPDATE, or other SQL clause requiring an list of identifiers
// (ie, "field_1, field_2, field_3, ...").
func (a *Generator) escapedcolnames(fields []*models.Field, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + internal.EscapeColumnName(f.ColumnName)
		i++
	}
	return str
}

// colnamesquery creates a list of the column names in fields as a query and
// joined by sep, excluding any models.Field with Name contained in ignoreNames.
//
// Used to create a list of column names in a WHERE clause (ie, "field_1 = $1
// AND field_2 = $2 AND ...") or in an UPDATE clause (ie, "field = $1, field =
// $2, ...").
func (a *Generator) colnamesquery(fields []*models.Field, sep string, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + sep
		}
		str = str + internal.EscapeColumnName(f.ColumnName) + " = " + a.loader.NthParam(i)
		i++
	}

	return str
}

// colprefixnames creates a list of the column names found in fields with the
// supplied prefix, excluding any models.Field with Name contained in ignoreNames.
//
// Used to present a comma separated list of column names with a prefix. Used in
// a SELECT, or UPDATE (ie, "t.field_1, t.field_2, t.field_3, ...").
func (a *Generator) colprefixnames(fields []*models.Field, prefix string, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + f.ColumnName
		i++
	}

	return str
}

// colvals creates a list of value place holders for fields excluding any models.Field
// with Name contained in ignoreNames.
//
// Used to present a comma separated list of column place holders, used in a
// SELECT or UPDATE statement (ie, "$1, $2, $3 ...").
func (a *Generator) colvals(fields []*models.Field, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + a.loader.NthParam(i)
		i++
	}

	return str
}

// fieldnames creates a list of field names from fields of the adding the
// provided prefix, and excluding any models.Field with Name contained in ignoreNames.
//
// Used to present a comma separated list of field names, ie in a Go statement
// (ie, "t.Field1, t.Field2, t.Field3 ...")
func (a *Generator) fieldnames(fields []*models.Field, prefix string, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}

		str = str + prefix + "." + f.Name
		i++
	}

	return str
}

// colcount returns the 1-based count of fields, excluding any models.Field with Name
// contained in ignoreNames.
//
// Used to get the count of fields, and useful for specifying the last SQL
// parameter.
func (a *Generator) colcount(fields []*models.Field, ignoreNames ...interface{}) int {
	ignore := ignoreFromMultiTypes(ignoreNames)

	i := 1
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		i++
	}
	return i
}

// columncount returns the number of fields
func (a *Generator) columncount(fields []*models.Field, ignoreNames ...interface{}) int {
	ignore := ignoreFromMultiTypes(ignoreNames)

	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		i++
	}
	return i
}

// goReservedNames is a map of of go reserved names to "safe" names.
var goReservedNames = map[string]string{
	"break":       "brk",
	"case":        "cs",
	"chan":        "chn",
	"const":       "cnst",
	"continue":    "cnt",
	"default":     "def",
	"defer":       "dfr",
	"else":        "els",
	"fallthrough": "flthrough",
	"for":         "fr",
	"func":        "fn",
	"go":          "goVal",
	"goto":        "gt",
	"if":          "ifVal",
	"import":      "imp",
	"interface":   "iface",
	"map":         "mp",
	"package":     "pkg",
	"range":       "rnge",
	"return":      "ret",
	"select":      "slct",
	"struct":      "strct",
	"switch":      "swtch",
	"type":        "typ",
	"var":         "vr",

	// go types
	"error":      "e",
	"bool":       "b",
	"string":     "str",
	"byte":       "byt",
	"rune":       "r",
	"uintptr":    "uptr",
	"int":        "i",
	"int8":       "i8",
	"int16":      "i16",
	"int32":      "i32",
	"int64":      "i64",
	"uint":       "u",
	"uint8":      "u8",
	"uint16":     "u16",
	"uint32":     "u32",
	"uint64":     "u64",
	"float32":    "z",
	"float64":    "f",
	"complex64":  "c",
	"complex128": "c128",
}

// goparamlist converts a list of fields into their named Go parameters,
// skipping any models.Field with Name contained in ignoreNames. addType will cause
// the go Type to be added after each variable name. addPrefix will cause the
// returned string to be prefixed with ", " if the generated string is not
// empty.
//
// Any field name encountered will be checked against goReservedNames, and will
// have its name substituted by its corresponding looked up value.
//
// Used to present a comma separated list of Go variable names for use with as
// either a Go func parameter list, or in a call to another Go func.
// (ie, ", a, b, c, ..." or ", a T1, b T2, c T3, ...").
func (a *Generator) goparamlist(fields []*models.Field, addPrefix bool, addType bool, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	i := 0
	vals := []string{}
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		s := "v" + strconv.Itoa(i)
		if len(f.Name) > 0 {
			s = a.goparamname(f.Name)
		}

		// add the go type
		if addType {
			s += " " + a.retype(f.Type)
		}

		// add to vals
		vals = append(vals, s)

		i++
	}

	// concat generated values
	str := strings.Join(vals, ", ")
	if addPrefix && str != "" {
		return ", " + str
	}

	return str
}

func (a *Generator) goEncodedParams(fields []*models.Field, addPrefix bool, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	i := 0
	vals := []string{}
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		s := "v" + strconv.Itoa(i)
		if len(f.Name) > 0 {
			s = a.goparamname(f.Name)
		}

		// add to vals
		vals = append(vals, fmt.Sprintf("yoEncode(%s)", s))

		i++
	}

	// concat generated values
	str := strings.Join(vals, ", ")
	if addPrefix && str != "" {
		return ", " + str
	}

	return str
}

// gocustomparamlist is a similar function like goparamlist
// which converts a list of fields into their named Go parameters
// and also do custom type conversion
func (a *Generator) gocustomparamlist(fields []*models.Field, addPrefix bool, addType bool, ignoreNames ...interface{}) string {
	ignore := ignoreFromMultiTypes(ignoreNames)

	i := 0
	vals := []string{}
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		s := "v" + strconv.Itoa(i)
		if len(f.Name) > 0 {
			s = a.goparamname(f.Name)
		}

		// add the go type
		if addType {
			s += " " + a.retype(f.Type)
		}
		// add to vals
		vals = append(vals, s)

		i++
	}

	// concat generated values
	str := strings.Join(vals, ", ")
	if addPrefix && str != "" {
		return ", " + str
	}

	return str
}

// goconvert generates the Go conversion for f in order for it
// to be assignable to t
func (a *Generator) goconvert(prefix string, f *models.Field, t *models.Field) string {
	expr := prefix + "." + f.Name
	if f.Type == t.Type {
		return expr
	}

	ft := f.Type
	if strings.HasPrefix(ft, "sql.Null") {
		expr = expr + "." + f.Type[8:]
		ft = strings.ToLower(f.Type[8:])
	}

	if t.Type != ft {
		expr = t.Type + "(" + expr + ")"
	}

	return expr
}

// escapedcolname returns the ColumnName of col. It is escaped for query.
func (a *Generator) escapedcolname(col string) string {
	return internal.EscapeColumnName(col)
}

// hascolumn takes a list of fields and determines if field with the specified
// column name is in the list.
func (a *Generator) hascolumn(fields []*models.Field, name string) bool {
	for _, f := range fields {
		if f.ColumnName == name {
			return true
		}
	}

	return false
}

// hasfield takes a list of fields and determines if field with the specified
// field name is in the list.
func (a *Generator) hasfield(fields []*models.Field, name string) bool {
	for _, f := range fields {
		if f.Name == name {
			return true
		}
	}

	return false
}

// getstartcount returns a starting count for numbering columsn in queries
func (a *Generator) getstartcount(fields []*models.Field, pkFields []*models.Field) int {
	return len(fields) - len(pkFields)
}

// customfieldcount returns the number of field has custom type
func (a *Generator) customfieldcount(fields []*models.Field) int {
	return 0
}

// customtypeparam returns temporary custom field variable name with prefix c
func (a *Generator) customtypeparam(name string) string {
	return "c" + name
}

// goparamname make the first word of name to lowercase
func (a *Generator) goparamname(name string) string {
	ns := strings.Split(snaker.CamelToSnake(name), "_")
	name = strings.ToLower(ns[0]) + name[len(ns[0]):]

	// check go reserved names
	if r, ok := goReservedNames[strings.ToLower(name)]; ok {
		name = r
	}

	return name
}

// tolower converts s to lower case.
func (a *Generator) tolower(s string) string {
	return strings.ToLower(s)
}

// nullcheck generates a code to check the field value is null.
func (a *Generator) nullcheck(field *models.Field) string {
	paramName := a.goparamname(field.Name)

	switch field.Type {
	case "spanner.NullInt64",
		"spanner.NullString",
		"spanner.NullFloat64",
		"spanner.NullBool",
		"spanner.NullTime",
		"spanner.NullDate":
		return fmt.Sprintf("%s.IsNull()", paramName)
	}

	return fmt.Sprintf("yo, ok := %s.(yoIsNull); ok && yo.IsNull()", paramName)
}

// pluralize converts s to plural.
func (a *Generator) pluralize(s string) string {
	return a.inflector.Pluralize(s)
}
