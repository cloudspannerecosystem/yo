# yo

`yo` is a command-line tool to generate Go code for [Google Cloud Spanner](https://cloud.google.com/spanner/),
forked from [xo](https://github.com/xo/xo) :rose:.

`yo` uses database schema to generate code by using [Information Schema](https://cloud.google.com/spanner/docs/information-schema). `yo` runs SQL queries against tables in `INFORMATION_SCHEMA` to fetch metadata for a database and applies the metadata to Go templates to generate code/models to access Cloud Spanner.

Please feel free to report issues and send pull requests, but note that this
application is not officially supported as part of the Cloud Spanner product.

## Installation

```sh
$ go get -u go.mercari.io/yo/v2
```

## Quickstart

The following is a quick overview of using `yo` on the command line:

```sh
# change to the project directory
$ cd $GOPATH/src/path/to/project

# make an output directory
$ mkdir -p models

# generate code for a schema
$ yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models
```

## Commands

### `generate`

The `generate` command generates the Go code from DDL.

#### Examples

```sh
# Generate models from DDL under the models directory
yo generate schema.sql --from-ddl -o models

# Generate models from DDL under the models directory with custom types
yo generate schema.sql --from-ddl -o models --custom-types-file custom_column_types.yml

# Generate models under the models directory
yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models

# Generate models under the models directory with custom types
yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models --custom-types-file custom_column_types.yml
```

#### Flags

```
-c, --config string               path to Yo config file
    --disable-default-modules     disable the default modules for code generation
    --disable-format              disable to apply gofmt to generated files
    --from-ddl                    toggle using DDL file
    --global-module stringArray   add a user defined module to global modules
    --header-module string        replace the default header module by user defined module
-h, --help                        help for generate
    --ignore-fields stringArray   fields to exclude from the generated Go code types
    --ignore-tables stringArray   tables to exclude from the generated Go code types
-o, --out string                  output path or file name
-p, --package string              package name used in generated Go code
    --suffix string               output file suffix (default ".yo.go")
    --tags string                 build tags to add to a package header
    --type-module stringArray     add a user defined module to type modules
    --use-legacy-index-module     use legacy index func name
```

### `create-template`

The `create-template` command generates default template files.

#### Examples

```sh
# Create default templates under the templates directory
yo create-template --template-path templates
```

#### Flags

```
-h, --help                   help for create-template
    --template-path string   destination template path
```

### `completion`

The `completion` generates the autocompletion script for yo for the specified shell.
See each sub-command's help for details on how to use the generated script.

#### Available sub-commands

```
bash        Generate the autocompletion script for bash
fish        Generate the autocompletion script for fish
powershell  Generate the autocompletion script for powershell
zsh         Generate the autocompletion script for zsh
```

## Generated code

`yo` generates a file per table by default. Each file has a struct, metadata, and methods for a table.

### struct

From this table definition:

```
CREATE TABLE Examples (
  PKey STRING(32) NOT NULL,
  Num INT64 NOT NULL,
  CreatedAt TIMESTAMP NOT NULL,
) PRIMARY KEY(PKey);
```

This struct is generated:

```golang
type Example struct {
	PKey      string    `spanner:"PKey" json:"PKey"`           // PKey
	Num       int64     `spanner:"Num" json:"Num"`             // Num
	CreatedAt time.Time `spanner:"CreatedAt" json:"CreatedAt"` // CreatedAt
}
```

### Mutation methods

An operation against a table is represented as a mutation in Cloud Spanner. `yo` generates methods to create a mutation to modify a table.

* Insert
   * A wrapper method of `spanner.Insert`, which embeds struct values implicitly to insert a new record with struct values.
* Update
   * A wrapper method of `spanner.Update`, which embeds struct values implicitly to update all columns into struct values.
* InsertOrUpdate
   * A wrapper method of `spanner.InsertOrUpdate`, which embeds struct values implicitly to insert a new record or update all columns to struct values.
* Replace
  * A wrapper method of `spanner.Replace`, which inserts a record, deleting any existing row. Unlike InsertOrUpdate, this means any values not explicitly written become NULL.
* UpdateColumns
   * A wrapper method of `spanner.Update`, which updates specified columns into struct values.

### Read functions

`yo` generates functions to read data from Cloud Spanner. The functions are generated based on index.

Naming convention of generated functions is `FindXXXByYYY`. The XXX is table name and YYY is index name. XXX will be singular if the index is unique index, or plural if the index is not unique.


**TODO**

* Generated functions use `Query` only even if it is secondary index. Need a function to use `Read`.

### Error handling

`yo` wraps all errors as internal `yoError`. It has some methods for error handling.

* `GRPCStatus()`
   * This returns gRPC's `*status.Status`. It is intended by used from status google.golang.org/grpc/status package.
* `DBTableName()`
   * A table name where the error happens.
* `NotFound()`
   * A helper to check the error is NotFound.

The `yoError` inherits an original error from [google-cloud-go](https://github.com/GoogleCloudPlatform/google-cloud-go). It stil can be used with `status.FromError` or `status.Code` to check status code of the error. So the typical error handling will be like:

```golang
result, err := SomeFunction()
if err != nil {
	code := status.Code(err)
	if code == codes.InvalidArgument {
		// error handling for invalid argument
	}
	...
	panic("unexpected")
}
```

## Modules

`yo` uses Go [template](https://golang.org/pkg/text/template/) package to generate code.  You can customize the templates using modules. There are three types of modules in `yo`.

### Global module
The global module is a component shared among various elements. `yo_db.yo.go` is the default global module. You may add your global module by specifying `--global-module` flag to the generate command.

### Header module
The header module defines the header template for each generated code.
See [the builtin default header template](https://github.com/cloudspannerecosystem/yo/blob/021c6c2f0f72be6004656898eb74bbf92a8e216f/v2/module/builtin/templates/header.go.tpl), or you may replace it by specifying `--header-module` flag to the generate command.

### Type module
The type module is a template for each Spanner table. You can add your type module by using `--type-module` flag to the generate command.

## Templates

### Template files

You can create template files by running the `create-template` command. See the section above for more details.

| Template File         | Type   | Description                                            |
|-----------------------|--------|--------------------------------------------------------|
| `header.go.tpl`       | Header | Header template used for all the generated code        |
| `yo_db.go.tpl`        | Global | Template for components shared by different components |
| `type.go.tpl`         | Type   | Template for schema tables                             |
| `operation.go.tpl`    | Type   | Template for CRUD operations                           |
| `index.go.tpl`        | Type   | Template for schema indexes                            |
| `legacy_index.go.tpl` | Type   | Legacy template for schema indexes                     |

### Template functions

`yo` provides helper functions for templates. Those functions are listed below.

#### [filterFields(fields []*models.Field, ignoreNames ...interface{}) []*models.Field](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L77-L89)

`filterFields` filters out fields from the given list of fields. `ignoreNames` can be either a name of the field or a list of `models.Field` pointers.

##### Arguments

- `fields` - A list of `models.Field` pointers to filter through.
- `ignoreNames` - A list of names to ignore. Each element can be either a string or a list of `models.Field` pointers.

##### Examples

Ignore a field whose name is "field2".

```gotemplate
{{/* .Fields = []*models.Field{{Name: "field1"}, {Name: "field2"}, {Name: "field3"}} */}}

{{- $fields := (filterFields .Fields "field2") -}}

{{/* returns []*models.Field{{Name: "field1"}, {Name: "field3"}} */}}
```

#### [shortName(typ string, scopeConflicts ...interface{}) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L131-L196)

`shortName` generates a conflict-free Go identifier for the given `typ`. 

##### Arguments

- `typ` - A name of the Go identifier. 
- `scopeConflicts` - A list of the scope-specific reserved names. Each element can be either a string or a list of `models.Field` pointers.

##### Examples

No conflicts.

```gotemplate
{{/* .Type = *models.Field{Name: "CustomField"} */}}

{{- $short := (shortName .Type.Name "err" "db") -}}

{{/* returns "cf" */}}
```

A conflict exists.

```gotemplate
{{/* .Type = *models.Field{Name: "CustomField"} */}}

{{- $short := (shortName .Type.Name "cf" "db") -}}

{{/* returns "cfz" */}}
```

#### [nullcheck(field *models.Field) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L395-L410)

`nullcheck` generates Go code to check if the given field is null or not.

##### Arguments

- `field` - A `models.Field` pointer to generate null check Go code.

##### Examples

The field type is not Spanner Null type.

```gotemplate
{{/* .Field = *models.Field{Type: "string", Name: "field1"} */}}

{{ nullcheck .Field }}

{{/* returns "yo, ok := field1.(yoIsNull); ok && yo.IsNull()" */}}
```

The field type is Spanner Null type.

```gotemplate
{{/* .Field = *models.Field{Type: "spanner.NullInt64", Name: "field1"} */}}

{{ nullcheck .Field }}

{{/* returns "field1.IsNull()" */}}
```

#### [hasColumn(fields []*models.Field, name string) bool](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L371-L381)

`hasColumn` receives a list of fields and checks if it contains a field with the specified column name.

##### Arguments

- `fields` - A list of `models.Field` pointers to check against.
- `name` - A column name to check.

##### Examples

A field with the column name exists.

```gotemplate
{{/* .Fields = []*models.Field{{ColumnName: "field1"}, {ColumnName: "field2"}, {ColumnName: "field3"}} */}}

{{ hasColumn .Fields "field1"  }}

{{/* returns true */}}
```

A field with the column name does not exist.

```gotemplate
{{/* .Fields = []*models.Field{{ColumnName: "field1"}, {ColumnName: "field2"}, {ColumnName: "field3"}} */}}

{{ hasColumn .Fields "field0"  }}

{{/* returns false */}}
```

#### [columnNames(fields []*models.Field) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L91-L109)

`columnNames` receives a list of fields and converts it into comma-separated column names for SQL statements.

##### Arguments

- `fields` - A list of `models.Field` pointers converted from.

##### Examples

Generates an SQL column names from a list of fields.

```gotemplate
{{/* .Fields = []*models.Field{{ColumnName: "field1"}, {ColumnName: "field2"}, {ColumnName: "field3"}} */}}

var sqlstr = "SELECT {{ columnNames .Fields }} FROM Singers"

{{/* returns "field1, field2, field3" */}}
```

#### [columnNamesQuery(fields []*models.Field, sep string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L111-L129)

`columnNamesQuery` receives a list of fields and converts it into an SQL query for a WHERE clause, which is joined by `sep`.

#### Arguments

- `fields` - A list of `models.Field` pointers converted from.
- `sep` - A separator for the query.

##### Examples

Generates an SQL column names from a list of fields.

```gotemplate
{{/* .Fields = []*models.Field{{ColumnName: "field1"}, {ColumnName: "field2"}, {ColumnName: "field3"}} */}}

var sqlstr = "SELECT * FROM Singers WHERE {{ columnNamesQuery .Fields " AND " }}"

{{/* returns "field1 = @param1 AND field2 = @param2 AND field3 = @param3" */}}
```

#### [columnPrefixNames(fields []*models.Field, prefix string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L198-L215)

`columnPrefixNames` receives a list of fields and converts it into comma-separated column names with given `prefix`.

#### Arguments

- `fields` - A list of `models.Field` pointers converted from.
- `prefix` - A prefix attached to each column name.

##### Examples

```gotemplate
{{/* .Field = []*models.Field{{ColumnName: "field1"}, {ColumnName: "field2"}, {ColumnName: "field3"}} */}}

{{ columnNamesQuery .Fields "t" }}

{{/* returns "t.field1, t.field2, t.field3" */}}
```

#### [hasField(fields []*models.Field, name string) bool](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L198-L215)

`hasField` receives a list of fields and checks if it contains a filed whose `Name` is `name`.

#### Arguments

- `fields` - A list of `models.Field` pointers checked against.
- `name` - A name of the field to check.

##### Examples

The field exists.

```gotemplate
{{/* .Fields = []*models.Field{{Name: "field1"}} */}}

{{ hasField .Fields "field1" }}

{{/* returns true */}}
```

The field does not exist.

```gotemplate
{{/* .Fields = []*models.Field{{Name: "field1"}} */}}

{{ hasField .Fields "field2" }}

{{/* returns false */}}
```

#### [fieldNames(fields []*models.Field, prefix string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L217-L235)

`fieldNames` receives a list of fields and converts into comma-separated names with given `prefix`.

#### Arguments

- `fields` - A list of `models.Field` pointers converted from.
- `prefix` - A prefix attached to each name.

##### Examples

```gotemplate
{{/* .Fields = []*models.Field{{Name: "field1"}, {Name: "field2"}, {Name: "field3"}} */}}

{{ fieldNames .Fields "t" }}

{{/* returns "t.field1, t.field2, t.field3" */}}
```

#### [goParam(name string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L288-L299)

`goParam` converts the first character of the given string into lowercase. The function is supposed to be used for changing a field name to a Go parameter name.

#### Arguments

- `name` - A name of Go parameter converted from.

##### Examples

The `name` is camel case.

```gotemplate
{{ goParam "NameOfSomething" }}

{{/* returns "nameOfSomething" */}}
```

The `name` is snake case.

```gotemplate
{{ goParam "name_of_something" }}

{{/* returns "nameOfSomething" */}}
```

#### [goEncodedParam(name string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L301-L304)

`goEncodedParam` generates Go code to encode a Go parameter.

#### Arguments

- `name` - A name of Go parameter encoded from.

##### Examples

The `name` is camel case.

```gotemplate
{{ goEncodedParam "NameOfSomething" }}

{{/* returns "yoEncode(nameOfSomething)" */}}
```

The `name` is snake case.

```gotemplate
{{ goEncodedParam "name_of_something" }}

{{/* returns "yoEncode(nameOfSomething)" */}}
```

#### [goParams(fields []*models.Field, addPrefix bool, addType bool) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L306-L345)

`goParams` converts a list of fields into named Go parameters.

#### Arguments

- `fields` - A list of `models.Field` pointers converted from.
- `addPrefix` - Whether prefixing ", " or not to the final result.
- `addType` - Whether adding Go type declaration or not.

##### Examples

Without any options.

```gotemplate
{{/* .Field = []*models.Field{{Name: "field1"}, {Name: "field2"}, {Name: "field3"}} */}}

{{ goParams .Field false false }}

{{/* returns "field1, field2, field3" */}}
```

With the `addPrefix` option.

```gotemplate
{{/* .Field = []*models.Field{{Name: "field1"}, {Name: "field2"}, {Name: "field3"}} */}}

{{ goParams .Field true false }}

{{/* returns ", field1, field2, field3" */}}
```

With the `addType` option.

```gotemplate
{{/* .Field = []*models.Field{{Name: "field1", Type: "Type1"}, {Name: "field2", Type: "Type2"}, {Name: "field3", Type: "Type2"}} */}}

{{ goParams .Field false true }}

{{/* returns "field1 Type1, field2 Type2, field3 Type3" */}}
```

#### [goEncodedParams(fields []*models.Field, addPrefix bool) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L347-L369)

`goEncodedParams` generates Go code to encode Go parameters.

#### Arguments

- `fields` - A list of `models.Field` pointers encoded from.
- `addPrefix` - Whether prefixing ", " or not to the final result.

##### Examples

Without the `addPrefix` option.

```gotemplate
{{/* .Field = []*models.Field{{Name: "field1"}, {Name: "field2"}, {Name: "field3"}} */}}

{{ goEncodedParams .Field false }}

{{/* returns "yoEncode(field1), yoEncode(field2), yoEncode(field3)" */}}
```

With the `addPrefix` option.

```gotemplate
{{/* .Field = []*models.Field{{Name: "field1"}, {Name: "field2"}, {Name: "field3"}} */}}

{{ goEncodedParams .Field true }}

{{/* returns ", yoEncode(field1), yoEncode(field2), yoEncode(field3)" */}}
```

#### [escape(col string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L412-L415)

`escape` escapes a column name with back quites if it conflicts with reserved keywords. For more information about reserved keywords, please refer to [the Spanner documentation](https://cloud.google.com/spanner/docs/reference/standard-sql/lexical#reserved_keywords).

#### Arguments

- `col` - A column name.

##### Examples

No conflicts.

```gotemplate
{{ escape "Name" }}

{{/* returns "Name" */}}
```

With a conflict with the reserved keywords.

```gotemplate
{{ escape "JOIN" }}

{{/* returns "`JOIN`" */}}
```

#### [toLower(s string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L417-L420)

`toLower` converts the given string into lower case.

#### Arguments

- `s` - A string to convert to lower case.

##### Examples

```gotemplate
{{ toLower "NAME" }}

{{/* returns "name" */}}
```

#### [pluralize(s string) string](https://github.com/cloudspannerecosystem/yo/blob/64d13dc0e8aa2b0ac5eef549ebb395a0d79284c6/v2/generator/funcs.go#L422-L425)

`pluralize` pluralizes the given string.

#### Arguments

- `s` - A string to pluralize.

##### Examples

```gotemplate
{{ pluralize "name" }}

{{/* returns "names" */}}
```

## Configuration

You may customize some configurations via a config file. Use the `--config` flag to specify the config file path.

### Custom type definitions

You may define custom type rules to overwrite the original Go types in a config file.

```
tables:
  - name: "Singers"
    columns:
      - name: Id
        customType: "uint64"
  - name: "Musics"
    columns:
      - name: Id
        customType: "uint64"
      - name: MusicType
        customType: "MusicType"
```

### Custom inflection rules

`yo` uses inflection to convert singular or plural name each other. You can add inflection rules with config file.

```
inflections:
  - singular: person
    plural: people
  - singular: live
    plural: lives
```

## Changes from V1

### Changes

* Function names for index are changed to names based on the index name instead of index column names
   * The original function name based on index column names is ambiguous if there are multiple index that use the same index columns
   * The naming rule for the new function names is `Find` + _TABLE_NAME_ + _INDEX_NAME
   * Use `--use-legacy-index-module` option if you still want to use function names based on index column names
* Generated filenames become snake_case names

### Deprecations

* `--single-file` option is deprecated
* Top-level command for code generation is deprecated
   * Use `yo generate` sub command instead.
* Remove `PrimaryKey` field from `internal.Type` struct
* `--template-path` option is deprecated
   * Use module system instead (TODO)
* `--custom-types-file` and `--inflection-rule-file` options are deprecated
   * Use `--config` option instead
* `YORODB` interface is deprecated.
   * Use `YODB` instead.

### Changes in teamplate functions

* rename to lowerCamelName functions basically
* `colcount` and `columncount` are depreacated
    * use `len` instead
* `colnames`, `colnamesquery`, `colprefixname` renamed to `columnNames`, `columnNamesQuery`, `columnPrefixNames`
* `colvals` is deprecated
* `escapedcolnames` is deprecated
   * `columnNames`, `columnNamesQuery`, `columnPrefixNames` return escaped column names by default
* `escapedcolname` is deprecated
   * use `escape` instead
* `goconvert`, `retype`, `reniltype` are deprecated
   * no expected usecase
* `gocustomparamlist`, `customtypeparam` are deprecated
   * use `goEncodedParam` or `goEncodedParams` instead
* `ignoreNames` for omitting field names as variadic arguments is deprecated
   * use `filterFields` function

## Contributions

Please read the [contribution guidelines](CONTRIBUTING.md) before submitting
pull requests.

## License

Copyright 2018 Mercari, Inc.

yo is released under the [MIT License](https://opensource.org/licenses/MIT).
