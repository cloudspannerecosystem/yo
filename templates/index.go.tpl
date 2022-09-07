{{- $short := (shortname .Type.Name "err" "sqlstr" "db" "q" "res" "YOLog" .Fields) -}}
{{- $table := (.Type.Table.TableName) -}}
{{- if not .Index.IsUnique }}
// Find{{ .FuncName }} retrieves multiple rows from '{{ $table }}' as a slice of {{ .Type.Name }}.
//
// Generated from index '{{ .Index.IndexName }}'.
func Find{{ .FuncName }}(ctx context.Context, db YORODB{{ gocustomparamlist .Fields true true }}) ([]*{{ .Type.Name }}, error) {
{{- else }}
// Find{{ .FuncName }} retrieves a row from '{{ $table }}' as a {{ .Type.Name }}.
//
// If no row is present with the given key, then ReadRow returns an error where
// spanner.ErrCode(err) is codes.NotFound.
//
// Generated from unique index '{{ .Index.IndexName }}'.
func Find{{ .FuncName }}(ctx context.Context, db YORODB{{ gocustomparamlist .Fields true true }}) (*{{ .Type.Name }}, error) {
{{- end }}
	{{- if not .NullableFields }}
	const sqlstr = "SELECT " +
		"{{ escapedcolnames .Type.Fields }} " +
		"FROM {{ $table }}@{FORCE_INDEX={{ .Index.IndexName }}} " +
		"WHERE {{ colnamesquery .Fields " AND " }}"
	{{- else }}
	var sqlstr = "SELECT " +
		"{{ escapedcolnames .Type.Fields }} " +
		"FROM {{ $table }}@{FORCE_INDEX={{ .Index.IndexName }}} "

	conds := make([]string, {{ columncount .Fields }})
	{{- range $i, $f := .Fields }}
	{{- if $f.Col.NotNull }}
		conds[{{ $i }}] = "{{ escapedcolname $f.Col }} = @param{{ $i }}"
	{{- else }}
	if {{ nullcheck $f }} {
		conds[{{ $i }}] = "{{ escapedcolname $f.Col }} IS NULL"
	} else {
		conds[{{ $i }}] = "{{ escapedcolname $f.Col }} = @param{{ $i }}"
	}
	{{- end }}
	{{- end }}
	sqlstr += "WHERE " + strings.Join(conds, " AND ")
	{{- end }}

	stmt := spanner.NewStatement(sqlstr)
	{{- range $i, $f := .Fields }}
		{{- if $f.CustomType }}
			stmt.Params["param{{ $i }}"] = {{ $f.Type }}({{ goparamname $f.Name }})
		{{- else }}
			stmt.Params["param{{ $i }}"] = {{ goparamname $f.Name }}
		{{- end }}
	{{- end}}


	decoder := new{{ .Type.Name }}_Decoder({{ .Type.Name }}Columns())

	// run query
	YOLog(ctx, sqlstr{{ goparamlist .Fields true false }})
{{- if .Index.IsUnique }}
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
// Generated from unique index '{{ .Index.IndexName }}'.
func Read{{ .FuncName }}(ctx context.Context, db YORODB, keys spanner.KeySet) ([]*{{ .Type.Name }}, error) {
	var res []*{{ .Type.Name }}
    columns := []string{
{{- range .Type.PrimaryKeyFields }}
		"{{ colname .Col }}",
{{- end }}
{{- range .Fields }}
		"{{ colname .Col }}",
{{- end }}
{{- range .StoringFields }}
		"{{ colname .Col }}",
{{- end }}
}

	decoder := new{{ .Type.Name }}_Decoder(columns)

	rows := db.ReadUsingIndex(ctx, "{{ $table }}", "{{ .Index.IndexName }}", keys, columns)
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

