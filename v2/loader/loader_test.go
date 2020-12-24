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

package loader

import (
	"fmt"
	"testing"

	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/models"
)

func Test_setIndexesToTables(t *testing.T) {
	tests := []struct {
		table  map[string]*internal.Type
		ix     map[string]*internal.Index
		result map[string]int
	}{
		{
			table: map[string]*internal.Type{
				"TableA": &internal.Type{
					Indexes: []*internal.Index{},
				},
			},
			ix: map[string]*internal.Index{
				"TableA_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableA_Index2": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
			},
			result: map[string]int{
				"TableA": 2,
			},
		},
		{
			table: map[string]*internal.Type{
				"TableA": &internal.Type{
					Indexes: []*internal.Index{},
				},
				"TableB": &internal.Type{
					Indexes: []*internal.Index{},
				},
			},
			ix: map[string]*internal.Index{
				"TableA_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableA_Index2": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
			},
			result: map[string]int{
				"TableA": 2,
				"TableB": 0,
			},
		},
		{
			table: map[string]*internal.Type{
				"TableA": &internal.Type{
					Indexes: []*internal.Index{},
				},
				"TableB": &internal.Type{
					Indexes: []*internal.Index{},
				},
			},
			ix: map[string]*internal.Index{
				"TableA_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableA_Index2": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableA"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
				"TableB_Index1": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableB"},
					},
					Index: &models.Index{IndexName: "Index1"},
				},
				"TableB_Index2forTableA_Hoge": &internal.Index{
					Type: &internal.Type{
						Table: &models.Table{TableName: "TableB"},
					},
					Index: &models.Index{IndexName: "Index2"},
				},
			},
			result: map[string]int{
				"TableA": 2,
				"TableB": 2,
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case:%d", i), func(t *testing.T) {
			setIndexesToTables(tt.table, tt.ix)
			for k, v := range tt.table {
				if len(v.Indexes) != tt.result[k] {
					t.Errorf("error. want:%d got:%d", tt.result[k], len(v.Indexes))
				}
			}
		})
	}
}
