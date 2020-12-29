{{- $short := (shortName .Name "err" "res" "sqlstr" "db" "YOLog") -}}
{{- $table := (.TableName) -}}

// Insert returns a Mutation to insert a row into a table. If the row already
// exists, the write or transaction fails.
func ({{ $short }} *{{ .Name }}) Insert(ctx context.Context) *spanner.Mutation {
	values, _ := {{ $short }}.columnsToValues({{ .Name }}Columns())
	return spanner.Insert("{{ $table }}", {{ .Name }}Columns(), values)
}

{{ if ne (len .Fields) (len .PrimaryKeyFields) }}
// Update returns a Mutation to update a row in a table. If the row does not
// already exist, the write or transaction fails.
func ({{ $short }} *{{ .Name }}) Update(ctx context.Context) *spanner.Mutation {
	values, _ := {{ $short }}.columnsToValues({{ .Name }}Columns())
	return spanner.Update("{{ $table }}", {{ .Name }}Columns(), values)
}

// InsertOrUpdate returns a Mutation to insert a row into a table. If the row
// already exists, it updates it instead. Any column values not explicitly
// written are preserved.
func ({{ $short }} *{{ .Name }}) InsertOrUpdate(ctx context.Context) *spanner.Mutation {
	values, _ := {{ $short }}.columnsToValues({{ .Name }}Columns())
	return spanner.InsertOrUpdate("{{ $table }}", {{ .Name }}Columns(), values)
}

// UpdateColumns returns a Mutation to update specified columns of a row in a table.
func ({{ $short }} *{{ .Name }}) UpdateColumns(ctx context.Context, cols ...string) (*spanner.Mutation, error) {
	// add primary keys to columns to update by primary keys
	colsWithPKeys := append(cols, {{ .Name }}PrimaryKeys()...)

	values, err := {{ $short }}.columnsToValues(colsWithPKeys)
	if err != nil {
		return nil, newErrorWithCode(codes.InvalidArgument, "{{ .Name }}.UpdateColumns", "{{ $table }}", err)
	}

	return spanner.Update("{{ $table }}", colsWithPKeys, values), nil
}

// Find{{ .Name }} gets a {{ .Name }} by primary key
func Find{{ .Name }}(ctx context.Context, db YORODB{{ goParams .PrimaryKeyFields true true }}) (*{{ .Name }}, error) {
	key := spanner.Key{ {{ goEncodedParams .PrimaryKeyFields false }} }
	row, err := db.ReadRow(ctx, "{{ $table }}", key, {{ .Name }}Columns())
	if err != nil {
		return nil, newError("Find{{ .Name }}", "{{ $table }}", err)
	}

	decoder := new{{ .Name }}_Decoder({{ .Name}}Columns())
	{{ $short }}, err := decoder(row)
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "Find{{ .Name }}", "{{ $table }}", err)
	}

	return {{ $short }}, nil
}

// Read{{ .Name }} retrieves multiples rows from {{ .Name }} by KeySet as a slice.
func Read{{ .Name }}(ctx context.Context, db YORODB, keys spanner.KeySet) ([]*{{ .Name }}, error) {
	var res []*{{ .Name }}

	decoder := new{{ .Name }}_Decoder({{ .Name}}Columns())

	rows := db.Read(ctx, "{{ $table }}", keys, {{ .Name }}Columns())
	err := rows.Do(func(row *spanner.Row) error {
		{{ $short }}, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, {{ $short }})

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "Read{{ .Name }}", "{{ $table }}", err)
	}

	return res, nil
}
{{ end }}

// Delete deletes the {{ .Name }} from the database.
func ({{ $short }} *{{ .Name }}) Delete(ctx context.Context) *spanner.Mutation {
	values, _ := {{ $short }}.columnsToValues({{ .Name }}PrimaryKeys())
	return spanner.Delete("{{ $table }}", spanner.Key(values))
}
