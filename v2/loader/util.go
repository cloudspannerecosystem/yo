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
	"regexp"
	"strconv"
	"strings"

	"go.mercari.io/yo/v2/internal"
)

var lengthRegexp = regexp.MustCompile(`\(([0-9]+|MAX)\)$`)

// SpanParseType parse a Spanner type into a Go type based on the column
// definition.
func parseSpannerType(dt string, nullable bool) (int, string, string) {
	nilVal := "nil"
	length := -1

	// separate type and length from dt with length such as STRING(32) or BYTES(256)
	m := lengthRegexp.FindStringSubmatchIndex(dt)
	if m != nil {
		lengthStr := dt[m[2]:m[3]]
		if lengthStr == "MAX" {
			length = -1
		} else {
			l, err := strconv.Atoi(lengthStr)
			if err != nil {
				panic("could not convert precision")
			}
			length = l
		}

		// trim length from dt
		dt = dt[:m[0]] + dt[m[1]:]
	}

	var typ string
	switch dt {
	case "BOOL":
		nilVal = "false"
		typ = "bool"
		if nullable {
			nilVal = "spanner.NullBool{}"
			typ = "spanner.NullBool"
		}

	case "STRING":
		nilVal = `""`
		typ = "string"
		if nullable {
			nilVal = "spanner.NullString{}"
			typ = "spanner.NullString"
		}

	case "INT64":
		nilVal = "0"
		typ = "int64"
		if nullable {
			nilVal = "spanner.NullInt64{}"
			typ = "spanner.NullInt64"
		}

	case "FLOAT64":
		nilVal = "0.0"
		typ = "float64"
		if nullable {
			nilVal = "spanner.NullFloat64{}"
			typ = "spanner.NullFloat64"
		}

	case "BYTES":
		typ = "[]byte"

	case "TIMESTAMP":
		nilVal = "time.Time{}"
		typ = "time.Time"
		if nullable {
			nilVal = "spanner.NullTime{}"
			typ = "spanner.NullTime"
		}

	case "DATE":
		nilVal = "civil.Date{}"
		typ = "civil.Date"
		if nullable {
			nilVal = "spanner.NullDate{}"
			typ = "spanner.NullDate"
		}

	default:
		if strings.HasPrefix(dt, "ARRAY<") {
			eleDataType := strings.TrimSuffix(strings.TrimPrefix(dt, "ARRAY<"), ">")
			_, _, eleTyp := parseSpannerType(eleDataType, false)
			typ, nilVal = "[]"+eleTyp, "nil"
			if !nullable {
				nilVal = typ + "{}"
			}
			break
		}

		typ = internal.SnakeToCamel(dt)
		nilVal = typ + "{}"
	}

	return length, nilVal, typ
}

func validateCustomType(dataType string, customType string) bool {
	// No custom type validation now
	return true
}
