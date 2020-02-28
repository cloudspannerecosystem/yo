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
	"strings"
)

// TBuf is to hold the executed templates.
type TBuf struct {
	TemplateType TemplateType
	Name         string
	Subname      string
	Buf          *bytes.Buffer
}

// TBufSlice is a slice of TBuf compatible with sort.Interface.
type TBufSlice []TBuf

func (t TBufSlice) Len() int {
	return len(t)
}

func (t TBufSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TBufSlice) Less(i, j int) bool {
	if t[i].TemplateType < t[j].TemplateType {
		return true
	} else if t[j].TemplateType < t[i].TemplateType {
		return false
	}

	if strings.Compare(t[i].Name, t[j].Name) < 0 {
		return true
	} else if strings.Compare(t[j].Name, t[i].Name) < 0 {
		return false
	}

	return strings.Compare(t[i].Subname, t[j].Subname) < 0
}
