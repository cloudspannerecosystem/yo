# Field list of {{ .Name }}

{{ range .Fields -}}
* {{ .Name }} {{ .Col.DataType }} {{ .Type }}
{{ end }}
# Primary Key

{{ range .PrimaryKeyFields -}}
* {{ .Name }} {{ .Col.DataType }} {{ .Type }}
{{ end }}
# Index list of {{ .Name }}

{{ range .Indexes -}}
* {{ .Index.IndexName }}
{{ end -}}
