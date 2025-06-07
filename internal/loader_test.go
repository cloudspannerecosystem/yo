package internal

import (
	"fmt"
	"os"
	"sort"    // Added sort import
	"strings" // Added strings import
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kenshaw/snaker"
	"go.mercari.io/yo/models"
)

// mockLoaderImpl provides a mock implementation of the loaderImpl interface for testing.
type mockLoaderImpl struct {
	// Define fields to control mock behavior, e.g., data to return
	mockTableList         func() ([]*models.Table, error)
	mockColumnList        func(string) ([]*models.Column, error)
	mockIndexList         func(string) ([]*models.Index, error)
	mockIndexColumnList   func(string, string) ([]*models.IndexColumn, error)
	mockParseType         func(string, bool) (int, string, string)
	mockValidCustomType   func(string, string) bool
	mockParamN            func(int) string
	mockMaskFunc          func() string
}

func (m *mockLoaderImpl) ParamN(n int) string {
	if m.mockParamN != nil {
		return m.mockParamN(n)
	}
	return fmt.Sprintf("@p%d", n)
}

func (m *mockLoaderImpl) MaskFunc() string {
	if m.mockMaskFunc != nil {
		return m.mockMaskFunc()
	}
	return "?"
}

func (m *mockLoaderImpl) ParseType(dt string, nullable bool) (int, string, string) {
	if m.mockParseType != nil {
		return m.mockParseType(dt, nullable)
	}
	// Basic mock implementation
	if strings.Contains(dt, "STRING") {
		if nullable {
			return -1, "spanner.NullString{}", "spanner.NullString"
		}
		return -1, `""`, "string"
	}
	if strings.Contains(dt, "INT64") {
		if nullable {
			return -1, "spanner.NullInt64{}", "spanner.NullInt64"
		}
		return -1, "0", "int64"
	}
	return -1, "nil", "interface{}"
}

func (m *mockLoaderImpl) ValidCustomType(dataType string, customType string) bool {
	if m.mockValidCustomType != nil {
		return m.mockValidCustomType(dataType, customType)
	}
	return true
}

func (m *mockLoaderImpl) TableList() ([]*models.Table, error) {
	if m.mockTableList != nil {
		return m.mockTableList()
	}
	return []*models.Table{}, nil
}

func (m *mockLoaderImpl) ColumnList(table string) ([]*models.Column, error) {
	if m.mockColumnList != nil {
		return m.mockColumnList(table)
	}
	return []*models.Column{}, nil
}

func (m *mockLoaderImpl) IndexList(table string) ([]*models.Index, error) {
	if m.mockIndexList != nil {
		return m.mockIndexList(table)
	}
	return []*models.Index{}, nil
}

