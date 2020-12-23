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
	"sort"
	"strings"

	"golang.org/x/tools/imports"
)

// importsOptions is the same as x/tools/cmd/goimports options except Fragment.
var importsOptions = &imports.Options{
	TabWidth:  8,
	TabIndent: true,
	Comments:  true,
}

type FileBuffer struct {
	FileName string
	BaseName string

	Header []byte
	Chunks []*TBuf

	TempDir      string
	TempFilePath string
}

func (f *FileBuffer) WriteTempFile() error {
	file, err := ioutil.TempFile(f.TempDir, fmt.Sprintf("%s_*", f.BaseName))
	if err != nil {
		return fmt.Errorf("failed to create temp file for %s: %v", f.BaseName, err)
	}

	if err := f.writeChunks(file); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to write temp file for %s: %v", f.BaseName, err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file for %s: %v", f.BaseName, err)
	}

	f.TempFilePath = file.Name()
	return nil
}

func (f *FileBuffer) writeChunks(file *os.File) error {
	// write a header to the file
	if _, err := file.Write(f.Header); err != nil {
		return err
	}

	chunks := TBufSlice(f.Chunks)

	// sort chunks
	sort.Sort(chunks)

	// write chunks to the file in order
	for _, chunk := range chunks {
		// check if generated template is only whitespace/empty
		bufStr := strings.TrimSpace(chunk.Buf.String())
		if len(bufStr) == 0 {
			continue
		}

		if _, err := chunk.Buf.WriteTo(file); err != nil {
			return err
		}
	}

	return nil
}

func (f *FileBuffer) Postprocess() error {
	// run gofmt for the temp file
	formatted, err := imports.Process(f.TempFilePath, nil, importsOptions)
	if err != nil {
		return fmt.Errorf("failed to fmt file for %s: %v", f.BaseName, err)
	}

	// overwrite the tempfile by gofmt result
	// since abs file exists, set perm to 0
	if err := ioutil.WriteFile(f.TempFilePath, formatted, 0); err != nil {
		return fmt.Errorf("failed to formatted file for %s: %v", f.BaseName, err)
	}

	// change permission
	if err := os.Chmod(f.TempFilePath, 0666); err != nil {
		return fmt.Errorf("failed to change file permission for %s: %v", f.BaseName, err)
	}

	return nil
}

func (f *FileBuffer) Finalize() error {
	if err := os.Rename(f.TempFilePath, f.FileName); err != nil {
		return fmt.Errorf("failed to put file for %s: %v", f.BaseName, err)
	}

	return nil
}

// TBuf is to hold the executed templates.
type TBuf struct {
	Name    string
	Subname string
	Buf     *bytes.Buffer
}

// TBufSlice is a slice of TBuf compatible with sort.Interface.
type TBufSlice []*TBuf

func (t TBufSlice) Len() int {
	return len(t)
}

func (t TBufSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TBufSlice) Less(i, j int) bool {
	if strings.Compare(t[i].Name, t[j].Name) < 0 {
		return true
	} else if strings.Compare(t[j].Name, t[i].Name) < 0 {
		return false
	}

	return strings.Compare(t[i].Subname, t[j].Subname) < 0
}
