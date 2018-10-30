package internal

import (
	"fmt"
	"os"
	pathpkg "path"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultSuffix = ".yo.go"
	exampleUsage  = `
  # Generate models under models directory
  yo $PROJECT_NAME $INSTANCE_NAME $DATABASE_NAME -o models

  # Generate models under models directory with custom types
  yo $PROJECT_NAME $INSTANCE_NAME $DATABASE_NAME -o models --custom-types-file custom_column_types.yml
`
)

var (
	rootOpts = ArgType{}
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

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func init() {
	rootCmd.Flags().StringVar(&rootOpts.CustomTypesFile, "custom-types-file", "", "custom table field type definition file")
	rootCmd.Flags().StringVarP(&rootOpts.Out, "out", "o", "", "output path or file name")
	rootCmd.Flags().StringVar(&rootOpts.Suffix, "suffix", defaultSuffix, "output file suffix")
	rootCmd.Flags().BoolVar(&rootOpts.SingleFile, "single-file", false, "toggle single file output")
	rootCmd.Flags().StringVarP(&rootOpts.Package, "package", "p", "", "package name used in generated Go code")
	rootCmd.Flags().StringVar(&rootOpts.CustomTypePackage, "custom-type-package", "", "Go package name to use for custom or unknown types")
	rootCmd.Flags().StringArrayVar(&rootOpts.IgnoreFields, "ignore-fields", nil, "fields to exclude from the generated Go code types")
	rootCmd.Flags().StringVar(&rootOpts.TemplatePath, "template-path", "", "user supplied template path")
	rootCmd.Flags().StringVar(&rootOpts.Tags, "tags", "", "build tags to add to package header")

	helpFn := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		helpFn(cmd, args)
		os.Exit(1)
	})
}

// ArgType is the type that specifies the command line arguments.
type ArgType struct {
	// Project is the GCP project string
	Project string

	// Instance is the instance string
	Instance string

	// Database is the database string
	Database string

	// CustomTypesFile is the path for custom table field type definition file (xx.yml)
	CustomTypesFile string

	// Out is the output path. If Out is a file, then that will be used as the
	// path. If Out is a directory, then the output file will be
	// Out/<$CWD>.yo.go
	Out string

	// Suffix is the output suffix for filenames.
	Suffix string

	// SingleFile when toggled changes behavior so that output is to one f ile.
	SingleFile bool

	// Package is the name used to generate package headers. If not specified,
	// the name of the output directory will be used instead.
	Package string

	// CustomTypePackage is the Go package name to use for unknown types.
	CustomTypePackage string

	// IgnoreFields allows the user to specify field names which should not be
	// handled by yo in the generated code.
	IgnoreFields []string

	// TemplatePath is the path to use the user supplied templates instead of
	// the built in versions.
	TemplatePath string

	// Tags is the list of build tags to add to generated Go files.
	Tags string

	Path     string
	Filename string
}

func ProcessArgs() (*ArgType, error) {
	if err := rootCmd.Execute(); err != nil {
		return nil, err
	}
	return &rootOpts, nil
}

func processArgs(args *ArgType, argv []string) error {
	rootOpts.Project = argv[0]
	rootOpts.Instance = argv[1]
	rootOpts.Database = argv[2]

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
