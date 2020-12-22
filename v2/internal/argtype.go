// Copyright (c) 2020 Mercari, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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

	Path     string
	Filename string

	// DDLFilepath is the filepath of the ddl file.
	DDLFilepath string

	// FromDDL indicates generating from ddl flie or not.
	FromDDL bool

	// InflectionRuleFile is custom inflection rule file.
	InflectionRuleFile string
}
