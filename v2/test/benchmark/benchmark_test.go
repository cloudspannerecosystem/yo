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

package test

import (
	"fmt"
	"testing"

	"cloud.google.com/go/spanner"
	"github.com/google/go-cmp/cmp"
)

type BenchmarkModel struct {
	Col0 string `spanner:"Col0"`
	Col1 string `spanner:"Col1"`
	Col2 string `spanner:"Col2"`
	Col3 string `spanner:"Col3"`
	Col4 string `spanner:"Col4"`
	Col5 string `spanner:"Col5"`
	Col6 string `spanner:"Col6"`
	Col7 string `spanner:"Col7"`
	Col8 string `spanner:"Col8"`
	Col9 string `spanner:"Col9"`
}

var benchmarkModelColumns = []string{
	"Col0",
	"Col1",
	"Col2",
	"Col3",
	"Col4",
	"Col5",
	"Col6",
	"Col7",
	"Col8",
	"Col9",
}

func (cpk *BenchmarkModel) columnsToPtrs(cols []string) ([]interface{}, error) {
	ret := make([]interface{}, 0, len(cols))
	for _, col := range cols {
		switch col {
		case "Col0":
			ret = append(ret, &cpk.Col0)
		case "Col1":
			ret = append(ret, &cpk.Col1)
		case "Col2":
			ret = append(ret, &cpk.Col2)
		case "Col3":
			ret = append(ret, &cpk.Col3)
		case "Col4":
			ret = append(ret, &cpk.Col4)
		case "Col5":
			ret = append(ret, &cpk.Col5)
		case "Col6":
			ret = append(ret, &cpk.Col6)
		case "Col7":
			ret = append(ret, &cpk.Col7)
		case "Col8":
			ret = append(ret, &cpk.Col8)
		case "Col9":
			ret = append(ret, &cpk.Col9)

		default:
			return nil, fmt.Errorf("unknown column: %s", col)
		}
	}
	return ret, nil
}

func newBenchmarkModel_Decoder(cols []string) func(*spanner.Row) (*BenchmarkModel, error) {
	return func(row *spanner.Row) (*BenchmarkModel, error) {
		var cpk BenchmarkModel
		ptrs, err := cpk.columnsToPtrs(cols)
		if err != nil {
			return nil, err
		}

		if err := row.Columns(ptrs...); err != nil {
			return nil, err
		}

		return &cpk, nil
	}
}

func testRow(t testing.TB) *spanner.Row {
	var colVals []interface{}
	for i := 0; i < 10; i++ {
		colVals = append(colVals, fmt.Sprintf("val%d", i))
	}
	row, err := spanner.NewRow(benchmarkModelColumns, colVals)
	if err != nil {
		t.Fatalf("new row: %v", err)
	}

	return row
}

func TestResult(t *testing.T) {
	row := testRow(t)

	var m1 BenchmarkModel
	if err := row.ToStruct(&m1); err != nil {
		t.Fatalf("error: %v", err)
	}

	decoder := newBenchmarkModel_Decoder(benchmarkModelColumns)
	m2, err := decoder(row)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if diff := cmp.Diff(&m1, m2); diff != "" {
		t.Errorf("(-m1, +m2)\n%s", diff)
	}
}

func BenchmarkToStruct(b *testing.B) {
	row := testRow(b)
	models := make([]BenchmarkModel, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := row.ToStruct(&models[i]); err != nil {
			b.Fatalf("error: %v", err)
		}
	}
}

func BenchmarkDecoder(b *testing.B) {
	row := testRow(b)
	decoder := newBenchmarkModel_Decoder(benchmarkModelColumns)
	models := make([]*BenchmarkModel, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, err := decoder(row)
		if err != nil {
			b.Fatalf("error: %v", err)
		}
		models[i] = m
	}
}
