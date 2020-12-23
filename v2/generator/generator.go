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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/module"
)

// Loader is the common interface for database drivers that can generate code
// from a database schema.
type Loader interface {
	// NthParam returns the 0-based Nth param for the Loader.
	NthParam(i int) string
}

type GeneratorOption struct {
	PackageName       string
	Tags              string
	CustomTypePackage string
	FilenameSuffix    string
	BaseDir           string

	HeaderModule  module.Module
	GlobalModules []module.Module
	TypeModules   []module.Module
}

func NewGenerator(loader Loader, inflector internal.Inflector, opt GeneratorOption) *Generator {
	return &Generator{
		loader:            loader,
		inflector:         inflector,
		packageName:       opt.PackageName,
		tags:              opt.Tags,
		customTypePackage: opt.CustomTypePackage,
		filenameSuffix:    opt.FilenameSuffix,
		baseDir:           opt.BaseDir,

		headerModule:  opt.HeaderModule,
		globalModules: opt.GlobalModules,
		typeModules:   opt.TypeModules,

		files:              make(map[string]*FileBuffer),
		nameConflictSuffix: "z",
	}
}

type Generator struct {
	loader            Loader
	inflector         internal.Inflector
	packageName       string
	tags              string
	customTypePackage string
	filenameSuffix    string
	filename          string
	baseDir           string
	tempDir           string

	headerModule  module.Module
	globalModules []module.Module
	typeModules   []module.Module

	files              map[string]*FileBuffer
	nameConflictSuffix string
}

func (g *Generator) newTemplateSet() *templateSet {
	return &templateSet{
		funcs: g.newTemplateFuncs(),
	}
}

func (g *Generator) Generate(schema *internal.Schema) error {
	tempDir, err := ioutil.TempDir("", "yo_")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}

	// create a workspace for the code generation and cleanup it after done
	g.tempDir = tempDir
	defer os.RemoveAll(g.tempDir)

	// execute type modules
	for _, mod := range g.typeModules {
		for _, tbl := range schema.Types {
			if err := g.ExecuteTemplate(mod, tbl.Name, tbl); err != nil {
				return err
			}
		}
	}

	ds := &basicDataSet{
		BuildTag: g.tags,
		Package:  g.packageName,
		Schema:   schema,
	}

	// execute global modules
	for _, mod := range g.globalModules {
		if err := g.ExecuteTemplate(mod, mod.Name(), ds); err != nil {
			return err
		}
	}

	if err := g.writeFiles(ds); err != nil {
		return err
	}

	return nil
}

func (g *Generator) getFile(name string) *FileBuffer {
	var filename = strings.ToLower(name) + g.filenameSuffix
	filename = path.Join(g.baseDir, filename)

	f, ok := g.files[filename]
	if ok {
		return f
	}

	file := &FileBuffer{
		FileName: filename,
		BaseName: name,
		TempDir:  g.tempDir,
	}

	g.files[filename] = file
	return file
}

// writeFiles writes the generated definitions.
func (g *Generator) writeFiles(ds *basicDataSet) error {
	for _, file := range g.files {
		if err := g.ExecuteHeaderTemplate(g.headerModule, file, ds); err != nil {
			return err
		}

		if err := file.WriteTempFile(); err != nil {
			return err
		}
	}

	for _, file := range g.files {
		if err := file.Postprocess(); err != nil {
			return err
		}
	}

	for _, file := range g.files {
		if err := file.Finalize(); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteTemplate loads and parses the supplied template with name and
// executes it with obj as the context.
func (g *Generator) ExecuteTemplate(mod module.Module, name string, obj interface{}) error {
	file := g.getFile(name)
	tbuf := TBuf{
		Name: name,
		Buf:  new(bytes.Buffer),
	}

	// execute template
	if err := g.newTemplateSet().Execute(tbuf.Buf, mod, obj); err != nil {
		return fmt.Errorf("template error: %v", err)
	}

	file.Chunks = append(file.Chunks, &tbuf)
	return nil
}

func (g *Generator) ExecuteHeaderTemplate(mod module.Module, file *FileBuffer, obj interface{}) error {
	buf := new(bytes.Buffer)

	if err := g.newTemplateSet().Execute(buf, mod, obj); err != nil {
		return err
	}

	file.Header = buf.Bytes()
	return nil
}
