package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
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

		return m.Run()
	}())
}

func TestDefaultCompositePrimaryKey(t *testing.T) {
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

func TestCustomCompositePrimaryKey(t *testing.T) {
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