func (m *mockLoaderImpl) IndexColumnList(table string, index string) ([]*models.IndexColumn, error) {
	if m.mockIndexColumnList != nil {
		return m.mockIndexColumnList(table, index)
	}
	return []*models.IndexColumn{}, nil
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

func TestLoadIndexColumns_EmulatorIndexColumnSorting_V1(t *testing.T) {
	const (
		tableName      = "CompositePrimaryKeysByErrorTable"
		indexName      = "CompositePrimaryKeysByError3Index"
		emulatorEnvVar = "SPANNER_EMULATOR_HOST"
	)

	// Columns that would be part of the table definition
	tableColumns := []*models.Column{
		{ColumnName: "PKey1", DataType: "STRING", NotNull: true},
		{ColumnName: "PKey2", DataType: "STRING", NotNull: true},
		{ColumnName: "YError", DataType: "STRING", NotNull: true},
		{ColumnName: "ZError", DataType: "STRING", NotNull: true},
	}

	// Initial fields for the Type (matching tableColumns)
	typeFields := []*Field{}
	for _, col := range tableColumns {
		typeFields = append(typeFields, &Field{
			Name: snaker.ForceCamelIdentifier(col.ColumnName),
			Col:  col,
			// Other Field properties like Type, NilType, Len would be set by loader.ParseType
		})
	}

	mockedIndexColumnsUnsorted := []*models.IndexColumn{
		{ColumnName: "ZError", SeqNo: 1}, // Should be second after sorting
		{ColumnName: "YError", SeqNo: 2}, // Should be first after sorting
	}

	mock := &mockLoaderImpl{
		mockIndexColumnList: func(tbl, idx string) ([]*models.IndexColumn, error) {
			if tbl == tableName && idx == indexName {
				// Return a new slice to prevent modification by the sort in one run affecting the other
				colsToReturn := make([]*models.IndexColumn, len(mockedIndexColumnsUnsorted))
				copy(colsToReturn, mockedIndexColumnsUnsorted)

				// If emulator is set, mock should mimic the sorting behavior of SpannerLoader
				if os.Getenv(emulatorEnvVar) != "" {
					sort.Slice(colsToReturn, func(i, j int) bool {
						return colsToReturn[i].ColumnName < colsToReturn[j].ColumnName
					})
				}
				return colsToReturn, nil
			}
			return nil, fmt.Errorf("unexpected call to IndexColumnList: table %s, index %s", tbl, idx)
		},
		mockParseType: func(dt string, nullable bool) (int, string, string) {
			// Simplified ParseType for the test
			if strings.HasPrefix(dt, "STRING") { // Use HasPrefix for checking type family
				return -1, `""`, "string"
			}
			return -1, "nil", "interface{}"
		},
	}

	inf, err := NewInflector("") // Use empty string for default inflector
	if err != nil {
		t.Fatalf("NewInflector failed: %v", err)
	}
	typeLoader := NewTypeLoader(mock, inf)

	// Prepare the Type and Index structures that LoadIndexColumns expects
	typeTpl := &Type{
		Name:   tableName,
		Table:  &models.Table{TableName: tableName},
		Fields: typeFields, // Crucial: Fields must already be populated for LoadIndexColumns to find them
	}
	ixTpl := &Index{
		Type:   typeTpl,
		Index:  &models.Index{IndexName: indexName},
		Fields: []*Field{}, // This will be populated by LoadIndexColumns
	}

	// --- Test with SPANNER_EMULATOR_HOST set ---
	originalEmulatorHost, ok := os.LookupEnv(emulatorEnvVar)
	os.Setenv(emulatorEnvVar, "testhost:1234")
	if ok { // Correct defer logic
		defer os.Setenv(emulatorEnvVar, originalEmulatorHost)
	} else {
		defer os.Unsetenv(emulatorEnvVar)
	}

	err = typeLoader.LoadIndexColumns(&ArgType{}, ixTpl)
	if err != nil {
		t.Fatalf("LoadIndexColumns failed with emulator: %v", err)
	}

	expectedSortedColumns := []string{"YError", "ZError"}
	actualSortedColumns := make([]string, len(ixTpl.Fields))
	for i, f := range ixTpl.Fields {
		actualSortedColumns[i] = f.Col.ColumnName
	}

	if diff := cmp.Diff(expectedSortedColumns, actualSortedColumns); diff != "" {
		t.Errorf("Sorted index columns mismatch when %s is set (-want +got):\n%s", emulatorEnvVar, diff)
	}

	// --- Test with SPANNER_EMULATOR_HOST NOT set ---
	// Reset ixTpl.Fields for the next run
	ixTpl.Fields = []*Field{}
	ixTpl.StoringFields = []*Field{}
	ixTpl.NullableFields = []*Field{}

	// Correctly restore/unset env var for this part of the test
	// The previous defer handles the overall test cleanup. This ensures isolation for this specific sub-test.
	if ok {
		os.Setenv(emulatorEnvVar, originalEmulatorHost)
	} else {
		os.Unsetenv(emulatorEnvVar)
	}

	err = typeLoader.LoadIndexColumns(&ArgType{}, ixTpl)
	if err != nil {
		t.Fatalf("LoadIndexColumns failed without emulator: %v", err)
	}

	expectedUnsortedColumns := []string{"ZError", "YError"} // Order as returned by mock (without emulator sort)
	actualUnsortedColumns := make([]string, len(ixTpl.Fields))
	for i, f := range ixTpl.Fields {
		actualUnsortedColumns[i] = f.Col.ColumnName
	}

	if diff := cmp.Diff(expectedUnsortedColumns, actualUnsortedColumns); diff != "" {
		t.Errorf("Unsorted index columns mismatch when %s is not set (-want +got):\n%s", emulatorEnvVar, diff)
	}
}
