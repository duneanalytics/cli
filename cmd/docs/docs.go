package docs

import "github.com/spf13/cobra"

// NewDocsCmd returns the `docs` parent command.
func NewDocsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "docs",
		Short:       "Search and browse Dune documentation",
		Annotations: map[string]string{"skipAuth": "true"},
	}
	cmd.AddCommand(newSearchCmd())
	return cmd
}
