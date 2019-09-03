package cmd

import (
	"github.com/spf13/cobra"
	"go.mercari.io/yo/generator"
)

var createTemplatePath string

var createTemplateCmd = &cobra.Command{
	Use:     "create-template",
	Short:   "yo create-template generates default template files ",
	Example: `yo create-template --template-path templates`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generator.CopyDefaultTemplates(createTemplatePath)
	},
}

func init() {
	createTemplateCmd.Flags().StringVar(&createTemplatePath, "template-path", "", "destination template path")
	rootCmd.AddCommand(createTemplateCmd)
}
