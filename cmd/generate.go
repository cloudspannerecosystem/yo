package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.mercari.io/yo/generator"
	"go.mercari.io/yo/internal"
	"go.mercari.io/yo/loaders"
)

var (
	generateOpts = internal.ArgType{}
	generateCmd  = &cobra.Command{
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
  yo generate $PROJECT_NAME $INSTANCE_NAME $DATABASE_NAME -o models

  # Generate models under models directory with custom types
  yo generate $PROJECT_NAME $INSTANCE_NAME $DATABASE_NAME -o models --custom-types-file custom_column_types.yml
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := processArgs(&generateOpts, args); err != nil {
				return err
			}

			var loader *internal.TypeLoader
			if generateOpts.FromDDL {
				spannerLoader, err := loaders.NewSpannerLoaderFromDDL(args[0])
				if err != nil {
					return fmt.Errorf("error: %v", err)
				}
				loader = internal.NewTypeLoader(spannerLoader)
			} else {
				spannerClient, err := connectSpanner(&rootOpts)
				if err != nil {
					return fmt.Errorf("error: %v", err)
				}
				spannerLoader := loaders.NewSpannerLoader(spannerClient)
				loader = internal.NewTypeLoader(spannerLoader)
			}

			// load custom type definitions
			if generateOpts.CustomTypesFile != "" {
				if err := loader.LoadCustomTypes(generateOpts.CustomTypesFile); err != nil {
					return fmt.Errorf("load custom types file failed: %v", err)
				}
			}

			// load defs into type map
			tableMap, ixMap, err := loader.LoadSchema(&generateOpts)
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}

			g := generator.NewGenerator(loader, generator.GeneratorOption{
				PackageName:       generateOpts.Package,
				Tags:              generateOpts.Tags,
				TemplatePath:      generateOpts.TemplatePath,
				CustomTypePackage: generateOpts.CustomTypePackage,
				FilenameSuffix:    generateOpts.Suffix,
				SingleFile:        generateOpts.SingleFile,
				Filename:          generateOpts.Filename,
				Path:              generateOpts.Path,
			})
			if err := g.Generate(tableMap, ixMap); err != nil {
				return fmt.Errorf("error: %v", err)
			}

			return nil
		},
	}
)

func init() {
	generateCmd.Flags().BoolVar(&generateOpts.FromDDL, "from-ddl", false, "toggle using ddl file")
	setRootOpts(generateCmd, &generateOpts)
	rootCmd.AddCommand(generateCmd)
}
