# Field list of {{ .Name }}

{{ range .Fields -}}
* {{ .Name }} {{ .SpannerDataType }} {{ .Type }}
{{ end }}
# Primary Key

{{ range .PrimaryKeyFields -}}
* {{ .Name }} {{ .SpannerDataType }} {{ .Type }}
{{ end }}
# Index list of {{ .Name }}

{{ range .Indexes -}}
* {{ .IndexName }}
{{ end -}}
