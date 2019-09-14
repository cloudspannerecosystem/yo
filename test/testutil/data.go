package testutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	dbadmin "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/spannertest"
	"google.golang.org/api/option"
	dbadminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	"google.golang.org/grpc"
)

func DeleteAllData(ctx context.Context, client *spanner.Client) error {
	tables := []string{
		"CompositePrimaryKeys",
		"FullTypes",
		"MaxLengths",
		"snake_cases",
	}
	var muts []*spanner.Mutation
	for _, table := range tables {
		muts = append(muts, spanner.Delete(table, spanner.AllKeys()))
	}

	if _, err := client.Apply(ctx, muts, spanner.ApplyAtLeastOnce()); err != nil {
		return err
	}

	return nil
}

// SetupFakeSpanner runs fake spanner server and create clients for the server.
// Please make sure to call stop func to stop the server and the clients.
func SetupFakeSpanner(ctx context.Context, dbname string) (*spanner.Client, *dbadmin.DatabaseAdminClient, func(), error) {
	srv, err := spannertest.NewServer("localhost:0")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to start fake spanner server: %v", err)
	}

	dialCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, srv.Addr, grpc.WithInsecure())
	if err != nil {
		srv.Close()
		return nil, nil, nil, fmt.Errorf("dialing to fake spanner server failed: %v", err)
	}

	client, err := spanner.NewClient(ctx, dbname, option.WithGRPCConn(conn))
	if err != nil {
		srv.Close()
		conn.Close()
		return nil, nil, nil, fmt.Errorf("creating spanner client: %v", err)
	}
	adminClient, err := dbadmin.NewDatabaseAdminClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		srv.Close()
		conn.Close()
		return nil, nil, nil, fmt.Errorf("creating spanner admin client: %v", err)
	}

	stop := func() {
		srv.Close()
		conn.Close()
	}

	return client, adminClient, stop, nil
}

// ApplyDDL applies DDL statements and waits until finished.
func ApplyDDL(ctx context.Context, adminClient *dbadmin.DatabaseAdminClient, dbname string, statements ...string) error {
	op, err := adminClient.UpdateDatabaseDdl(ctx, &dbadminpb.UpdateDatabaseDdlRequest{
		Database:   dbname,
		Statements: statements,
	})
	if err != nil {
		return fmt.Errorf("apply DDL failed: %v", err)
	}
	return op.Wait(ctx)
}

func findProjectRootDir() string {
	dir, _ := os.Getwd()
	for {
		next := filepath.Dir(dir)
		if dir == next {
			panic("cannot find project root")
		}

		if filepath.Base(dir) == "yo" {
			break
		}

		dir = next
	}

	return dir
}

// ApplyTestScheme applies test schema in testdata.
func ApplyTestSchema(ctx context.Context, adminClient *dbadmin.DatabaseAdminClient, dbname string) error {
	dir := findProjectRootDir()

	// Open test schema
	file, err := os.Open(filepath.Join(dir, "./test/testdata/schema.sql"))
	if err != nil {
		return fmt.Errorf("scheme file cannot open: %v", err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read error: %v", err)
	}

	// Split scheme definition to DDL statements.
	// This assuemes there is no comments and each statement is separated by semi-colon
	var statements []string
	for _, s := range strings.Split(string(b), ";") {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		statements = append(statements, s)
	}

	// Apply DDL statements to create tables
	return ApplyDDL(ctx, adminClient, dbname, statements...)
}
