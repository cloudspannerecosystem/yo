# Field list of {{ .Type.Name }}

{{ range .Type.Fields -}}
* {{ .Name }} {{ .SpannerDataType }} {{ .Type.GetType $.PackageRegistry }}
{{ end }}
# Primary Key

{{ range .Type.PrimaryKeyFields -}}
* {{ .Name }} {{ .SpannerDataType }} {{ .Type.GetType $.PackageRegistry }}
{{ end }}
# Index list of {{ .Type.Name }}

{{ range .Type.Indexes -}}
* {{ .IndexName }}
{{ end -}}
