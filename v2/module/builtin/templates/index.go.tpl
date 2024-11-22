{{- range .Indexes }}
{{- $short := (shortName .Type.Name "err" "sqlstr" "db" "q" "res" "YOLog" .Fields) -}}
{{- $table := (.Type.TableName) -}}

{{- if not .IsUnique }}
// Find{{ .FuncName }} retrieves multiple rows from '{{ $table }}' as a slice of {{ .Type.Name }}.
//
// Generated from index '{{ .IndexName }}'.
func Find{{ .FuncName }}(ctx context.Context, db YODB{{ goParams .Fields true true }}) ([]*{{ .Type.Name }}, error) {
{{- else }}
// Find{{ .FuncName }} retrieves a row from '{{ $table }}' as a {{ .Type.Name }}.
//
// If no row is present with the given key, then ReadRow returns an error where
// spanner.ErrCode(err) is codes.NotFound.
//
// Generated from unique index '{{ .IndexName }}'.
func Find{{ .FuncName }}(ctx context.Context, db YODB{{ goParams .Fields true true }}) (*{{ .Type.Name }}, error) {
{{- end }}
	{{- if not .NullableFields }}
	const sqlstr = "SELECT " +
		"{{ columnNamesWithoutHidden .Type.Fields }} " +
		"FROM {{ $table }}@{FORCE_INDEX={{ .IndexName }}} " +
		"WHERE {{ columnNamesQuery .Fields " AND " }}"
	{{- else }}
	var sqlstr = "SELECT " +
		"{{ columnNamesWithoutHidden .Type.Fields }} " +
		"FROM {{ $table }}@{FORCE_INDEX={{ .IndexName }}} "

	conds := make([]string, {{ len .Fields }})
	{{- range $i, $f := .Fields }}
	{{- if $f.IsNotNull }}
		conds[{{ $i }}] = "{{ escape $f.ColumnName }} = @param{{ $i }}"
	{{- else }}
	if {{ nullcheck $f }} {
		conds[{{ $i }}] = "{{ escape $f.ColumnName }} IS NULL"
	} else {
		conds[{{ $i }}] = "{{ escape $f.ColumnName }} = @param{{ $i }}"
	}
	{{- end }}
	{{- end }}
	sqlstr += "WHERE " + strings.Join(conds, " AND ")
	{{- end }}

	stmt := spanner.NewStatement(sqlstr)
	{{- range $i, $f := .Fields }}
		stmt.Params["param{{ $i }}"] = {{ goEncodedParam $f.Name }}
	{{- end}}


	decoder := new{{ .Type.Name }}_Decoder({{ .Type.Name }}Columns())

	// run query
	YOLog(ctx, sqlstr{{ goParams .Fields true false }})
{{- if .IsUnique }}
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, newErrorWithCode(codes.NotFound, "Find{{ .FuncName }}", "{{ $table }}", err)
		}
		return nil, newError("Find{{ .FuncName }}", "{{ $table }}", err)
	}

	{{ $short }}, err := decoder(row)
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "Find{{ .FuncName }}", "{{ $table }}", err)
	}

	return {{ $short }}, nil
{{- else }}
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	// load results
	res := []*{{ .Type.Name }}{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, newError("Find{{ .FuncName }}", "{{ $table }}", err)
		}

		{{ $short }}, err := decoder(row)
        if err != nil {
            return nil, newErrorWithCode(codes.Internal, "Find{{ .FuncName }}", "{{ $table }}", err)
        }

		res = append(res, {{ $short }})
	}

	return res, nil
{{- end }}
}


// Read{{ .FuncName }} retrieves multiples rows from '{{ $table }}' by KeySet as a slice.
//
// This does not retrieve all columns of '{{ $table }}' because an index has only columns
// used for primary key, index key and storing columns. If you need more columns, add storing
// columns or Read by primary key or Query with join.
//
// Generated from index '{{ .IndexName }}'.
func Read{{ .FuncName }}(ctx context.Context, db YODB, keys spanner.KeySet) ([]*{{ .Type.Name }}, error) {
	var res []*{{ .Type.Name }}
    columns := []string{
{{- range .Type.PrimaryKeyFields }}
		"{{ .ColumnName }}",
{{- end }}
{{- range .Fields }}
		"{{ .ColumnName }}",
{{- end }}
{{- range .StoringFields }}
		"{{ .ColumnName }}",
{{- end }}
}

	decoder := new{{ .Type.Name }}_Decoder(columns)

	rows := db.ReadUsingIndex(ctx, "{{ $table }}", "{{ .IndexName }}", keys, columns)
	err := rows.Do(func(row *spanner.Row) error {
		{{ $short }}, err := decoder(row)
		if err != nil {
			return err
		}
		res = append(res, {{ $short }})

		return nil
	})
	if err != nil {
		return nil, newErrorWithCode(codes.Internal, "Read{{ .FuncName }}", "{{ $table }}", err)
	}

    return res, nil
}
{{- end }}
