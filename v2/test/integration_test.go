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
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/spanner"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go/v2/apierror"
	default_models "go.mercari.io/yo/v2/test/testmodels/default"
	legacy_models "go.mercari.io/yo/v2/test/testmodels/legacy_default"
	"go.mercari.io/yo/v2/test/testutil"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
)

var (
	spannerProjectName  = os.Getenv("SPANNER_PROJECT_NAME")
	spannerInstanceName = os.Getenv("SPANNER_INSTANCE_NAME")
	spannerDatabaseName = os.Getenv("SPANNER_DATABASE_NAME")
	spannerEmulatorHost = os.Getenv("SPANNER_EMULATOR_HOST")
)

var (
	client *spanner.Client
)

func testNotFound(t *testing.T, err error, b bool) {
	t.Helper()

	nf, ok := err.(interface {
		NotFound() bool
	})
	if !ok {
		t.Fatal("err must implement NotFound() bool")
	}

	if want, got := b, nf.NotFound(); want != got {
		t.Fatalf("expect NotFound() to %v, but got %v", want, got)
	}
}

func testGRPCStatus(t *testing.T, err error, c codes.Code) {
	t.Helper()

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("err must be grpc error")
	}

	if want, got := c, st.Code(); want != got {
		t.Fatalf("expect grpc code %q, but got %q", want, got)
	}
}

func testTableName(t *testing.T, err error, name string) {
	t.Helper()

	tn, ok := err.(interface {
		DBTableName() string
	})
	if !ok {
		t.Fatal("err must implement DBTableName() string")
	}

	if want, got := name, tn.DBTableName(); want != got {
		t.Fatalf("expect DBTableName() to %q, but got %q", want, got)
	}
}

const sessionResourceType = "type.googleapis.com/google.spanner.v1.Session"

func newSessionNotFoundError(name string) error {
	s := status.Newf(codes.NotFound, "Session not found: Session with id %s not found", name)
	s, _ = s.WithDetails(&errdetails.ResourceInfo{ResourceType: sessionResourceType, ResourceName: name})
	return s.Err()
}

func newAbortedWithRetryInfo() error {
	s := status.New(codes.Aborted, "")
	s, err := s.WithDetails(&errdetails.RetryInfo{
		RetryDelay: durationpb.New(100 * time.Millisecond),
	})
	if err != nil {
		panic(fmt.Sprintf("with details failed: %v", err))
	}
	return s.Err()
}

func TestMain(m *testing.M) {
	// explicitly call flag.Parse() to use testing.Short() in TestMain
	if !flag.Parsed() {
		flag.Parse()
	}

	os.Exit(func() int {
		ctx := context.Background()

		if !testing.Short() {
			if err := testutil.SetupDatabase(ctx, spannerProjectName, spannerInstanceName, spannerDatabaseName, ""); err != nil {
				panic(err)
			}
		}

		spanCli, err := testutil.TestClient(ctx, spannerProjectName, spannerInstanceName, spannerDatabaseName)
		if err != nil {
			panic(err)
		}

		client = spanCli

		if testing.Short() {
			if err := testutil.DeleteAllData(ctx, client); err != nil {
				panic(err)
			}
		}

		return m.Run()
	}())
}

func TestDefaultCompositePrimaryKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	cpk := &default_models.CompositePrimaryKey{
		ID:    200,
		PKey1: "x200",
		PKey2: 200,
		Error: 200,
		X:     "x200",
		Y:     "y200",
		Z:     "z200",
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{cpk.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("FindByPrimaryKey", func(t *testing.T) {
		got, err := default_models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(cpk, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByPrimaryKey", func(t *testing.T) {
		got, err := default_models.ReadCompositePrimaryKey(ctx, client.Single(), spanner.Key{"x200", 200})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		if diff := cmp.Diff(cpk, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := default_models.FindCompositePrimaryKey(ctx, client.Single(), "default", 100)
		if err == nil {
			t.Fatal("unexpected success")
		}

		testGRPCStatus(t, err, codes.NotFound)
		testNotFound(t, err, true)
		testTableName(t, err, "CompositePrimaryKeys")
	})

	t.Run("FindByError", func(t *testing.T) {
		got, err := default_models.FindCompositePrimaryKeysByCompositePrimaryKeysByError(ctx, client.Single(), cpk.Error)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		if diff := cmp.Diff(cpk, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError", func(t *testing.T) {
		got, err := default_models.ReadCompositePrimaryKeysByCompositePrimaryKeysByError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &default_models.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError2", func(t *testing.T) {
		got, err := default_models.ReadCompositePrimaryKeysByCompositePrimaryKeysByError2(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &default_models.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
			Z:     cpk.Z,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError3", func(t *testing.T) {
		got, err := default_models.ReadCompositePrimaryKeysByCompositePrimaryKeysByError3(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &default_models.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
			Y:     cpk.Y,
			Z:     cpk.Z,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestDefaultFullType(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	now := time.Now()
	date := civil.DateOf(now)
	tomorrow := now.AddDate(0, 0, 1)
	nextdate := civil.DateOf(tomorrow)
	json := spanner.NullJSON{
		Valid: true,
		Value: `{"a": "b"}`,
	}
	jsonNull := spanner.NullJSON{}

	table := map[string]struct {
		ft *default_models.FullType
	}{
		"case1": {
			ft: &default_models.FullType{
				PKey:     "pkey1",
				FTString: "xxx1",
				FTStringNull: spanner.NullString{
					StringVal: "xxx1",
					Valid:     true,
				},
				FTBool: true,
				FTBoolNull: spanner.NullBool{
					Bool:  true,
					Valid: true,
				},
				FTBytes:     []byte("xxx1"),
				FTBytesNull: []byte("xxx1"),
				FTTimestamp: now,
				FTTimestampNull: spanner.NullTime{
					Time:  now,
					Valid: true,
				},
				FTInt: 101,
				FTIntNull: spanner.NullInt64{
					Int64: 101,
					Valid: true,
				},
				FTFloat: 0.123,
				FTFloatNull: spanner.NullFloat64{
					Float64: 0.123,
					Valid:   true,
				},
				FTDate: date,
				FTDateNull: spanner.NullDate{
					Date:  date,
					Valid: true,
				},
				FTJSON:               json,
				FTJSONNull:           json,
				FTArrayStringNull:    []string{"xxx1", "yyy1"},
				FTArrayString:        []string{"xxx1", "yyy1"},
				FTArrayBoolNull:      []bool{true, false},
				FTArrayBool:          []bool{true, false},
				FTArrayBytesNull:     [][]byte{[]byte("xxx1"), []byte("yyy1")},
				FTArrayBytes:         [][]byte{[]byte("xxx1"), []byte("yyy1")},
				FTArrayTimestampNull: []time.Time{now, tomorrow},
				FTArrayTimestamp:     []time.Time{now, tomorrow},
				FTArrayIntNull:       []int64{100, 200},
				FTArrayInt:           []int64{100, 200},
				FTArrayFloatNull:     []float64{0.111, 0.222},
				FTArrayFloat:         []float64{0.111, 0.222},
				FTArrayDateNull:      []civil.Date{date, nextdate},
				FTArrayDate:          []civil.Date{date, nextdate},
				FTArrayJSONNull:      []spanner.NullJSON{json, jsonNull},
				FTArrayJSON:          []spanner.NullJSON{json, jsonNull},
			},
		},
		"case2": {
			ft: &default_models.FullType{
				PKey:                 "pkey2",
				FTString:             "xxx2",
				FTStringNull:         spanner.NullString{},
				FTBool:               true,
				FTBoolNull:           spanner.NullBool{},
				FTBytes:              []byte("xxx2"),
				FTBytesNull:          nil,
				FTTimestamp:          now,
				FTTimestampNull:      spanner.NullTime{},
				FTInt:                101,
				FTIntNull:            spanner.NullInt64{},
				FTFloat:              0.123,
				FTFloatNull:          spanner.NullFloat64{},
				FTDate:               date,
				FTDateNull:           spanner.NullDate{},
				FTArrayStringNull:    []string{"xxx2", "yyy2"},
				FTArrayString:        []string{"xxx2", "yyy2"},
				FTJSON:               json,
				FTJSONNull:           jsonNull,
				FTArrayBoolNull:      nil,
				FTArrayBool:          []bool{true, false},
				FTArrayBytesNull:     nil,
				FTArrayBytes:         [][]byte{[]byte("xxx2"), []byte("yyy2")},
				FTArrayTimestampNull: nil,
				FTArrayTimestamp:     []time.Time{now, tomorrow},
				FTArrayIntNull:       nil,
				FTArrayInt:           []int64{100, 200},
				FTArrayFloatNull:     nil,
				FTArrayFloat:         []float64{0.111, 0.222},
				FTArrayDateNull:      nil,
				FTArrayDate:          []civil.Date{date, nextdate},
				FTArrayJSONNull:      nil,
				FTArrayJSON:          []spanner.NullJSON{json, jsonNull},
			},
		},
		"case3": {
			ft: &default_models.FullType{
				PKey:                 "pkey3",
				FTString:             "xxx3",
				FTStringNull:         spanner.NullString{},
				FTBool:               true,
				FTBoolNull:           spanner.NullBool{},
				FTBytes:              []byte("xxx3"),
				FTBytesNull:          nil,
				FTTimestamp:          now,
				FTTimestampNull:      spanner.NullTime{},
				FTInt:                101,
				FTIntNull:            spanner.NullInt64{},
				FTFloat:              0.123,
				FTFloatNull:          spanner.NullFloat64{},
				FTDate:               date,
				FTDateNull:           spanner.NullDate{},
				FTJSON:               json,
				FTJSONNull:           jsonNull,
				FTArrayStringNull:    []string{},
				FTArrayString:        []string{},
				FTArrayBoolNull:      []bool{},
				FTArrayBool:          []bool{},
				FTArrayBytesNull:     [][]byte{},
				FTArrayBytes:         [][]byte{},
				FTArrayTimestampNull: []time.Time{},
				FTArrayTimestamp:     []time.Time{},
				FTArrayIntNull:       []int64{},
				FTArrayInt:           []int64{},
				FTArrayFloatNull:     []float64{},
				FTArrayFloat:         []float64{},
				FTArrayDateNull:      []civil.Date{},
				FTArrayDate:          []civil.Date{},
				FTArrayJSONNull:      []spanner.NullJSON{},
				FTArrayJSON:          []spanner.NullJSON{},
			},
		},
	}

	for name, tt := range table {
		t.Run(name, func(t *testing.T) {
			if _, err := client.Apply(ctx, []*spanner.Mutation{tt.ft.Insert(ctx)}); err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			got, err := default_models.FindFullType(ctx, client.Single(), tt.ft.PKey)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.ft, got); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}

	t.Run("FindWithNonNull", func(t *testing.T) {
		fts, err := default_models.FindFullTypesByFullTypesByInTimestampNull(ctx, client.Single(), 101, spanner.NullTime{
			Time:  now,
			Valid: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var pkeys []string
		for i := range fts {
			pkeys = append(pkeys, fts[i].PKey)
		}

		expected := []string{"pkey1"}
		if diff := cmp.Diff(expected, pkeys); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("FindWithNull", func(t *testing.T) {
		fts, err := default_models.FindFullTypesByFullTypesByInTimestampNull(ctx, client.Single(), 101, spanner.NullTime{
			Valid: false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var pkeys []string
		for i := range fts {
			pkeys = append(pkeys, fts[i].PKey)
		}

		expected := []string{"pkey3", "pkey2"}
		if diff := cmp.Diff(expected, pkeys); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestLegacyDefaultCompositePrimaryKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	cpk := &legacy_models.CompositePrimaryKey{
		ID:    200,
		PKey1: "x200",
		PKey2: 200,
		Error: 200,
		X:     "x200",
		Y:     "y200",
		Z:     "z200",
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{cpk.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("FindByPrimaryKey", func(t *testing.T) {
		got, err := legacy_models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(cpk, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByPrimaryKey", func(t *testing.T) {
		got, err := legacy_models.ReadCompositePrimaryKey(ctx, client.Single(), spanner.Key{"x200", 200})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		if diff := cmp.Diff(cpk, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := legacy_models.FindCompositePrimaryKey(ctx, client.Single(), "default", 100)
		if err == nil {
			t.Fatal("unexpected success")
		}

		testGRPCStatus(t, err, codes.NotFound)
		testNotFound(t, err, true)
		testTableName(t, err, "CompositePrimaryKeys")
	})

	t.Run("FindByError", func(t *testing.T) {
		got, err := legacy_models.FindCompositePrimaryKeysByError(ctx, client.Single(), cpk.Error)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		if diff := cmp.Diff(cpk, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError", func(t *testing.T) {
		got, err := legacy_models.ReadCompositePrimaryKeysByError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &legacy_models.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError2", func(t *testing.T) {
		got, err := legacy_models.ReadCompositePrimaryKeysByZError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &legacy_models.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
			Z:     cpk.Z,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError3", func(t *testing.T) {
		got, err := legacy_models.ReadCompositePrimaryKeysByZYError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &legacy_models.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
			Y:     cpk.Y,
			Z:     cpk.Z,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestLegacyDefaultFullType(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	now := time.Now()
	date := civil.DateOf(now)
	tomorrow := now.AddDate(0, 0, 1)
	nextdate := civil.DateOf(tomorrow)
	json := spanner.NullJSON{
		Valid: true,
		Value: `{"a": "b"}`,
	}
	jsonNull := spanner.NullJSON{}

	table := map[string]struct {
		ft *legacy_models.FullType
	}{
		"case1": {
			ft: &legacy_models.FullType{
				PKey:     "pkey1",
				FTString: "xxx1",
				FTStringNull: spanner.NullString{
					StringVal: "xxx1",
					Valid:     true,
				},
				FTBool: true,
				FTBoolNull: spanner.NullBool{
					Bool:  true,
					Valid: true,
				},
				FTBytes:     []byte("xxx1"),
				FTBytesNull: []byte("xxx1"),
				FTTimestamp: now,
				FTTimestampNull: spanner.NullTime{
					Time:  now,
					Valid: true,
				},
				FTInt: 101,
				FTIntNull: spanner.NullInt64{
					Int64: 101,
					Valid: true,
				},
				FTFloat: 0.123,
				FTFloatNull: spanner.NullFloat64{
					Float64: 0.123,
					Valid:   true,
				},
				FTDate: date,
				FTDateNull: spanner.NullDate{
					Date:  date,
					Valid: true,
				},
				FTJSON:               json,
				FTJSONNull:           json,
				FTArrayStringNull:    []string{"xxx1", "yyy1"},
				FTArrayString:        []string{"xxx1", "yyy1"},
				FTArrayBoolNull:      []bool{true, false},
				FTArrayBool:          []bool{true, false},
				FTArrayBytesNull:     [][]byte{[]byte("xxx1"), []byte("yyy1")},
				FTArrayBytes:         [][]byte{[]byte("xxx1"), []byte("yyy1")},
				FTArrayTimestampNull: []time.Time{now, tomorrow},
				FTArrayTimestamp:     []time.Time{now, tomorrow},
				FTArrayIntNull:       []int64{100, 200},
				FTArrayInt:           []int64{100, 200},
				FTArrayFloatNull:     []float64{0.111, 0.222},
				FTArrayFloat:         []float64{0.111, 0.222},
				FTArrayDateNull:      []civil.Date{date, nextdate},
				FTArrayDate:          []civil.Date{date, nextdate},
				FTArrayJSONNull:      []spanner.NullJSON{json, jsonNull},
				FTArrayJSON:          []spanner.NullJSON{json, jsonNull},
			},
		},
		"case2": {
			ft: &legacy_models.FullType{
				PKey:                 "pkey2",
				FTString:             "xxx2",
				FTStringNull:         spanner.NullString{},
				FTBool:               true,
				FTBoolNull:           spanner.NullBool{},
				FTBytes:              []byte("xxx2"),
				FTBytesNull:          nil,
				FTTimestamp:          now,
				FTTimestampNull:      spanner.NullTime{},
				FTInt:                101,
				FTIntNull:            spanner.NullInt64{},
				FTFloat:              0.123,
				FTFloatNull:          spanner.NullFloat64{},
				FTDate:               date,
				FTDateNull:           spanner.NullDate{},
				FTJSON:               json,
				FTJSONNull:           jsonNull,
				FTArrayStringNull:    []string{"xxx2", "yyy2"},
				FTArrayString:        []string{"xxx2", "yyy2"},
				FTArrayBoolNull:      nil,
				FTArrayBool:          []bool{true, false},
				FTArrayBytesNull:     nil,
				FTArrayBytes:         [][]byte{[]byte("xxx2"), []byte("yyy2")},
				FTArrayTimestampNull: nil,
				FTArrayTimestamp:     []time.Time{now, tomorrow},
				FTArrayIntNull:       nil,
				FTArrayInt:           []int64{100, 200},
				FTArrayFloatNull:     nil,
				FTArrayFloat:         []float64{0.111, 0.222},
				FTArrayDateNull:      nil,
				FTArrayDate:          []civil.Date{date, nextdate},
				FTArrayJSONNull:      nil,
				FTArrayJSON:          []spanner.NullJSON{json, jsonNull},
			},
		},
		"case3": {
			ft: &legacy_models.FullType{
				PKey:                 "pkey3",
				FTString:             "xxx3",
				FTStringNull:         spanner.NullString{},
				FTBool:               true,
				FTBoolNull:           spanner.NullBool{},
				FTBytes:              []byte("xxx3"),
				FTBytesNull:          nil,
				FTTimestamp:          now,
				FTTimestampNull:      spanner.NullTime{},
				FTInt:                101,
				FTIntNull:            spanner.NullInt64{},
				FTFloat:              0.123,
				FTFloatNull:          spanner.NullFloat64{},
				FTDate:               date,
				FTDateNull:           spanner.NullDate{},
				FTJSON:               json,
				FTJSONNull:           jsonNull,
				FTArrayStringNull:    []string{},
				FTArrayString:        []string{},
				FTArrayBoolNull:      []bool{},
				FTArrayBool:          []bool{},
				FTArrayBytesNull:     [][]byte{},
				FTArrayBytes:         [][]byte{},
				FTArrayTimestampNull: []time.Time{},
				FTArrayTimestamp:     []time.Time{},
				FTArrayIntNull:       []int64{},
				FTArrayInt:           []int64{},
				FTArrayFloatNull:     []float64{},
				FTArrayFloat:         []float64{},
				FTArrayDateNull:      []civil.Date{},
				FTArrayDate:          []civil.Date{},
				FTArrayJSONNull:      []spanner.NullJSON{},
				FTArrayJSON:          []spanner.NullJSON{},
			},
		},
	}

	for name, tt := range table {
		t.Run(name, func(t *testing.T) {
			if _, err := client.Apply(ctx, []*spanner.Mutation{tt.ft.Insert(ctx)}); err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			got, err := legacy_models.FindFullType(ctx, client.Single(), tt.ft.PKey)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.ft, got); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}

	t.Run("FindWithNonNull", func(t *testing.T) {
		fts, err := legacy_models.FindFullTypesByFTIntFTTimestampNull(ctx, client.Single(), 101, spanner.NullTime{
			Time:  now,
			Valid: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var pkeys []string
		for i := range fts {
			pkeys = append(pkeys, fts[i].PKey)
		}

		expected := []string{"pkey1"}
		if diff := cmp.Diff(expected, pkeys); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("FindWithNull", func(t *testing.T) {
		fts, err := legacy_models.FindFullTypesByFTIntFTTimestampNull(ctx, client.Single(), 101, spanner.NullTime{
			Valid: false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var pkeys []string
		for i := range fts {
			pkeys = append(pkeys, fts[i].PKey)
		}

		expected := []string{"pkey3", "pkey2"}
		if diff := cmp.Diff(expected, pkeys); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestCustomCompositePrimaryKey(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	cpk := &default_models.CustomCompositePrimaryKey{
		ID:    300,
		PKey1: "x300",
		PKey2: 300,
		Error: 3,
		X:     "x300",
		Y:     "y300",
		Z:     "z300",
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{cpk.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("FindByPrimaryKey", func(t *testing.T) {
		got, err := default_models.FindCustomCompositePrimaryKey(ctx, client.Single(), "x300", 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(cpk, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByPrimaryKey", func(t *testing.T) {
		got, err := default_models.ReadCustomCompositePrimaryKey(ctx, client.Single(), spanner.Key{"x300", 300})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		if diff := cmp.Diff(cpk, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := default_models.FindCustomCompositePrimaryKey(ctx, client.Single(), "custom", 100)
		if err == nil {
			t.Fatal("unexpected success")
		}

		testGRPCStatus(t, err, codes.NotFound)
		testNotFound(t, err, true)
		testTableName(t, err, "CustomCompositePrimaryKeys")
	})

	t.Run("FindByError", func(t *testing.T) {
		got, err := default_models.FindCustomCompositePrimaryKeysByCustomCompositePrimaryKeysByError(ctx, client.Single(), cpk.Error)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		if diff := cmp.Diff(cpk, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError", func(t *testing.T) {
		got, err := default_models.ReadCustomCompositePrimaryKeysByCustomCompositePrimaryKeysByError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &default_models.CustomCompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestCustomPrimitiveTypes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	cpk := &default_models.CustomPrimitiveType{
		PKey:              "pkey1",
		FTInt64:           1,
		FTInt64null:       1,
		FTInt32:           1,
		FTInt32null:       1,
		FTInt16:           1,
		FTInt16null:       1,
		FTInt8:            1,
		FTInt8null:        1,
		FTUInt64:          1,
		FTUInt64null:      1,
		FTUInt32:          1,
		FTUInt32null:      1,
		FTUInt16:          1,
		FTUInt16null:      1,
		FTUInt8:           1,
		FTUInt8null:       1,
		FTArrayInt64:      []int64{1, 2},
		FTArrayInt64null:  []int64{1, 2},
		FTArrayInt32:      []int64{1, 2},
		FTArrayInt32null:  []int64{1, 2},
		FTArrayInt16:      []int64{1, 2},
		FTArrayInt16null:  []int64{1, 2},
		FTArrayInt8:       []int64{1, 2},
		FTArrayInt8null:   []int64{1, 2},
		FTArrayUINt64:     []int64{1, 2},
		FTArrayUINt64null: []int64{1, 2},
		FTArrayUINt32:     []int64{1, 2},
		FTArrayUINt32null: []int64{1, 2},
		FTArrayUINt16:     []int64{1, 2},
		FTArrayUINt16null: []int64{1, 2},
		FTArrayUINt8:      []int64{1, 2},
		FTArrayUINt8null:  []int64{1, 2},
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{cpk.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("FindByPrimaryKey", func(t *testing.T) {
		got, err := default_models.FindCustomPrimitiveType(ctx, client.Single(), "pkey1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(cpk, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestGeneratedColumn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	gc := &default_models.GeneratedColumn{
		ID:        300,
		FirstName: "John",
		LastName:  "Doe",
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("Insert", func(t *testing.T) {
		got, err := default_models.FindGeneratedColumn(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &default_models.GeneratedColumn{
			ID:        300,
			FirstName: "John",
			LastName:  "Doe",
			FullName:  "John Doe",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("Update", func(t *testing.T) {
		gc := &default_models.GeneratedColumn{
			ID:        300,
			FirstName: "Jane",
			LastName:  "Doe",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Update(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := default_models.FindGeneratedColumn(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &default_models.GeneratedColumn{
			ID:        300,
			FirstName: "Jane",
			LastName:  "Doe",
			FullName:  "Jane Doe",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("InsertOrUpdate", func(t *testing.T) {
		gc := &default_models.GeneratedColumn{
			ID:        300,
			FirstName: "Paul",
			LastName:  "Doe",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.InsertOrUpdate(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := default_models.FindGeneratedColumn(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &default_models.GeneratedColumn{
			ID:        300,
			FirstName: "Paul",
			LastName:  "Doe",
			FullName:  "Paul Doe",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("Replace", func(t *testing.T) {
		gc := &default_models.GeneratedColumn{
			ID:        300,
			FirstName: "George",
			LastName:  "Doe",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Replace(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := default_models.FindGeneratedColumn(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &default_models.GeneratedColumn{
			ID:        300,
			FirstName: "George",
			LastName:  "Doe",
			FullName:  "George Doe",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestSessionNotFound(t *testing.T) {
	dbName := testutil.DatabaseName(spannerProjectName, spannerInstanceName, spannerDatabaseName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, spannerEmulatorHost, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				return invoker(ctx, method, req, reply, cc, opts...)
			},
		),
		grpc.WithStreamInterceptor(
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				if method == "/google.spanner.v1.Spanner/StreamingRead" {
					return nil, newSessionNotFoundError("xxx")
				}
				return streamer(ctx, desc, cc, method, opts...)
			},
		),
	)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	// Prepare spanner client
	client, err := spanner.NewClient(ctx, dbName, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("failed to connect fake spanner server: %v", err)
	}

	t.Run("ConvertToSpannerError", func(t *testing.T) {
		_, err = default_models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
		var ae *apierror.APIError
		if !errors.As(err, &ae) {
			t.Fatalf("the error returned by yo can be apierror.APIError: %T", err)
		}

		st := status.Convert(ae)
		ri := extractResourceInfo(st)

		expectedResourceInfo := &errdetails.ResourceInfo{
			ResourceType: "type.googleapis.com/google.spanner.v1.Session",
			ResourceName: "xxx",
		}

		if diff := cmp.Diff(expectedResourceInfo, ri, protocmp.Transform()); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ConvertToStatus", func(t *testing.T) {
		_, err = default_models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
		st := status.Convert(err)
		ri := extractResourceInfo(st)

		expectedResourceInfo := &errdetails.ResourceInfo{
			ResourceType: "type.googleapis.com/google.spanner.v1.Session",
			ResourceName: "xxx",
		}

		if diff := cmp.Diff(expectedResourceInfo, ri, protocmp.Transform()); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestAborted(t *testing.T) {
	dbName := testutil.DatabaseName(spannerProjectName, spannerInstanceName, spannerDatabaseName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var retried bool

	conn, err := grpc.DialContext(ctx, spannerEmulatorHost, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				if method == "/google.spanner.v1.Spanner/Commit" {
					if !retried {
						retried = true
						return newAbortedWithRetryInfo()
					}
				}
				return invoker(ctx, method, req, reply, cc, opts...)
			},
		),
		grpc.WithStreamInterceptor(
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				return streamer(ctx, desc, cc, method, opts...)
			},
		),
	)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	// Prepare spanner client
	client, err := spanner.NewClient(ctx, dbName, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("failed to connect fake spanner server: %v", err)
	}

	t.Run("OnCommit", func(t *testing.T) {
		_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
			return nil
		})
		if err != nil {
			t.Fatalf("transaction failed: %v", err)
		}

		if !retried {
			t.Fatalf("aborted on Commit should be retried")
		}
	})
}

func extractResourceInfo(st *status.Status) *errdetails.ResourceInfo {
	for _, detail := range st.Details() {
		if ri, ok := detail.(*errdetails.ResourceInfo); ok {
			return ri
		}
	}
	return nil
}

func TestInflectionzz(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := testutil.DeleteAllData(ctx, client); err != nil {
		t.Fatalf("failed to clear data: %v", err)
	}

	cpk := &default_models.Inflection{
		X: "x",
		Y: "y",
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{cpk.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("FindByPrimaryKey", func(t *testing.T) {
		got, err := default_models.FindInflection(ctx, client.Single(), "x")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(cpk, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByPrimaryKey", func(t *testing.T) {
		got, err := default_models.ReadInflection(ctx, client.Single(), spanner.Key{"x"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		if diff := cmp.Diff(cpk, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}
