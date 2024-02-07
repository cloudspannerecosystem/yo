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

package testutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/spanner"
	dbadmin "cloud.google.com/go/spanner/admin/database/apiv1"
	insadmin "cloud.google.com/go/spanner/admin/instance/apiv1"
	parser "github.com/cloudspannerecosystem/memefish"
	mtoken "github.com/cloudspannerecosystem/memefish/token"
	dbadminpb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	"google.golang.org/grpc/codes"
)

func DatabaseName(projectName, instanceName, dbName string) string {
	return fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		projectName, instanceName, dbName)
}

func TestClient(ctx context.Context, projectName, instanceName, dbName string) (*spanner.Client, error) {
	fullDBName := DatabaseName(projectName, instanceName, dbName)

	client, err := spanner.NewClient(ctx, fullDBName)
	if err != nil {
		return nil, fmt.Errorf("failed to create spanner client: %v", err)
	}

	// preflight query to check database ready
	spanCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	iter := client.Single().Query(spanCtx, spanner.NewStatement("SELECT 1"))
	defer iter.Stop()
	if _, err := iter.Next(); err != nil {
		return nil, err
	}

	return client, nil
}

func DeleteAllData(ctx context.Context, client *spanner.Client) error {
	tables := []string{
		"CompositePrimaryKeys",
		"FullTypes",
		"MaxLengths",
		"snake_cases",
		"GeneratedColumns",
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

func SetupDatabase(ctx context.Context, projectName, instanceName, dbName string) error {
	if v := os.Getenv("SPANNER_EMULATOR_HOST"); v == "" {
		return fmt.Errorf("test must use spanner emulator")
	}

	fullDBName := DatabaseName(projectName, instanceName, dbName)

	insAdminCli, err := insadmin.NewInstanceAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instance admin client: %v", err)
	}
	defer insAdminCli.Close()

	dbAdminCli, err := dbadmin.NewDatabaseAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create database admin client: %v", err)
	}
	defer dbAdminCli.Close()

	found, err := CheckDatabaseExist(ctx, dbAdminCli, fullDBName)
	if err != nil {
		return err
	}

	if found {
		if err := DropDatabase(ctx, dbAdminCli, fullDBName); err != nil {
			return err
		}
	} else {
		if err := CreateInstance(ctx, insAdminCli, projectName, instanceName); err != nil {
			return err
		}
	}

	if err := CreateDatabase(ctx, dbAdminCli, projectName, instanceName, dbName); err != nil {
		return err
	}

	if err := ApplyTestSchema(ctx, dbAdminCli, fullDBName); err != nil {
		return err
	}

	return nil
}

func CheckDatabaseExist(ctx context.Context, dbAdminCli *dbadmin.DatabaseAdminClient, fullDBName string) (bool, error) {
	_, err := dbAdminCli.GetDatabase(ctx, &dbadminpb.GetDatabaseRequest{
		Name: fullDBName,
	})
	if err != nil {
		if code := spanner.ErrCode(err); code == codes.NotFound {
			return false, nil
		}

		return false, fmt.Errorf("failed to get database: %v", err)
	}

	return true, nil
}

func DropDatabase(ctx context.Context, dbAdminCli *dbadmin.DatabaseAdminClient, fullDBName string) error {
	err := dbAdminCli.DropDatabase(ctx, &dbadminpb.DropDatabaseRequest{
		Database: fullDBName,
	})
	if err != nil {
		return fmt.Errorf("failed to drop database: %v", err)
	}

	return nil
}

func CreateInstance(ctx context.Context, insAdminCli *insadmin.InstanceAdminClient, projectName, instanceName string) error {
	insOp, err := insAdminCli.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", projectName),
		InstanceId: instanceName,
	})
	if err != nil {
		return fmt.Errorf("failed to create instance operation: %v", err)
	}

	if _, err := insOp.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait operation done: %v", err)
	}

	return nil
}

func CreateDatabase(ctx context.Context, dbAdminCli *dbadmin.DatabaseAdminClient, projectName, instanceName, dbName string) error {
	dbOp, err := dbAdminCli.CreateDatabase(ctx, &dbadminpb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", projectName, instanceName),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", dbName),
	})
	if err != nil {
		return fmt.Errorf("failed to create database operation: %v", err)
	}

	if _, err := dbOp.Wait(ctx); err != nil {
		return fmt.Errorf("failed to wait operation done: %v", err)
	}

	return nil
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
	fpath := filepath.Join(dir, "./test/testdata/schema.sql")
	file, err := os.Open(fpath)
	if err != nil {
		return fmt.Errorf("scheme file cannot open: %v", err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read error: %v", err)
	}

	// find the first DDL to skip comments.
	contents := string(b)
	ddls, err := (&parser.Parser{
		Lexer: &parser.Lexer{
			File: &mtoken.File{FilePath: fpath, Buffer: contents},
		},
	}).ParseDDLs()
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}
	// Split scheme definition to DDL statements.
	// This assuemes there is no comments and each statement is separated by semi-colon
	var statements []string
	for _, ddl := range ddls {
		statements = append(statements, ddl.SQL())
	}

	// Apply DDL statements to create tables
	return ApplyDDL(ctx, adminClient, dbname, statements...)
}
