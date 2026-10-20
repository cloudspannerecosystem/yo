package internal

import (
	"fmt"
	"testing"

	"go.mercari.io/yo/models"
)

func Test_buildIndexFuncName(t *testing.T) {
	inflector, err := NewInflector("")
	if err != nil {
		t.Fatal(err)
	}
	tl := NewTypeLoader(nil, inflector)

	field := func(name string) *Field {
		return &Field{Name: name}
	}

	tests := []struct {
		name          string
		ix            *Index
		want          string
		wantIndexName string
	}{
		{
			name: "unique index with storing",
			ix: &Index{
				Type:          &Type{Name: "Example"},
				Fields:        []*Field{field("Num")},
				StoringFields: []*Field{field("CreatedAt")},
				Index:         &models.Index{IndexName: "ExamplesByNum", IsUnique: true},
			},
			want:          "ExampleByCreatedAtNum",
			wantIndexName: "ExampleByExamplesByNum",
		},
		{
			name: "non unique index",
			ix: &Index{
				Type:   &Type{Name: "Example"},
				Fields: []*Field{field("Num"), field("Name")},
				Index:  &models.Index{IndexName: "ExamplesByNumName"},
			},
			want:          "ExamplesByNumName",
			wantIndexName: "ExamplesByExamplesByNumName",
		},
		{
			name: "snake case index name",
			ix: &Index{
				Type:   &Type{Name: "Example"},
				Fields: []*Field{field("UserID")},
				Index:  &models.Index{IndexName: "examples_by_user_id", IsUnique: true},
			},
			want:          "ExampleByUserID",
			wantIndexName: "ExampleByExamplesByUserID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tl.buildIndexFuncName(tt.ix); got != tt.want {
				t.Errorf("buildIndexFuncName want:%s got:%s", tt.want, got)
			}
			if got := tl.buildIndexFuncNameFromIndexName(tt.ix); got != tt.wantIndexName {
				t.Errorf("buildIndexFuncNameFromIndexName want:%s got:%s", tt.wantIndexName, got)
			}
		})
	}
}

func Test_setIndexesToTables(t *testing.T) {
	tests := []struct {
		table  map[string]*Type
		ix     map[string]*Index
		result map[string]int
	}{
		{
			table: map[string]*Type{
				"TableA": &Type{
					Indexes: []*Index{},
				},
			},
			ix: map[string]*Index{
				"TableA_Index1": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableA",
						},
					},
				},
				"TableA_Index2": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableA",
						},
					},
				},
			},
			result: map[string]int{
				"TableA": 2,
			},
		},
		{
			table: map[string]*Type{
				"TableA": &Type{
					Indexes: []*Index{},
				},
				"TableB": &Type{
					Indexes: []*Index{},
				},
			},
			ix: map[string]*Index{
				"TableA_Index1": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableA",
						},
					},
				},
				"TableA_Index2": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableA",
						},
					},
				},
			},
			result: map[string]int{
				"TableA": 2,
				"TableB": 0,
			},
		},
		{
			table: map[string]*Type{
				"TableA": &Type{
					Indexes: []*Index{},
				},
				"TableB": &Type{
					Indexes: []*Index{},
				},
			},
			ix: map[string]*Index{
				"TableA_Index1": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableA",
						},
					},
				},
				"TableA_Index2": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableA",
						},
					},
				},
				"TableB_Index1": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableB",
						},
					},
				},
				"TableB_Index2forTableA_Hoge": &Index{
					Type: &Type{
						Table: &models.Table{
							TableName: "TableB",
						},
					},
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

func Test_useIndexNameForFuncName(t *testing.T) {
	tests := []struct {
		name      string
		args      *ArgType
		indexName string
		want      bool
	}{
		{name: "default", args: &ArgType{}, indexName: "IndexA", want: false},
		{name: "all", args: &ArgType{UseIndexNameForFuncName: true}, indexName: "IndexA", want: true},
		{name: "specified", args: &ArgType{UseIndexNameForFuncNameIndexes: []string{"IndexA", "IndexB"}}, indexName: "IndexB", want: true},
		{name: "not specified", args: &ArgType{UseIndexNameForFuncNameIndexes: []string{"IndexA"}}, indexName: "IndexB", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := useIndexNameForFuncName(tt.args, tt.indexName); got != tt.want {
				t.Errorf("want:%v got:%v", tt.want, got)
			}
		})
	}
}
