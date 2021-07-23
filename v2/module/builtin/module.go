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

package builtin

import (
	"fmt"
	"io/ioutil"

	"go.mercari.io/yo/v2/module"
	templates "go.mercari.io/yo/v2/module/builtin/tplbin"
)

type builtinMod struct {
	typ  module.ModuleType
	name string
}

func newBuiltin(typ module.ModuleType, name string) module.Module {
	return &builtinMod{
		typ:  typ,
		name: name,
	}
}

func (m *builtinMod) Name() string {
	return m.name
}

func (m *builtinMod) Type() module.ModuleType {
	return m.typ
}

func (m *builtinMod) Load() ([]byte, error) {
	f, err := templates.Assets.Open(fmt.Sprintf("%s.go.tpl", m.name))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", m.name, err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from assets: %w", err)
	}

	return b, nil
}
