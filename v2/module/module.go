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

package module

import (
	"fmt"
	"os"
)

// ModuleType represents a module type.
type ModuleType uint

const (
	GlobalModule ModuleType = iota
	TypeModule
	HeaderModule
)

type Module interface {
	Type() ModuleType
	Name() string
	Load() ([]byte, error)
}

type module struct {
	typ  ModuleType
	name string
	path string
}

func New(typ ModuleType, name string, path string) Module {
	return &module{
		typ:  typ,
		name: name,
		path: path,
	}
}

func (m *module) Name() string {
	return m.name
}

func (m *module) Type() ModuleType {
	return m.typ
}

func (m *module) Load() ([]byte, error) {
	b, err := os.ReadFile(m.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", m.path, err)
	}

	return b, nil
}
