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
	"fmt"
	"io"
	"text/template"

	"go.mercari.io/yo/v2/models"
	"go.mercari.io/yo/v2/module"
)

var (
	// KnownTypeMap is the collection of known Go types.
	KnownTypeMap = map[string]bool{
		"bool":        true,
		"string":      true,
		"byte":        true,
		"rune":        true,
		"int":         true,
		"int8":        true,
		"int16":       true,
		"int32":       true,
		"int64":       true,
		"uint":        true,
		"uint8":       true,
		"uint16":      true,
		"uint32":      true,
		"uint64":      true,
		"float32":     true,
		"float64":     true,
		"Slice":       true,
		"StringSlice": true,
	}

	ShortNameTypeMap = map[string]string{
		"bool":    "b",
		"string":  "s",
		"byte":    "b",
		"rune":    "r",
		"int":     "i",
		"int8":    "i",
		"int16":   "i",
		"int32":   "i",
		"int64":   "i",
		"uint":    "u",
		"uint8":   "u",
		"uint16":  "u",
		"uint32":  "u",
		"uint64":  "u",
		"float32": "f",
		"float64": "f",
	}

	ConflictedShortNames = map[string]bool{
		"context":  true,
		"errors":   true,
		"fmt":      true,
		"regexp":   true,
		"strings":  true,
		"time":     true,
		"iterator": true,
		"spanner":  true,
		"civil":    true,
		"codes":    true,
		"status":   true,
	}
)

// basicDataSet is used for template data for yo_db and yo_package.
type basicDataSet struct {
	BuildTag string
	Package  string
	Schema   *models.Schema
}

// templateSet is a set of templates.
type templateSet struct {
	funcs template.FuncMap
}

// Execute executes a specified template in the template set using the supplied
// obj as its parameters and writing the output to w.
func (ts *templateSet) Execute(w io.Writer, mod module.Module, obj interface{}) error {
	buf, err := mod.Load()
	if err != nil {
		return fmt.Errorf("Load module(%s): %v", mod.Name(), err)
	}

	// parse template
	tpl, err := template.New(mod.Name()).Funcs(ts.funcs).Parse(string(buf))
	if err != nil {
		return fmt.Errorf("Parse module(%s): %v", mod.Name(), err)
	}

	if err := tpl.Execute(w, obj); err != nil {
		return fmt.Errorf("Execute module(%s): %v", mod.Name(), err)
	}

	return nil
}
