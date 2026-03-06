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
		Short: "Search for tables and datasets across the Dune catalog",
		Long: "Natural-language table discovery across the Dune catalog. Use this command\n" +
			"to find concrete table names for use in SQL queries.\n\n" +
			"Filter by category (canonical for chain primitives, decoded for ABI-level\n" +
			"events/calls, spell for curated datasets, community for user-contributed),\n" +
			"by blockchain, schema, dataset type, or ownership scope.",
		RunE: runSearch,
	}

	cmd.Flags().String("query", "", "natural-language search intent or entity hints (e.g. 'uniswap v3 swaps'); use '*' to browse without keyword bias")
	cmd.Flags().StringArray("categories", nil, "filter by table family: canonical (chain primitives), decoded (ABI-level events/calls), spell (curated datasets), community (user-contributed)")
	cmd.Flags().StringArray("blockchains", nil, "chain scope to reduce ambiguity and improve ranking (e.g. ethereum, solana)")
	cmd.Flags().StringArray("dataset-types", nil, "fine-grained dataset type filter: dune_table, decoded_table, spell, uploaded_table, transformation_table, transformation_view")
	cmd.Flags().StringArray("schemas", nil, "schema/namespace constraint for high precision (e.g. dex, uniswap_v3_ethereum)")
	cmd.Flags().String("owner-scope", "", "ownership filter: all, me, or team; does NOT automatically include private datasets")
	cmd.Flags().Bool("include-private", false, "widen results to include private datasets visible to the authenticated user/team alongside public ones")
	cmd.Flags().Bool("include-schema", false, "include column-level schema (name, type, nullable) for every result; useful when preparing SQL")
	cmd.Flags().Bool("include-metadata", false, "include category-specific metadata (page_rank_score, description, abi_type, contract_name, project_name, etc.)")
	cmd.Flags().Int32("limit", 20, "number of results per page; use 5-15 for quick checks, 20-50 for deeper exploration")
	cmd.Flags().Int32("offset", 0, "pagination offset; use previous response pagination info for next page")
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
