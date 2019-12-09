package models

// Table represents table info.
type Table struct {
	Type      string // type
	TableName string // table_name
	ManualPk  bool   // manual_pk
}

// Column represents column info.
type Column struct {
	FieldOrdinal int    // field_ordinal
	ColumnName   string // column_name
	DataType     string // data_type
	NotNull      bool   // not_null
	IsPrimaryKey bool   // is_primary_key
}

// Index represents an index.
type Index struct {
	IndexName string // index_name
	IsUnique  bool   // is_unique
	IsPrimary bool   // is_primary
	SeqNo     int    // seq_no
	Origin    string // origin
	IsPartial bool   // is_partial
}

// IndexColumn represents index column info.
type IndexColumn struct {
	SeqNo      int    // seq_no. If is'a Storing Column, this value is 0.
	ColumnName string // column_name
	Storing    bool   // storing column or not
}

// CustomTypes represents custom type definitions
type CustomTypes struct {
	Tables []struct {
		Name    string            `yaml:"name"`
		Columns map[string]string `yaml:"columns"`
	}
}
