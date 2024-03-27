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

package cmd

import (
	"context"
	"fmt"
	"os"
	pathpkg "path"
	"runtime/debug"
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/spf13/cobra"
	"go.mercari.io/yo/generator"
	"go.mercari.io/yo/internal"
	"go.mercari.io/yo/loaders"
)

const (
	defaultSuffix = ".yo.go"
	exampleUsage  = `
  # Generate models under models directory
  yo $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models

  # Generate models under models directory with custom types
  yo $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models --custom-types-file custom_column_types.yml
`
)

var version string

var (
	rootOpts = internal.ArgType{}
	rootCmd  = &cobra.Command{
		Use:   "yo PROJECT_NAME INSTANCE_NAME DATABASE_NAME",
		Short: "yo is a command-line tool to generate Go code for Google Cloud Spanner.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return fmt.Errorf("must specify 3 arguments")
			}
			return nil
		},
		Example: strings.Trim(exampleUsage, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := processArgs(&rootOpts, args); err != nil {
				return err
			}

			spannerClient, err := connectSpanner(&rootOpts)
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}
			spannerLoader := loaders.NewSpannerLoader(spannerClient)
			inflector, err := internal.NewInflector(rootOpts.InflectionRuleFile)
			if err != nil {
				return fmt.Errorf("load inflection rule failed: %v", err)
			}
			loader := internal.NewTypeLoader(spannerLoader, inflector)

			// load custom type definitions
			if rootOpts.CustomTypesFile != "" {
				if err := loader.LoadCustomTypes(rootOpts.CustomTypesFile); err != nil {
					return fmt.Errorf("load custom types file failed: %v", err)
				}
			}

			// load defs into type map
			tableMap, ixMap, err := loader.LoadSchema(&rootOpts)
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}

			g := generator.NewGenerator(loader, inflector, generator.GeneratorOption{
				PackageName:       rootOpts.Package,
				Tags:              rootOpts.Tags,
				TemplatePath:      rootOpts.TemplatePath,
				CustomTypePackage: rootOpts.CustomTypePackage,
				FilenameSuffix:    rootOpts.Suffix,
				SingleFile:        rootOpts.SingleFile,
				Filename:          rootOpts.Filename,
				Path:              rootOpts.Path,
			})
			if err := g.Generate(tableMap, ixMap); err != nil {
				return fmt.Errorf("error: %v", err)
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       versionInfo(),
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	setRootOpts(rootCmd, &rootOpts)
}

func setRootOpts(cmd *cobra.Command, opts *internal.ArgType) {
	cmd.Flags().StringVar(&opts.CustomTypesFile, "custom-types-file", "", "custom table field type definition file")
	cmd.Flags().StringVarP(&opts.Out, "out", "o", "", "output path or file name")
	cmd.Flags().StringVar(&opts.Suffix, "suffix", defaultSuffix, "output file suffix")
	cmd.Flags().BoolVar(&opts.SingleFile, "single-file", false, "toggle single file output")
	cmd.Flags().BoolVar(&opts.FilenameWithUnderscores, "with-underscores", false, "toggle underscores in file names")
	cmd.Flags().StringVarP(&opts.Package, "package", "p", "", "package name used in generated Go code")
	cmd.Flags().StringVar(&opts.CustomTypePackage, "custom-type-package", "", "Go package name to use for custom or unknown types")
	cmd.Flags().StringArrayVar(&opts.TargetTables, "target-tables", nil, "tables to include from the generated Go code")
	cmd.Flags().StringArrayVar(&opts.IgnoreFields, "ignore-fields", nil, "fields to exclude from the generated Go code")
	cmd.Flags().StringArrayVar(&opts.IgnoreTables, "ignore-tables", nil, "tables to exclude from the generated Go code")
	cmd.Flags().StringVar(&opts.TemplatePath, "template-path", "", "user supplied template path")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "build tags to add to package header")
	cmd.Flags().StringVar(&opts.InflectionRuleFile, "inflection-rule-file", "", "custom inflection rule file")

	helpFn := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		helpFn(cmd, args)
		os.Exit(1)
	})
}

func processArgs(args *internal.ArgType, argv []string) error {
	if len(argv) == 3 {
		rootOpts.Project = argv[0]
		rootOpts.Instance = argv[1]
		rootOpts.Database = argv[2]
	} else {
		rootOpts.DDLFilepath = argv[0]
	}

	path := ""
	filename := ""

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// determine out path
	if args.Out == "" {
		path = cwd
	} else {
		// determine what to do with Out
		fi, err := os.Stat(args.Out)
		if err == nil && fi.IsDir() {
			// out is directory
			path = args.Out
		} else if err == nil && !fi.IsDir() {
			// file exists (will truncate later)
			path = pathpkg.Dir(args.Out)
			filename = pathpkg.Base(args.Out)

			// error if not split was set, but destination is not a directory
			if !args.SingleFile {
				return fmt.Errorf("output path is not directory")
			}
		} else if _, ok := err.(*os.PathError); ok {
			// path error (ie, file doesn't exist yet)
			path = pathpkg.Dir(args.Out)
			filename = pathpkg.Base(args.Out)

			// error if split was set, but dest doesn't exist
			if !args.SingleFile {
				return fmt.Errorf("output path must be a directory and already exist when not writing to a single file")
			}
		} else {
			return err
		}
	}

	// check template path
	if args.TemplatePath != "" {
		info, err := os.Stat(args.TemplatePath)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("template path is not directory")
		}
	}

	// fix path
	if path == "." {
		path = cwd
	}

	// determine package name
	if args.Package == "" {
		args.Package = pathpkg.Base(path)
	}

	// determine filename if not previously set
	if filename == "" {
		filename = args.Package + args.Suffix
	}

	args.Path = path
	args.Filename = filename

	return nil
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

func versionInfo() string {
	if version != "" {
		return version
	}

	// For those who "go install" yo
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return info.Main.Version
}
