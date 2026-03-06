package dataset

import (
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for datasets across the Dune catalog",
		Long: "Search for datasets across the Dune catalog by keyword, category, blockchain, and more.\n\n" +
			"The catalog includes canonical blockchain data, decoded contract tables,\n" +
			"Spellbook transformations, and community datasets. Use --include-schema\n" +
			"to get column names and types for SQL generation.\n\n" +
			"Examples:\n" +
			"  dune dataset search --query \"uniswap swaps\"\n" +
			"  dune dataset search --query \"transfers\" --categories decoded --blockchains ethereum\n" +
			"  dune dataset search --query \"dex trades\" --categories spell --include-schema --output json\n" +
			"  dune dataset search --owner-scope me",
		RunE: runSearch,
	}

	cmd.Flags().String("query", "", "search query text")
	cmd.Flags().StringArray("categories", nil, "filter by category (canonical, decoded, spell, community)")
	cmd.Flags().StringArray("blockchains", nil, "filter by blockchain")
	cmd.Flags().StringArray("dataset-types", nil,
		"filter by dataset type (dune_table, decoded_table, spell, uploaded_table, transformation_table, transformation_view)")
	cmd.Flags().StringArray("schemas", nil, "filter by schema")
	cmd.Flags().String("owner-scope", "", "ownership filter (all, me, team)")
	cmd.Flags().Bool("include-private", false, "include private datasets")
	cmd.Flags().Bool("include-schema", false, "include column schema in results")
	cmd.Flags().Bool("include-metadata", false, "include metadata in results")
	cmd.Flags().Int32("limit", 20, "maximum number of results")
	cmd.Flags().Int32("offset", 0, "pagination offset")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runSearch(cmd *cobra.Command, _ []string) error {
	client := cmdutil.ClientFromCmd(cmd)

	req := models.SearchDatasetsRequest{}

	if cmd.Flags().Changed("query") {
		v, _ := cmd.Flags().GetString("query")
		req.Query = &v
	}
	if cmd.Flags().Changed("categories") {
		v, _ := cmd.Flags().GetStringArray("categories")
		req.Categories = v
	}
	if cmd.Flags().Changed("blockchains") {
		v, _ := cmd.Flags().GetStringArray("blockchains")
		req.Blockchains = v
	}
	if cmd.Flags().Changed("dataset-types") {
		v, _ := cmd.Flags().GetStringArray("dataset-types")
		req.DatasetTypes = v
	}
	if cmd.Flags().Changed("schemas") {
		v, _ := cmd.Flags().GetStringArray("schemas")
		req.Schemas = v
	}
	if cmd.Flags().Changed("owner-scope") {
		v, _ := cmd.Flags().GetString("owner-scope")
		req.OwnerScope = &v
	}
	if cmd.Flags().Changed("include-private") {
		v, _ := cmd.Flags().GetBool("include-private")
		req.IncludePrivate = &v
	}
	if cmd.Flags().Changed("include-schema") {
		v, _ := cmd.Flags().GetBool("include-schema")
		req.IncludeSchema = &v
	}
	if cmd.Flags().Changed("include-metadata") {
		v, _ := cmd.Flags().GetBool("include-metadata")
		req.IncludeMetadata = &v
	}
	if cmd.Flags().Changed("limit") {
		v, _ := cmd.Flags().GetInt32("limit")
		req.Limit = &v
	}
	if cmd.Flags().Changed("offset") {
		v, _ := cmd.Flags().GetInt32("offset")
		req.Offset = &v
	}

	resp, err := client.SearchDatasets(req)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		columns := []string{"FULL_NAME", "CATEGORY", "DATASET_TYPE", "BLOCKCHAINS"}
		rows := make([][]string, len(resp.Results))
		for i, r := range resp.Results {
			dt := ""
			if r.DatasetType != nil {
				dt = *r.DatasetType
			}
			rows[i] = []string{
				r.FullName,
				r.Category,
				dt,
				strings.Join(r.Blockchains, ", "),
			}
		}
		output.PrintTable(w, columns, rows)
		fmt.Fprintf(w, "\n%d of %d results\n", len(resp.Results), resp.Total)
		return nil
	}
}
