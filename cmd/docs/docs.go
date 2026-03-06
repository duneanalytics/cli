package docs

import "github.com/spf13/cobra"

// NewDocsCmd returns the `docs` parent command.
func NewDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "docs",
		Short:       "Search the Dune documentation for guides, API references, and examples",
		Annotations: map[string]string{"skipAuth": "true"},
	}
	cmd.AddCommand(newSearchCmd())
	return cmd
}
