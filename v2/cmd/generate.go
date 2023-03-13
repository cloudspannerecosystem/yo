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
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.mercari.io/yo/v2/config"
	"go.mercari.io/yo/v2/generator"
	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/loader"
	"go.mercari.io/yo/v2/module"
	"go.mercari.io/yo/v2/module/builtin"
)

var (
	defaultHeaderModule  = builtin.Header
	defaultGlobalModules = []module.Module{builtin.Interface}
	defaultTypeModules   = []module.Module{builtin.Type, builtin.Operation}
)

// generateCmdOption is the type that specifies the command line arguments.
type generateCmdOption struct {
	// Project is the GCP project string
	Project string

	// Instance is the instance string
	Instance string

	// Database is the database string
	Database string

	// Out is the output path. If Out is a file, then that will be used as the
	// path. If Out is a directory, then the output file will be
	// Out/<$CWD>.yo.go
	Out string

	// Suffix is the output suffix for filenames.
	Suffix string

	// Package is the name used to generate package headers. If not specified,
	// the name of the output directory will be used instead.
	Package string

	// Tags is the list of build tags to add to generated Go files.
	Tags string

	// DDLFilepath is the filepath of the ddl file.
	DDLFilepath string

	// FromDDL indicates generating from ddl flie or not.
	FromDDL bool

	// IgnoreFields allows the user to specify field names which should not be
	// handled by yo in the generated code.
	IgnoreFields []string

	// IgnoreTables allows the user to specify table names which should not be
	// handled by yo in the generated code.
	IgnoreTables []string

	// Path to config file
	ConfigFile string

	// DisableDefaultModules disable to use the default modules for code generation
	DisableDefaultModules bool

	// DisableFormat disable to apply gofmt to generated files
	DisableFormat bool

	HeaderModule            string
	AdditionalGlobalModules []string
	AdditionalTypeModules   []string

	// UseLegacyIndexModule uses legacy index module instead of the default index module
	UseLegacyIndexModule bool

	// IgnoreUnsupportedStatements enable to ignore yo unsupported statement when parse statements
	IgnoreUnsupportedStatements bool
	baseDir                     string
}

