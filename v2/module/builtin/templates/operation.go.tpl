{{- $short := (shortName .Type.Name "err" "res" "sqlstr" "db" "YOLog") -}}
{{- $table := (.Type.TableName) -}}

// Insert returns a Mutation to insert a row into a table. If the row already
// exists, the write or transaction fails.
func ({{ $short }} *{{ .Type.Name }}) Insert(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}) *{{ .PackageRegistry.Use presetPackages.goSpanner "Mutation" }} {
	values, _ := {{ $short }}.columnsToValues({{ .Type.Name }}WritableColumns())
	return {{ .PackageRegistry.Use presetPackages.goSpanner "Insert" }}("{{ $table }}", {{ .Type.Name }}WritableColumns(), values)
}

{{ if ne (len .Type.Fields) (len .Type.PrimaryKeyFields) }}
// Update returns a Mutation to update a row in a table. If the row does not
// already exist, the write or transaction fails.
func ({{ $short }} *{{ .Type.Name }}) Update(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}) *{{ .PackageRegistry.Use presetPackages.goSpanner "Mutation" }} {
	values, _ := {{ $short }}.columnsToValues({{ .Type.Name }}WritableColumns())
	return {{ .PackageRegistry.Use presetPackages.goSpanner "Update" }}("{{ $table }}", {{ .Type.Name }}WritableColumns(), values)
}

// InsertOrUpdate returns a Mutation to insert a row into a table. If the row
// already exists, it updates it instead. Any column values not explicitly
// written are preserved.
func ({{ $short }} *{{ .Type.Name }}) InsertOrUpdate(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}) *{{ .PackageRegistry.Use presetPackages.goSpanner "Mutation" }} {
	values, _ := {{ $short }}.columnsToValues({{ .Type.Name }}WritableColumns())
	return {{ .PackageRegistry.Use presetPackages.goSpanner "InsertOrUpdate" }}("{{ $table }}", {{ .Type.Name }}WritableColumns(), values)
}

// Replace returns a Mutation to insert a row into a table, deleting any
// existing row. Unlike InsertOrUpdate, this means any values not explicitly
// written become NULL.
func ({{ $short }} *{{ .Type.Name }}) Replace(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}) *{{ .PackageRegistry.Use presetPackages.goSpanner "Mutation" }} {
	values, _ := {{ $short }}.columnsToValues({{ .Type.Name }}WritableColumns())
	return  {{ .PackageRegistry.Use presetPackages.goSpanner "Replace" }}("{{ $table }}", {{ .Type.Name }}WritableColumns(), values)
}

// UpdateColumns returns a Mutation to update specified columns of a row in a table.
func ({{ $short }} *{{ .Type.Name }}) UpdateColumns(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}, cols ...string) (*{{ .PackageRegistry.Use presetPackages.goSpanner "Mutation" }}, error) {
	// add primary keys to columns to update by primary keys
	colsWithPKeys := append(cols, {{ .Type.Name }}PrimaryKeys()...)

	values, err := {{ $short }}.columnsToValues(colsWithPKeys)
	if err != nil {
		return nil, newErrorWithCode({{ .PackageRegistry.Use presetPackages.gRPCCodes "InvalidArgument" }}, "{{ .Type.Name }}.UpdateColumns", "{{ $table }}", err)
	}

	return {{ .PackageRegistry.Use presetPackages.goSpanner "Update" }}("{{ $table }}", colsWithPKeys, values), nil
}

// Find{{ .Type.Name }} gets a {{ .Type.Name }} by primary key
func Find{{ .Type.Name }}(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}, db YODB{{ goParamDefs .PackageRegistry .Type.PrimaryKeyFields true }}) (*{{ .Type.Name }}, error) {
	key := {{ .PackageRegistry.Use presetPackages.goSpanner "Key" }}{ {{ goEncodedParams .Type.PrimaryKeyFields false }} }
	row, err := db.ReadRow(ctx, "{{ $table }}", key, {{ .Type.Name }}Columns())
	if err != nil {
		return nil, newError("Find{{ .Type.Name }}", "{{ $table }}", err)
	}

	decoder := new{{ .Type.Name }}_Decoder({{ .Type.Name}}Columns())
	{{ $short }}, err := decoder(row)
	if err != nil {
		return nil, newErrorWithCode({{ .PackageRegistry.Use presetPackages.gRPCCodes "Internal" }}, "Find{{ .Type.Name }}", "{{ $table }}", err)
	}

	return {{ $short }}, nil
}

// Read{{ .Type.Name }} retrieves multiples rows from {{ .Type.Name }} by KeySet as a slice.
func Read{{ .Type.Name }}(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}, db YODB, keys {{ .PackageRegistry.Use presetPackages.goSpanner "KeySet" }}) ([]*{{ .Type.Name }}, error) {
	var res []*{{ .Type.Name }}

	decoder := new{{ .Type.Name }}_Decoder({{ .Type.Name}}Columns())

	rows := db.Read(ctx, "{{ $table }}", keys, {{ .Type.Name }}Columns())
	err := rows.Do(func(row *{{ .PackageRegistry.Use presetPackages.goSpanner "Row" }}) error {
		{{ $short }}, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, {{ $short }})

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode({{ .PackageRegistry.Use presetPackages.gRPCCodes "Internal" }}, "Read{{ .Type.Name }}", "{{ $table }}", err)
	}

	return res, nil
}
{{ end }}

// Delete deletes the {{ .Type.Name }} from the database.
func ({{ $short }} *{{ .Type.Name }}) Delete(ctx {{ .PackageRegistry.Use presetPackages.context "Context" }}) *{{ .PackageRegistry.Use presetPackages.goSpanner "Mutation" }} {
	values, _ := {{ $short }}.columnsToValues({{ .Type.Name }}PrimaryKeys())
	return {{ .PackageRegistry.Use presetPackages.goSpanner "Delete" }}("{{ $table }}", {{ .PackageRegistry.Use presetPackages.goSpanner "Key" }}(values))
}
