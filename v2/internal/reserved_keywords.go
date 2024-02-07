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

package internal

// This list was created with reference to https://cloud.google.com/spanner/docs/reference/standard-sql/lexical#reserved_keywords
var reservedKeywords = map[string]struct{}{
	"ALL":                  struct{}{},
	"AND":                  struct{}{},
	"ANY":                  struct{}{},
	"ARRAY":                struct{}{},
	"AS":                   struct{}{},
	"ASC":                  struct{}{},
	"ASSERT_ROWS_MODIFIED": struct{}{},
	"AT":                   struct{}{},
	"BETWEEN":              struct{}{},
	"BY":                   struct{}{},
	"CASE":                 struct{}{},
	"CAST":                 struct{}{},
	"COLLATE":              struct{}{},
	"CONTAINS":             struct{}{},
	"CREATE":               struct{}{},
	"CROSS":                struct{}{},
	"CUBE":                 struct{}{},
	"CURRENT":              struct{}{},
	"DEFAULT":              struct{}{},
	"DEFINE":               struct{}{},
	"DESC":                 struct{}{},
	"DISTINCT":             struct{}{},
	"ELSE":                 struct{}{},
	"END":                  struct{}{},
	"ENUM":                 struct{}{},
	"ESCAPE":               struct{}{},
	"EXCEPT":               struct{}{},
	"EXCLUDE":              struct{}{},
	"EXISTS":               struct{}{},
	"EXTRACT":              struct{}{},
	"FALSE":                struct{}{},
	"FETCH":                struct{}{},
	"FOLLOWING":            struct{}{},
	"FOR":                  struct{}{},
	"FROM":                 struct{}{},
	"FULL":                 struct{}{},
	"GROUP":                struct{}{},
	"GROUPING":             struct{}{},
	"GROUPS":               struct{}{},
	"HASH":                 struct{}{},
	"HAVING":               struct{}{},
	"IF":                   struct{}{},
	"IGNORE":               struct{}{},
	"IN":                   struct{}{},
	"INNER":                struct{}{},
	"INTERSECT":            struct{}{},
	"INTERVAL":             struct{}{},
	"INTO":                 struct{}{},
	"IS":                   struct{}{},
	"JOIN":                 struct{}{},
	"LATERAL":              struct{}{},
	"LEFT":                 struct{}{},
	"LIKE":                 struct{}{},
	"LIMIT":                struct{}{},
	"LOOKUP":               struct{}{},
	"MERGE":                struct{}{},
	"NATURAL":              struct{}{},
	"NEW":                  struct{}{},
	"NO":                   struct{}{},
	"NOT":                  struct{}{},
	"NULL":                 struct{}{},
	"NULLS":                struct{}{},
	"OF":                   struct{}{},
	"ON":                   struct{}{},
	"OR":                   struct{}{},
	"ORDER":                struct{}{},
	"OUTER":                struct{}{},
	"OVER":                 struct{}{},
	"PARTITION":            struct{}{},
	"PRECEDING":            struct{}{},
	"PROTO":                struct{}{},
	"RANGE":                struct{}{},
	"RECURSIVE":            struct{}{},
	"RESPECT":              struct{}{},
	"RIGHT":                struct{}{},
	"ROLLUP":               struct{}{},
	"ROWS":                 struct{}{},
	"SELECT":               struct{}{},
	"SET":                  struct{}{},
	"SOME":                 struct{}{},
	"STRUCT":               struct{}{},
	"TABLESAMPLE":          struct{}{},
	"THEN":                 struct{}{},
	"TO":                   struct{}{},
	"TREAT":                struct{}{},
	"TRUE":                 struct{}{},
	"UNBOUNDED":            struct{}{},
	"UNION":                struct{}{},
	"UNNEST":               struct{}{},
	"USING":                struct{}{},
	"WHEN":                 struct{}{},
	"WHERE":                struct{}{},
	"WINDOW":               struct{}{},
	"WITH":                 struct{}{},
	"WITHIN":               struct{}{},
}
