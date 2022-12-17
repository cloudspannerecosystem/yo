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
	"go.mercari.io/yo/v2/models"
)

var lengthRegexp = regexp.MustCompile(`\(([0-9]+|MAX)\)$`)

// SpanParseType parse a Spanner type into a Go type based on the column
// definition.
func parseSpannerType(dt string, ct string, nullable bool) (int, models.FieldType) {
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

	var fieldType models.FieldType
	switch dt {
	case "BOOL":
		fieldType = models.PlainFieldType{Type: "bool", CustomType: ct, NullValue: "false"}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullBool", CustomType: ct, NullValue: "NullBool{}"}
		}
	case "STRING":
		fieldType = models.PlainFieldType{Type: "string", CustomType: ct, NullValue: `""`}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullString", CustomType: ct, NullValue: "NullString{}"}
		}
	case "INT64":
		fieldType = models.PlainFieldType{Type: "int64", CustomType: ct, NullValue: `0`}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullInt64", CustomType: ct, NullValue: "NullInt64{}"}
		}
	case "FLOAT64":
		fieldType = models.PlainFieldType{Type: "float64", CustomType: ct, NullValue: `0.0`}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullFloat64", CustomType: ct, NullValue: "NullFloat64{}"}
		}
	case "BYTES":
		fieldType = models.PlainFieldType{Type: "[]byte", CustomType: ct, NullValue: models.GoNil}
	case "TIMESTAMP":
		fieldType = models.PlainFieldType{Pkg: models.TimePackage, Type: "Time", CustomType: ct, NullValue: `Time{}`}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullTime", CustomType: ct, NullValue: "NullTime{}"}
		}
	case "DATE":
		fieldType = models.PlainFieldType{Pkg: models.GoCivilPackage, Type: "Date", CustomType: ct, NullValue: "Date{}"}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullDate", CustomType: ct, NullValue: "NullDate{}"}
		}
	case "NUMERIC":
		fieldType = models.PlainFieldType{Pkg: models.MathBigPackage, Type: "Rat", CustomType: ct, NullValue: "Rat{}"}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullNumeric", CustomType: ct, NullValue: "NullNumeric{}"}
		}
	case "JSON":
		fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullJSON", CustomType: ct, NullValue: "NullJSON{Valid: true}"}
		if nullable {
			fieldType = models.PlainFieldType{Pkg: models.GoSpannerPackage, Type: "NullJSON", CustomType: ct, NullValue: "NullJSON"}
		}
	default:
		if strings.HasPrefix(dt, "ARRAY<") {
			eleDataType := strings.TrimSuffix(strings.TrimPrefix(dt, "ARRAY<"), ">")
			_, eleTyp := parseSpannerType(eleDataType, "", false)
			fieldType = models.ArrayFieldType{Nullable: nullable, CustomType: ct, Element: eleTyp}
			break
		}

		t := internal.SnakeToCamel(dt)
		fieldType = models.PlainFieldType{Type: t, CustomType: ct, NullValue: t + "{}"}
	}

	return length, fieldType
}
