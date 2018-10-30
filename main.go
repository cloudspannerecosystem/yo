package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/spanner"
	"go.mercari.io/yo/generator"
	"go.mercari.io/yo/internal"
	"go.mercari.io/yo/loaders"
)

func main() {
	err := func() error {
		args, err := internal.ProcessArgs()
		if err != nil {
			return err
		}

		spannerClient, err := connectSpanner(args)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		spannerLoader := loaders.NewSpannerLoader(spannerClient)
		loader := internal.NewTypeLoader(spannerLoader)

		// load custom type definitions
		if args.CustomTypesFile != "" {
			if err := loader.LoadCustomTypes(args.CustomTypesFile); err != nil {
				return fmt.Errorf("load custom types file failed: %v", err)
			}
		}

		// load defs into type map
		tableMap, ixMap, err := loader.LoadSchema(args)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}

		g := generator.NewGenerator(loader, generator.GeneratorOption{
			PackageName:       args.Package,
			Tags:              args.Tags,
			TemplatePath:      args.TemplatePath,
			CustomTypePackage: args.CustomTypePackage,
			FilenameSuffix:    args.Suffix,
			SingleFile:        args.SingleFile,
			Filename:          args.Filename,
			Path:              args.Path,
		})
		if err := g.Generate(tableMap, ixMap); err != nil {
			return fmt.Errorf("error: %v", err)
		}

		return nil
	}()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func connectSpanner(args *internal.ArgType) (*spanner.Client, error) {
	ctx := context.Background()

	databaseName := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		args.Project, args.Instance, args.Database)
	spannerClient, err := spanner.NewClient(ctx, databaseName)
	if err != nil {
		return nil, err
	}

	return spannerClient, nil
}
