{{- $short := (shortName .Name "err" "res" "sqlstr" "db" "YOLog") -}}
{{- $table := (.TableName) -}}

// {{ .Name }} represents a row from '{{ $table }}'.
type {{ .Name }} struct {
{{- range .Fields }}
{{- if eq (.SpannerDataType) (.ColumnName) }}
	{{ .Name }} string `spanner:"{{ .ColumnName }}" json:"{{ .ColumnName }}"` // {{ .ColumnName }} enum
{{- else }}
	{{ .Name }} {{ .Type }} `spanner:"{{ .ColumnName }}" json:"{{ .ColumnName }}"` // {{ .ColumnName }}
{{- end }}
{{- end }}
}

func {{ .Name }}PrimaryKeys() []string {
     return []string{
{{- range .PrimaryKeyFields }}
		"{{ .ColumnName }}",
{{- end }}
	}
}

func {{ .Name }}Columns() []string {
	return []string{
{{- range .Fields }}
		"{{ .ColumnName }}",
{{- end }}
	}
}

func ({{ $short }} *{{ .Name }}) columnsToPtrs(cols []string) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		switch col {
{{- range .Fields }}
		case "{{ .ColumnName }}":
			ret = append(ret, yoDecode(&{{ $short }}.{{ .Name }}))
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
		case "{{ .ColumnName }}":
			ret = append(ret, yoEncode({{ $short }}.{{ .Name }}))
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
	return func(row *spanner.Row) (*{{ .Name }}, error) {
        var {{ $short }} {{ .Name }}
        ptrs, err := {{ $short }}.columnsToPtrs(cols)
        if err != nil {
            return nil, err
        }

        if err := row.Columns(ptrs...); err != nil {
            return nil, err
        }

		return &{{ $short }}, nil
	}
}
