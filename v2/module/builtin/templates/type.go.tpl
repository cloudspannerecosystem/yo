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

func {{ .Name }}PrimaryKeys() []string {
     return []string{
{{- range .PrimaryKeyFields }}
		"{{ colname .Col }}",
{{- end }}
	}
}

func {{ .Name }}Columns() []string {
	return []string{
{{- range .Fields }}
		"{{ colname .Col }}",
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
