{{- $short := (shortName .Type.Name "err" "res" "sqlstr" "db" "YOLog") -}}
{{- $table := (.Type.TableName) -}}

// {{ .Type.Name }} represents a row from '{{ $table }}'.
type {{ .Type.Name }} struct {
{{- range .Type.Fields }}
{{- if eq (.SpannerDataType) (.ColumnName) }}
	{{ .Name }} string `spanner:"{{ .ColumnName }}" json:"{{ .ColumnName }}"` // {{ .ColumnName }} enum
{{- else }}
	{{ .Name }} {{ .Type.GetType $.PackageRegistry }} `spanner:"{{ .ColumnName }}" json:"{{ .ColumnName }}"` // {{ .ColumnName }}
{{- end }}
{{- end }}
}

func {{ .Type.Name }}PrimaryKeys() []string {
     return []string{
{{- range .Type.PrimaryKeyFields }}
		"{{ .ColumnName }}",
{{- end }}
	}
}

func {{ .Type.Name }}Columns() []string {
	return []string{
{{- range .Type.Fields }}
		"{{ .ColumnName }}",
{{- end }}
	}
}

func {{ .Type.Name }}WritableColumns() []string {
	return []string{
{{- range .Type.Fields }}
	{{- if not .IsGenerated }}
		"{{ .ColumnName }}",
	{{- end }}
{{- end }}
	}
}

func ({{ $short }} *{{ .Type.Name }}) columnsToPtrs(cols []string) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		switch col {
{{- range .Type.Fields }}
		case "{{ .ColumnName }}":
			ret = append(ret, yoDecode(&{{ $short }}.{{ .Name }}))
{{- end }}
		default:
			return nil, fmt.Errorf("unknown column: %s", col)
		}
	}
	return ret, nil
}

func ({{ $short }} *{{ .Type.Name }}) columnsToValues(cols []string) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		switch col {
{{- range .Type.Fields }}
		case "{{ .ColumnName }}":
			ret = append(ret, yoEncode({{ $short }}.{{ .Name }}))
{{- end }}
		default:
			return nil, {{ .PackageRegistry.Use presetPackages.fmt "Errorf" }}("unknown column: %s", col)
		}
	}

	return ret, nil
}

// new{{ .Type.Name }}_Decoder returns a decoder which reads a row from *spanner.Row
// into {{ .Type.Name }}. The decoder is not goroutine-safe. Don't use it concurrently.
func new{{ .Type.Name }}_Decoder(cols []string) func(*{{ .PackageRegistry.Use presetPackages.goSpanner "Row" }}) (*{{ .Type.Name }}, error) {
	return func(row *{{ .PackageRegistry.Use presetPackages.goSpanner "Row" }}) (*{{ .Type.Name }}, error) {
        var {{ $short }} {{ .Type.Name }}
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
