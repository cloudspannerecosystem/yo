package generator

import (
	"io"
	"text/template"
)

// TemplateType represents a template type.
type TemplateType uint

// the order here will be the alter the output order per file.
const (
	TypeTemplate TemplateType = iota
	IndexTemplate

	// always last
	YOTemplate
)

// String returns the name for the associated template type.
func (tt TemplateType) String() string {
	var s string
	switch tt {
	case YOTemplate:
		s = "yo_db"
	case TypeTemplate:
		s = "type"
	case IndexTemplate:
		s = "index"
	default:
		panic("unknown TemplateType")
	}
	return s
}

var (
	// KnownTypeMap is the collection of known Go types.
	KnownTypeMap = map[string]bool{
		"bool":        true,
		"string":      true,
		"byte":        true,
		"rune":        true,
		"int":         true,
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
	Package string
}

// templateSet is a set of templates.
type templateSet struct {
	funcs template.FuncMap
	l     func(string) ([]byte, error)
	tpls  map[string]*template.Template
}

// Execute executes a specified template in the template set using the supplied
// obj as its parameters and writing the output to w.
func (ts *templateSet) Execute(w io.Writer, name string, obj interface{}) error {
	tpl, ok := ts.tpls[name]
	if !ok {
		// attempt to load and parse the template
		buf, err := ts.l(name)
		if err != nil {
			return err
		}

		// parse template
		tpl, err = template.New(name).Funcs(ts.funcs).Parse(string(buf))
		if err != nil {
			return err
		}
	}

	return tpl.Execute(w, obj)
}
