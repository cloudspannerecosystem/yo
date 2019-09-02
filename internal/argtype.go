package internal

// ArgType is the type that specifies the command line arguments.
type ArgType struct {
	// Project is the GCP project string
	Project string

	// Instance is the instance string
	Instance string

	// Database is the database string
	Database string

	// CustomTypesFile is the path for custom table field type definition file (xx.yml)
	CustomTypesFile string

	// Out is the output path. If Out is a file, then that will be used as the
	// path. If Out is a directory, then the output file will be
	// Out/<$CWD>.yo.go
	Out string

	// Suffix is the output suffix for filenames.
	Suffix string

	// SingleFile when toggled changes behavior so that output is to one f ile.
	SingleFile bool

	// Package is the name used to generate package headers. If not specified,
	// the name of the output directory will be used instead.
	Package string

	// CustomTypePackage is the Go package name to use for unknown types.
	CustomTypePackage string

	// IgnoreFields allows the user to specify field names which should not be
	// handled by yo in the generated code.
	IgnoreFields []string

	// IgnoreTables allows the user to specify table names which should not be
	// handled by yo in the generated code.
	IgnoreTables []string

	// TemplatePath is the path to use the user supplied templates instead of
	// the built in versions.
	TemplatePath string

	// Tags is the list of build tags to add to generated Go files.
	Tags string

	// CreateTemplates changes command behavior.
	// If true, default template files are created to TemplatePath.
	CreateTemplates bool

	Path     string
	Filename string
}
