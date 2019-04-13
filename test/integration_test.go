package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"cloud.google.com/go/spanner"
	"github.com/google/go-cmp/cmp"
	"go.mercari.io/yo/test/testmodels/customtypes"
	models "go.mercari.io/yo/test/testmodels/default"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	spannerProjectName  = os.Getenv("SPANNER_PROJECT_NAME")
	spannerInstanceName = os.Getenv("SPANNER_INSTANCE_NAME")
	spannerDatabaseName = os.Getenv("SPANNER_DATABASE_NAME")
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

func DeleteAllData(ctx context.Context, client *spanner.Client) error {
	tables := []string{"CompositePrimaryKeys", "FullTypes", "MaxLengths", "snake_cases"}
	var muts []*spanner.Mutation
	for _, table := range tables {
		muts = append(muts, spanner.Delete(table, spanner.AllKeys()))
	}

	_, err := client.Apply(ctx, muts, spanner.ApplyAtLeastOnce())
	if err != nil {
		return err
	}

	return nil
}

func TestMain(m *testing.M) {
	os.Exit(func() int {
		ctx := context.Background()
		databaseName := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
			spannerProjectName, spannerInstanceName, spannerDatabaseName)
		spannerClient, err := spanner.NewClient(ctx, databaseName)
		if err != nil {
			panic(fmt.Sprintf("failed to create spanner client: %v", err))
		}
		defer spannerClient.Close()

		// preflight query to check database ready
		spanCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		iter := spannerClient.Single().Query(spanCtx, spanner.NewStatement("SELECT 1"))
		if _, err = iter.Next(); err != nil {
			panic(err)
		}
		iter.Stop()

		client = spannerClient

		if err := DeleteAllData(ctx, client); err != nil {
			panic(err)
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

	if _, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		var muts []*spanner.Mutation
		muts = append(muts, cpk.Insert(ctx))

		if err := tx.BufferWrite(muts); err != nil {
			return err
		}

		return nil
	}); err != nil {
		t.Fatalf("ReadWriteTransaction failed: %v", err)
	}

	got, err := models.FindCompositePrimaryKey(ctx, client.Single(), "x200", 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(cpk, got); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestDefaultCompositePrimaryKey_NotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := models.FindCompositePrimaryKey(ctx, client.Single(), "default", 100)
	if err == nil {
		t.Fatal("unexpected success")
	}

	testGRPCStatus(t, err, codes.NotFound)
	testNotFound(t, err, true)
	testTableName(t, err, "CompositePrimaryKeys")
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
			if _, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
				var muts []*spanner.Mutation
				muts = append(muts, tt.ft.Insert(ctx))

				if err := tx.BufferWrite(muts); err != nil {
					return err
				}

				return nil
			}); err != nil {
				t.Fatalf("ReadWriteTransaction failed: %v", err)
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

	if _, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		var muts []*spanner.Mutation
		muts = append(muts, cpk.Insert(ctx))

		if err := tx.BufferWrite(muts); err != nil {
			return err
		}

		return nil
	}); err != nil {
		t.Fatalf("ReadWriteTransaction failed: %v", err)
	}

	got, err := customtypes.FindCompositePrimaryKey(ctx, client.Single(), "x300", 300)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if diff := cmp.Diff(cpk, got); diff != "" {
		t.Errorf("(-got, +want)\n%s", diff)
	}
}

func TestCustomCompositePrimaryKey_NotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := customtypes.FindCompositePrimaryKey(ctx, client.Single(), "custom", 100)
	if err == nil {
		t.Fatal("unexpected success")
	}

	testGRPCStatus(t, err, codes.NotFound)
	testNotFound(t, err, true)
	testTableName(t, err, "CompositePrimaryKeys")
}
