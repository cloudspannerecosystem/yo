{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "YOLog") -}}
{{- $table := (.Table.TableName) -}}
// {{ .Name }} represents a row from '{{ $table }}'.
type {{ .Name }} struct {
{{- range .Fields }}
{{- if eq (.Col.DataType) (.Col.ColumnName) }}
	{{ .Name }} string `spanner:"{{ .Col.ColumnName }}" json:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }} enum
{{- else if .CustomType }}
	{{ .Name }} {{ retype .CustomType }} `spanner:"{{ .Col.ColumnName }}" json:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- else }}
	{{ .Name }} {{ .Type }} `spanner:"{{ .Col.ColumnName }}" json:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- end }}
{{- end }}
}

{{ if .PrimaryKey }}
func {{ .Name }}PrimaryKeys() []string {
     return []string{
{{- range .PrimaryKeyFields }}
		"{{ colname .Col }}",
{{- end }}
	}
}
{{- end }}

func {{ .Name }}Columns() []string {
	return []string{
{{- range .Fields }}
		"{{ colname .Col }}",
{{- end }}
	}
}

func {{ .Name }}WritableColumns() []string {
	return []string{
{{- range .Fields }}
	{{- if not .Col.IsGenerated }}
		"{{ colname .Col }}",
	{{- end }}
{{- end }}
	}
}

func ({{ $short }} *{{ .Name }}) columnsToPtrs(cols []string, customPtrs map[string]interface{}) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		if val, ok := customPtrs[col]; ok {
			ret = append(ret, val)
			continue
		}

		switch col {
{{- range .Fields }}
		case "{{ colname .Col }}":
			ret = append(ret, &{{ $short }}.{{ .Name }})
{{- end }}
		default:
			return nil, fmt.Errorf("unknown column: %s", col)
		}
	}
	return ret, nil
}

func ({{ $short }} *{{ .Name }}) columnsToValues(cols []string) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		switch col {
{{- range .Fields }}
		case "{{ colname .Col }}":
			{{- if .CustomType }}
			ret = append(ret, {{ .Type }}({{ $short }}.{{ .Name }}))
			{{- else }}
			ret = append(ret, {{ $short }}.{{ .Name }})
			{{- end }}
{{- end }}
		default:
			return nil, fmt.Errorf("unknown column: %s", col)
		}
	}

	return ret, nil
}

// new{{ .Name }}_Decoder returns a decoder which reads a row from *spanner.Row
// into {{ .Name }}. The decoder is not goroutine-safe. Don't use it concurrently.
func new{{ .Name }}_Decoder(cols []string) func(*spanner.Row) (*{{ .Name }}, error) {
	{{- range .Fields }}
		{{- if .CustomType }}
			var {{ customtypeparam .Name }} {{ .Type }}
		{{- end }}
	{{- end }}
	customPtrs := map[string]interface{}{
		{{- range .Fields }}
			{{- if .CustomType }}
				"{{ colname .Col }}": &{{ customtypeparam .Name }},
			{{- end }}
	{{- end }}
	}

	return func(row *spanner.Row) (*{{ .Name }}, error) {
        var {{ $short }} {{ .Name }}
        ptrs, err := {{ $short }}.columnsToPtrs(cols, customPtrs)
        if err != nil {
            return nil, err
        }

        if err := row.Columns(ptrs...); err != nil {
            return nil, err
        }
        {{- range .Fields }}
            {{- if .CustomType }}
                {{ $short }}.{{ .Name }} = {{ retype .CustomType }}({{ customtypeparam .Name }})
            {{- end }}
        {{- end }}


		return &{{ $short }}, nil
	}
}

// Insert returns a Mutation to insert a row into a table. If the row already
// exists, the write or transaction fails.
func ({{ $short }} *{{ .Name }}) Insert(ctx context.Context) *spanner.Mutation {
	values, _ := {{ $short }}.columnsToValues({{ .Name }}WritableColumns())
	return spanner.Insert("{{ $table }}", {{ .Name }}WritableColumns(), values)
}

{{ if ne (fieldnames .Fields $short .PrimaryKeyFields) "" }}
// Update returns a Mutation to update a row in a table. If the row does not
// already exist, the write or transaction fails.
func ({{ $short }} *{{ .Name }}) Update(ctx context.Context) *spanner.Mutation {
	values, _ := {{ $short }}.columnsToValues({{ .Name }}WritableColumns())
	return spanner.Update("{{ $table }}", {{ .Name }}WritableColumns(), values)
}

// InsertOrUpdate returns a Mutation to insert a row into a table. If the row
// already exists, it updates it instead. Any column values not explicitly
// written are preserved.
func ({{ $short }} *{{ .Name }}) InsertOrUpdate(ctx context.Context) *spanner.Mutation {
	values, _ := {{ $short }}.columnsToValues({{ .Name }}WritableColumns())
	return spanner.InsertOrUpdate("{{ $table }}", {{ .Name }}WritableColumns(), values)
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
func Find{{ .Name }}(ctx context.Context, db YORODB{{ gocustomparamlist .PrimaryKeyFields true true }}) (*{{ .Name }}, error) {
	key := spanner.Key{ {{ gocustomparamlist .PrimaryKeyFields false false }} }
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
