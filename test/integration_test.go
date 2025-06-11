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
	"go.mercari.io/yo/loaders"
	"go.mercari.io/yo/test/testmodels/customtypes"
	models "go.mercari.io/yo/test/testmodels/default"
	"go.mercari.io/yo/test/testutil"
	"google.golang.org/api/option"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
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
			if err := testutil.SetupDatabase(ctx, spannerProjectName, spannerInstanceName, spannerDatabaseName); err != nil {
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

	cpk := &models.CompositePrimaryKey{
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
		got, err := models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(cpk, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByPrimaryKey", func(t *testing.T) {
		got, err := models.ReadCompositePrimaryKey(ctx, client.Single(), spanner.Key{"x200", 200})
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
		_, err := models.FindCompositePrimaryKey(ctx, client.Single(), "default", 100)
		if err == nil {
			t.Fatal("unexpected success")
		}

		testGRPCStatus(t, err, codes.NotFound)
		testNotFound(t, err, true)
		testTableName(t, err, "CompositePrimaryKeys")
	})

	t.Run("FindByError", func(t *testing.T) {
		got, err := models.FindCompositePrimaryKeysByError(ctx, client.Single(), cpk.Error)
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
		got, err := models.ReadCompositePrimaryKeysByError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &models.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByError2", func(t *testing.T) {
		got, err := models.ReadCompositePrimaryKeysByZError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &models.CompositePrimaryKey{
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
		got, err := models.ReadCompositePrimaryKeysByZYError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &models.CompositePrimaryKey{
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
		ft *models.FullType
	}{
		"case1": {
			ft: &models.FullType{
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
			ft: &models.FullType{
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
			ft: &models.FullType{
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

			got, err := models.FindFullType(ctx, client.Single(), tt.ft.PKey)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.ft, got); diff != "" {
				t.Errorf("(-got, +want)\n%s", diff)
			}
		})
	}

	t.Run("FindWithNonNull", func(t *testing.T) {
		fts, err := models.FindFullTypesByFTIntFTTimestampNull(ctx, client.Single(), 101, spanner.NullTime{
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
		fts, err := models.FindFullTypesByFTIntFTTimestampNull(ctx, client.Single(), 101, spanner.NullTime{
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

	cpk := &customtypes.CompositePrimaryKey{
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
		got, err := customtypes.FindCompositePrimaryKey(ctx, client.Single(), "x300", 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if diff := cmp.Diff(cpk, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("ReadByPrimaryKey", func(t *testing.T) {
		got, err := customtypes.ReadCompositePrimaryKey(ctx, client.Single(), spanner.Key{"x300", 300})
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
		_, err := customtypes.FindCompositePrimaryKey(ctx, client.Single(), "custom", 100)
		if err == nil {
			t.Fatal("unexpected success")
		}

		testGRPCStatus(t, err, codes.NotFound)
		testNotFound(t, err, true)
		testTableName(t, err, "CompositePrimaryKeys")
	})

	t.Run("FindByError", func(t *testing.T) {
		got, err := customtypes.FindCompositePrimaryKeysByError(ctx, client.Single(), cpk.Error)
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
		got, err := customtypes.ReadCompositePrimaryKeysByError(ctx, client.Single(), spanner.Key{cpk.Error})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expect the number of rows %v, but got %v", 1, len(got))
		}

		expected := &customtypes.CompositePrimaryKey{
			PKey1: cpk.PKey1,
			PKey2: cpk.PKey2,
			Error: cpk.Error,
		}
		if diff := cmp.Diff(expected, got[0]); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestAllowCommitTimestamp(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	v := &models.AllowCommitTimestamp{
		ID:        300,
		UpdatedAt: spanner.CommitTimestamp,
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{v.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	var insertTime time.Time
	t.Run("Insert", func(t *testing.T) {
		got, err := models.FindAllowCommitTimestamp(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.UpdatedAt.IsZero() {
			t.Errorf("got UpdatedAt.IsZero")
		}
		insertTime = got.UpdatedAt
	})

	t.Run("Update", func(t *testing.T) {
		gc := &models.AllowCommitTimestamp{
			ID:        300,
			UpdatedAt: spanner.CommitTimestamp,
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Update(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := models.FindAllowCommitTimestamp(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !got.UpdatedAt.After(insertTime) {
			t.Errorf("expected UpdatedAt (%v) to be after insertTime (%v)", got.UpdatedAt, insertTime)
		}
	})

	t.Run("InsertOrUpdate", func(t *testing.T) {
		gc := &models.AllowCommitTimestamp{
			ID:        300,
			UpdatedAt: spanner.CommitTimestamp,
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.InsertOrUpdate(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := models.FindAllowCommitTimestamp(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.UpdatedAt.IsZero() {
			t.Errorf("got UpdatedAt.IsZero")
		}
	})

	t.Run("IsAllowCommitTimestamp", func(t *testing.T) {
		gc := &models.AllowCommitTimestamp{
			ID:        300,
			UpdatedAt: spanner.CommitTimestamp,
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Update(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		cols, err := loaders.SpanTableColumns(client, "AllowCommitTimestamp")
		if err != nil {
			t.Fatalf("SpanTableColumns failed: %v", err)
		}

		for _, col := range cols {
			if col.ColumnName == "UpdatedAt" {
				if !col.IsAllowCommitTimestamp {
					t.Errorf("updated_at is not AllowCommitTimestamp")
				}
			}
		}
	})
}

func TestGeneratedColumn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gc := &models.GeneratedColumn{
		ID:        300,
		FirstName: "John",
		LastName:  "Doe",
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("Insert", func(t *testing.T) {
		got, err := models.FindGeneratedColumn(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &models.GeneratedColumn{
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
		gc := &models.GeneratedColumn{
			ID:        300,
			FirstName: "Jane",
			LastName:  "Doe",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Update(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := models.FindGeneratedColumn(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &models.GeneratedColumn{
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
		gc := &models.GeneratedColumn{
			ID:        300,
			FirstName: "Paul",
			LastName:  "Doe",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.InsertOrUpdate(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := models.FindGeneratedColumn(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &models.GeneratedColumn{
			ID:        300,
			FirstName: "Paul",
			LastName:  "Doe",
			FullName:  "Paul Doe",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})
}

func TestHiddenColumn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gc := &models.FullTextSearch{
		ID:      300,
		Content: "Hello",
	}

	if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Insert(ctx)}); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	t.Run("Insert", func(t *testing.T) {
		got, err := models.FindFullTextSearch(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &models.FullTextSearch{
			ID:      300,
			Content: "Hello",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("Update", func(t *testing.T) {
		gc := &models.FullTextSearch{
			ID:      300,
			Content: "Good Bye",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Update(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := models.FindFullTextSearch(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &models.FullTextSearch{
			ID:      300,
			Content: "Good Bye",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("InsertOrUpdate", func(t *testing.T) {
		gc := &models.FullTextSearch{
			ID:      300,
			Content: "Good Morning",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.InsertOrUpdate(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		got, err := models.FindFullTextSearch(ctx, client.Single(), 300)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := &models.FullTextSearch{
			ID:      300,
			Content: "Good Morning",
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("(-got, +want)\n%s", diff)
		}
	})

	t.Run("IsHidden", func(t *testing.T) {
		gc := &models.FullTextSearch{
			ID:      300,
			Content: "Hello",
		}

		if _, err := client.Apply(ctx, []*spanner.Mutation{gc.Update(ctx)}); err != nil {
			t.Fatalf("Apply failed: %v", err)
		}

		cols, err := loaders.SpanTableColumns(client, "FullTextSearch")
		if err != nil {
			t.Fatalf("SpanTableColumns failed: %v", err)
		}

		for _, col := range cols {
			if col.ColumnName == "Content_Tokens" {
				if !col.IsHidden {
					t.Errorf("Content_Tokens is not Hidden")
				}
			}
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
		_, err = models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
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
		_, err = models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
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
