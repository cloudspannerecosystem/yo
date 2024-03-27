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

package generator

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/tools/imports"

	"go.mercari.io/yo/internal"
	templates "go.mercari.io/yo/tplbin"
)

// Loader is the common interface for database drivers that can generate code
// from a database schema.
type Loader interface {
	// NthParam returns the 0-based Nth param for the Loader.
	NthParam(i int) string
}

type GeneratorOption struct {
	PackageName             string
	Tags                    string
	TemplatePath            string
	CustomTypePackage       string
	FilenameSuffix          string
	SingleFile              bool
	Filename                string
	FilenameWithUnderscores bool
	Path                    string
}

func NewGenerator(loader Loader, inflector internal.Inflector, opt GeneratorOption) *Generator {
	return &Generator{
		loader:                  loader,
		inflector:               inflector,
		templatePath:            opt.TemplatePath,
		nameConflictSuffix:      "z",
		packageName:             opt.PackageName,
		tags:                    opt.Tags,
		customTypePackage:       opt.CustomTypePackage,
		filenameSuffix:          opt.FilenameSuffix,
		singleFile:              opt.SingleFile,
		filename:                opt.Filename,
		FilenameWithUnderscores: opt.FilenameWithUnderscores,
		path:                    opt.Path,
		files:                   make(map[string]*os.File),
	}
}

type Generator struct {
	loader       Loader
	inflector    internal.Inflector
	templatePath string

	// files is a map of filenames to open file handles.
	files map[string]*os.File

	// generated is the generated templates after a run.
	generated []TBuf

	packageName             string
	tags                    string
	customTypePackage       string
	filenameSuffix          string
	singleFile              bool
	filename                string
	FilenameWithUnderscores bool
	path                    string

	nameConflictSuffix string
}

func (g *Generator) newTemplateSet() *templateSet {
	return &templateSet{
		funcs: g.newTemplateFuncs(),
		l:     g.templateLoader,
		tpls:  map[string]*template.Template{},
	}
}

// TemplateLoader loads templates from the specified name.
func (g *Generator) templateLoader(name string) ([]byte, error) {
	// no template path specified
	if g.templatePath == "" {
		f, err := templates.Assets.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}

	return ioutil.ReadFile(path.Join(g.templatePath, name))
}

func (g *Generator) Generate(tableMap map[string]*internal.Type, ixMap map[string]*internal.Index) error {
	// generate table templates
	for _, t := range tableMap {
		if err := g.ExecuteTemplate(TypeTemplate, t.Name, "", t); err != nil {
			return err
		}
	}

	// generate index templates
	for _, ix := range ixMap {
		if err := g.ExecuteTemplate(IndexTemplate, ix.Type.Name, ix.Index.IndexName, ix); err != nil {
			return err
		}
	}

	ds := &basicDataSet{
		Package:  g.packageName,
		TableMap: tableMap,
	}

	if err := g.ExecuteTemplate(YOTemplate, "yo_db", "", ds); err != nil {
		return err
	}

	if err := g.writeTypes(ds); err != nil {
		return err
	}

	return nil
}

// getFile builds the filepath from the TBuf information, and retrieves the
// file from files. If the built filename is not already defined, then it calls
// the os.OpenFile with the correct parameters depending on the state of args.
func (g *Generator) getFile(ds *basicDataSet, t *TBuf) (*os.File, error) {
	// determine filename
	var filename string
	switch {
	case g.singleFile:
		filename = g.filename
	case g.FilenameWithUnderscores:
		filename = toSnakeCase(t.Name) + g.filenameSuffix
	default:
		filename = strings.ToLower(t.Name) + g.filenameSuffix
	}
	filename = path.Join(g.path, filename)

	// lookup file
	f, ok := g.files[filename]
	if ok {
		return f, nil
	}

	// default open mode
	mode := os.O_RDWR | os.O_CREATE | os.O_TRUNC

	// stat file to determine if file is a directory.
	fi, err := os.Stat(filename)
	if err == nil && fi.IsDir() {
		return nil, errors.New("filename cannot be directory")
	}

	// open file
	f, err = os.OpenFile(filename, mode, 0o666)
	if err != nil {
		return nil, err
	}

	// add build tags
	if g.tags != "" {
		_, _ = f.WriteString(`// +build ` + g.tags + "\n\n")
	}

	// execute
	if err := g.newTemplateSet().Execute(f, "yo_package.go.tpl", ds); err != nil {
		return nil, err
	}

	// store file
	g.files[filename] = f

	return f, nil
}

func toSnakeCase(s string) string {
	b := &strings.Builder{}
	for i, r := range s {
		if i == 0 {
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		if unicode.IsUpper(r) {
			b.WriteRune('_')
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// importsOptions is the same as x/tools/cmd/goimports options except Fragment.
var importsOptions = &imports.Options{
	TabWidth:  8,
	TabIndent: true,
	Comments:  true,
}

// writeTypes writes the generated definitions.
func (g *Generator) writeTypes(ds *basicDataSet) error {
	var err error

	out := TBufSlice(g.generated)

	// sort segments
	sort.Sort(out)

	// loop, writing in order
	for _, t := range out {
		// check if generated template is only whitespace/empty
		bufStr := strings.TrimSpace(t.Buf.String())
		if len(bufStr) == 0 {
			continue
		}

		var f *os.File
		// get file and filename
		f, err = g.getFile(ds, &t)
		if err != nil {
			return err
		}

		// write segment
		if _, err = t.Buf.WriteTo(f); err != nil {
			return err
		}
	}

	// format by imports, closing files
	for k, f := range g.files {
		// close
		err = f.Close()
		if err != nil {
			return err
		}

		// imports.Process needs absolute filepath for accurate fix
		abs, err := filepath.Abs(k)
		if err != nil {
			return err
		}
		// format
		formatted, err := imports.Process(abs, nil, importsOptions)
		if err != nil {
			return err
		}

		// since abs file exists, set perm to 0
		if err := ioutil.WriteFile(abs, formatted, 0); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteTemplate loads and parses the supplied template with name and
// executes it with obj as the context.
func (g *Generator) ExecuteTemplate(tt TemplateType, name string, sub string, obj interface{}) error {
	var err error

	// setup generated
	if g.generated == nil {
		g.generated = []TBuf{}
	}

	// create store
	v := TBuf{
		TemplateType: tt,
		Name:         name,
		Subname:      sub,
		Buf:          new(bytes.Buffer),
	}

	// build template name
	templateName := fmt.Sprintf("%s.go.tpl", tt)

	// execute template
	err = g.newTemplateSet().Execute(v.Buf, templateName, obj)
	if err != nil {
		return err
	}

	g.generated = append(g.generated, v)
	return nil
}
