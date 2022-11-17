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
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultSuffix = ".yo.go"
	exampleUsage  = `
  # Generate models from ddl under models directory
  yo generate schema.sql --from-ddl -o models

  # Generate models under models directory
  yo generate $SPANNER_PROJECT_NAME $SPANNER_INSTANCE_NAME $SPANNER_DATABASE_NAME -o models
`
)

var version string

var (
	rootCmd = &cobra.Command{
		Use:   "yo",
		Short: "yo is a command-line tool to generate Go code for Google Cloud Spanner.",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Example:       strings.Trim(exampleUsage, "\n"),
		RunE:          nil,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       versionInfo(),
	}
)

func Execute() error {
	return rootCmd.Execute()
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
