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
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/spanner"
	dbadmin "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/google/go-cmp/cmp"
	"github.com/cloudspannerecosystem/yo/test/testmodels/customtypes"
	models "github.com/cloudspannerecosystem/yo/test/testmodels/default"
	"github.com/cloudspannerecosystem/yo/test/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	spannerProjectName  = os.Getenv("SPANNER_PROJECT_NAME")
	spannerInstanceName = os.Getenv("SPANNER_INSTANCE_NAME")
	spannerDatabaseName = os.Getenv("SPANNER_DATABASE_NAME")
)

var (
	client      *spanner.Client
	adminClient *dbadmin.DatabaseAdminClient
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

func TestMain(m *testing.M) {
	// explicitly call flag.Parse() to use testing.Short() in TestMain
	if !flag.Parsed() {
		flag.Parse()
	}

	os.Exit(func() int {
		ctx := context.Background()
		var databaseName string

		// If test.short is enabled, use fake spanner server for testing.
		// Otherwise use a real spanner server.
		if testing.Short() {
			databaseName = "projects/yo/instances/yo/databases/integration-test"
			cli, adminCli, stop, err := testutil.SetupFakeSpanner(ctx, databaseName)
			if err != nil {
				panic(err)
			}
			defer stop()

			client = cli
			adminClient = adminCli
		} else {
			databaseName = fmt.Sprintf("projects/%s/instances/%s/databases/%s",
				spannerProjectName, spannerInstanceName, spannerDatabaseName)

			spannerClient, err := spanner.NewClient(ctx, databaseName)
			if err != nil {
				panic(fmt.Sprintf("failed to create spanner client: %v", err))
			}
			defer spannerClient.Close()

			client = spannerClient
		}

		// preflight query to check database ready
		spanCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		iter := client.Single().Query(spanCtx, spanner.NewStatement("SELECT 1"))
		if _, err := iter.Next(); err != nil {
			panic(err)
		}
		iter.Stop()

		if testing.Short() {
			// Apply test scheme to create tables
			if err := testutil.ApplyTestSchema(ctx, adminClient, databaseName); err != nil {
				panic(err)
			}
		} else {
			// Delete exsitng data to make sure all tables are clean
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
				FTInt:                102,
				FTIntNull:            spanner.NullInt64{},
				FTFloat:              0.123,
				FTFloatNull:          spanner.NullFloat64{},
				FTDate:               date,
				FTDateNull:           spanner.NullDate{},
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
				FTInt:                103,
				FTIntNull:            spanner.NullInt64{},
				FTFloat:              0.123,
				FTFloatNull:          spanner.NullFloat64{},
				FTDate:               date,
				FTDateNull:           spanner.NullDate{},
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
