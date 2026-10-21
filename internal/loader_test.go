package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mercari.io/yo/loaders"
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

func TestLoadSchema_UseIndexNameForFuncName(t *testing.T) {
	const ddl = `
CREATE TABLE Examples (
  PKey STRING(32) NOT NULL,
  Num INT64 NOT NULL,
  Name STRING(32) NOT NULL,
  CreatedAt TIMESTAMP NOT NULL,
) PRIMARY KEY(PKey);
CREATE UNIQUE INDEX ExamplesByNum ON Examples(Num) STORING (CreatedAt);
CREATE INDEX ExamplesByName ON Examples(Name);
CREATE INDEX Name ON Examples(Name, Num);

CREATE TABLE Others (
  PKey STRING(32) NOT NULL,
  Value INT64 NOT NULL,
) PRIMARY KEY(PKey);
CREATE INDEX OthersByValue ON Others(Value);
`
	path := filepath.Join(t.TempDir(), "schema.sql")
	if err := os.WriteFile(path, []byte(ddl), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    *ArgType
		want    map[string]string // index name -> func name
		wantErr error
		errText string
	}{
		{
			name: "default",
			args: &ArgType{IgnoreTables: []string{"Others"}},
			want: map[string]string{
				"ExamplesByNum":  "ExampleByCreatedAtNum",
				"ExamplesByName": "ExamplesByName",
				"Name":           "ExamplesByNameNum",
			},
		},
		{
			name: "specified index only, case insensitive",
			args: &ArgType{IgnoreTables: []string{"Others"}, UseIndexNameForFuncNameIndexes: []string{"examplesbynum"}},
			want: map[string]string{
				"ExamplesByNum":  "ExampleByExamplesByNum",
				"ExamplesByName": "ExamplesByName",
				"Name":           "ExamplesByNameNum",
			},
		},
		{
			name: "index of excluded table is not an error",
			args: &ArgType{IgnoreTables: []string{"Others"}, UseIndexNameForFuncNameIndexes: []string{"OthersByValue"}},
			want: map[string]string{
				"ExamplesByNum":  "ExampleByCreatedAtNum",
				"ExamplesByName": "ExamplesByName",
				"Name":           "ExamplesByNameNum",
			},
		},
		{
			name:    "index not found",
			args:    &ArgType{UseIndexNameForFuncNameIndexes: []string{"NoSuchIndex", "ExamplesByNum"}},
			wantErr: ErrIndexNotFound,
			errText: "NoSuchIndex",
		},
		{
			name:    "func name conflict",
			args:    &ArgType{IgnoreTables: []string{"Others"}, UseIndexNameForFuncNameIndexes: []string{"Name"}},
			errText: `index func name "ExamplesByName" conflicts`,
		},
		{
			name: "all indexes",
			args: &ArgType{IgnoreTables: []string{"Others"}, UseIndexNameForFuncName: true},
			want: map[string]string{
				"ExamplesByNum":  "ExampleByExamplesByNum",
				"ExamplesByName": "ExamplesByExamplesByName",
				"Name":           "ExamplesByName",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := loaders.NewSpannerLoaderFromDDL(path)
			if err != nil {
				t.Fatal(err)
			}
			inflector, err := NewInflector("")
			if err != nil {
				t.Fatal(err)
			}
			tl := NewTypeLoader(l, inflector)

			_, ixMap, err := tl.LoadSchema(tt.args)
			if tt.wantErr != nil || tt.errText != "" {
				if err == nil {
					t.Fatalf("want error, got nil")
				}
				if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
					t.Errorf("want error %v, got %v", tt.wantErr, err)
				}
				if !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("want error containing %q, got %v", tt.errText, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			got := map[string]string{}
			for _, ix := range ixMap {
				got[ix.Index.IndexName] = ix.FuncName
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
