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
	"fmt"

	"github.com/spf13/cobra"
	"go.mercari.io/yo/v2/generator"
	"go.mercari.io/yo/v2/internal"
	"go.mercari.io/yo/v2/loaders"
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
  yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models

  # Generate models under models directory with custom types
  yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models --custom-types-file custom_column_types.yml
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := processArgs(&generateOpts, args); err != nil {
				return err
			}

			inflector, err := internal.NewInflector(generateOpts.InflectionRuleFile)
			if err != nil {
				return fmt.Errorf("load inflection rule failed: %v", err)
			}
			var loader *internal.TypeLoader
			if generateOpts.FromDDL {
				spannerLoader, err := loaders.NewSpannerLoaderFromDDL(args[0])
				if err != nil {
					return fmt.Errorf("error: %v", err)
				}
				loader = internal.NewTypeLoader(spannerLoader, inflector)
			} else {
				spannerClient, err := connectSpanner(&rootOpts)
				if err != nil {
					return fmt.Errorf("error: %v", err)
				}
				spannerLoader := loaders.NewSpannerLoader(spannerClient)
				loader = internal.NewTypeLoader(spannerLoader, inflector)
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

			g := generator.NewGenerator(loader, inflector, generator.GeneratorOption{
				PackageName:       generateOpts.Package,
				Tags:              generateOpts.Tags,
				TemplatePath:      generateOpts.TemplatePath,
				CustomTypePackage: generateOpts.CustomTypePackage,
				FilenameSuffix:    generateOpts.Suffix,
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