var (
	generateCmdOpts = generateCmdOption{}
	generateCmd     = &cobra.Command{
		Use:   "generate",
		Short: "yo generate generates Go code from ddl file.",
		Args: func(cmd *cobra.Command, args []string) error {
			if l := len(args); l != 1 && l != 3 {
				return fmt.Errorf("must specify 1 or 3 arguments")
			}
			return nil
		},
		Example: `  # Generate models from ddl under models directory
  yo generate schema.sql --from-ddl -o models

  # Generate models from ddl under models directory with custom types
  yo generate schema.sql --from-ddl -o models --custom-types-file custom_column_types.yml

  # Generate models under models directory
  yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models

  # Generate models under models directory with custom types
  yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models --custom-types-file custom_column_types.yml
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			cfg, err := config.Load(generateCmdOpts.ConfigFile)
			if err != nil {
				return err
			}

			if err := processGenerateCmdOption(&generateCmdOpts, args); err != nil {
				return err
			}

			inflector, err := internal.NewInflector(cfg.Inflections)
			if err != nil {
				return fmt.Errorf("load inflection rule failed: %v", err)
			}

			var source loader.SchemaSource
			if generateCmdOpts.FromDDL {
				source, err = loader.NewSchemaParserSource(generateCmdOpts.DDLFilepath, generateCmdOpts.IgnoreUnsupportedStatements)
				if err != nil {
					return fmt.Errorf("failed to create spanner loader: %v", err)
				}
			} else {
				spannerClient, err := connectSpanner(ctx, generateCmdOpts.Project, generateCmdOpts.Instance, generateCmdOpts.Database)
				if err != nil {
					return fmt.Errorf("failed to connect spanner: %v", err)
				}
				source, err = loader.NewInformationSchemaSource(spannerClient)
				if err != nil {
					return fmt.Errorf("failed to create spanner loader: %v", err)
				}
			}

			typeLoader := loader.NewTypeLoader(source, inflector, loader.Option{
				Config:       cfg,
				IgnoreTables: generateCmdOpts.IgnoreTables,
				IgnoreFields: generateCmdOpts.IgnoreFields,
			})

			// load defs into type map
			schema, err := typeLoader.LoadSchema()
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}

			headerModule, globalModules, typeModules := decideModules(&generateCmdOpts)

			g := generator.NewGenerator(typeLoader, inflector, generator.GeneratorOption{
				PackageName:    generateCmdOpts.Package,
				Tags:           generateCmdOpts.Tags,
				FilenameSuffix: generateCmdOpts.Suffix,
				BaseDir:        generateCmdOpts.baseDir,
				DisableFormat:  generateCmdOpts.DisableFormat,

				HeaderModule:  headerModule,
				GlobalModules: globalModules,
				TypeModules:   typeModules,
			})
			if err := g.Generate(schema); err != nil {
				return fmt.Errorf("error: %v", err)
			}

			return nil
		},
	}
)

func init() {
	generateCmd.Flags().StringVarP(&generateCmdOpts.ConfigFile, "config", "c", "", "path to Yo config file")
	generateCmd.Flags().BoolVar(&generateCmdOpts.FromDDL, "from-ddl", false, "toggle using ddl file")
	generateCmd.Flags().StringVarP(&generateCmdOpts.Out, "out", "o", "", "output path or file name")
	generateCmd.Flags().StringVar(&generateCmdOpts.Suffix, "suffix", defaultSuffix, "output file suffix")
	generateCmd.Flags().StringVarP(&generateCmdOpts.Package, "package", "p", "", "package name used in generated Go code")
	generateCmd.Flags().StringArrayVar(&generateCmdOpts.IgnoreFields, "ignore-fields", nil, "fields to exclude from the generated Go code types")
	generateCmd.Flags().StringArrayVar(&generateCmdOpts.IgnoreTables, "ignore-tables", nil, "tables to exclude from the generated Go code types")
	generateCmd.Flags().StringVar(&generateCmdOpts.Tags, "tags", "", "build tags to add to package header")
	generateCmd.Flags().BoolVar(&generateCmdOpts.DisableDefaultModules, "disable-default-modules", false, "disable the default modules for code generation")
	generateCmd.Flags().BoolVar(&generateCmdOpts.DisableFormat, "disable-format", false, "disable to apply gofmt to generated files")
	generateCmd.Flags().StringVar(&generateCmdOpts.HeaderModule, "header-module", "", "replace the default header module by user defined module")
	generateCmd.Flags().StringArrayVar(&generateCmdOpts.AdditionalGlobalModules, "global-module", nil, "add user defined module to global modules")
	generateCmd.Flags().StringArrayVar(&generateCmdOpts.AdditionalTypeModules, "type-module", nil, "add user defined module to type modules")
	generateCmd.Flags().BoolVar(&generateCmdOpts.UseLegacyIndexModule, "use-legacy-index-module", false, "use legacy index func name")
	generateCmd.Flags().BoolVar(&generateCmdOpts.IgnoreUnsupportedStatements, "ignore-unsupported-statements", false, "enable to ignore yo unsupported statement when parse statements")

	helpFn := generateCmd.HelpFunc()
	generateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		helpFn(cmd, args)
		os.Exit(1)
	})

	rootCmd.AddCommand(generateCmd)
}

func processGenerateCmdOption(opts *generateCmdOption, argv []string) error {
	if len(argv) == 3 {
		opts.Project = argv[0]
		opts.Instance = argv[1]
		opts.Database = argv[2]
	} else {
		opts.DDLFilepath = argv[0]
	}

	path := ""

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// determine out path
	if opts.Out == "" {
		path = cwd
	} else {
		// determine what to do with Out
		fi, err := os.Stat(opts.Out)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			// out is directory
			path = opts.Out
		} else {
			return fmt.Errorf("output path must be a directory")
		}
	}

	// fix path
	if path == "." {
		path = cwd
	}

	// determine package name
	if opts.Package == "" {
		opts.Package = pathpkg.Base(path)
	}

	opts.baseDir = path

	return nil
}

func decideModules(opts *generateCmdOption) (module.Module, []module.Module, []module.Module) {
	// header module uses null module that generates nothing when disabling default
	headerModule := builtin.NullHeader
	var globalModules []module.Module
	var typeModules []module.Module

	if !generateCmdOpts.DisableDefaultModules {
		headerModule = defaultHeaderModule
		globalModules = defaultGlobalModules
		typeModules = defaultTypeModules
		if generateCmdOpts.UseLegacyIndexModule {
			typeModules = append(typeModules, builtin.LegacyIndex)
		} else {
			typeModules = append(typeModules, builtin.Index)
		}
	}

	for _, path := range generateCmdOpts.AdditionalGlobalModules {
		basename := filepath.Base(path)
		for i := 0; i < 3; i++ {
			basename = basename[:len(basename)-len(filepath.Ext(basename))]
		}
		globalModules = append(globalModules, module.New(module.GlobalModule, basename, path))
	}

	for _, path := range generateCmdOpts.AdditionalTypeModules {
		basename := filepath.Base(path)
		for i := 0; i < 3; i++ {
			basename = basename[:len(basename)-len(filepath.Ext(basename))]
		}
		typeModules = append(typeModules, module.New(module.TypeModule, basename, path))
	}

	if path := generateCmdOpts.HeaderModule; path != "" {
		basename := filepath.Base(path)
		for i := 0; i < 3; i++ {
			basename = basename[:len(basename)-len(filepath.Ext(basename))]
		}
		headerModule = module.New(module.HeaderModule, basename, path)
	}

	return headerModule, globalModules, typeModules
}
