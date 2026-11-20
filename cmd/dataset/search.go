package dataset

import (
	"encoding/json"
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
			"The catalog includes:\n" +
			"  - canonical: chain-native primitives (blocks, transactions, logs, traces)\n" +
			"  - decoded: ABI-level events and function calls parsed from contract interactions\n" +
			"  - spell: curated Spellbook transformation datasets (e.g. dex.trades)\n" +
			"  - community: user-contributed tables and uploads\n\n" +
			"Filter by category, blockchain, schema, dataset type, or ownership scope. Each\n" +
			"result lists the columns the table is partitioned by: filter on them in SQL (e.g.\n" +
			"WHERE block_month >= DATE '2026-01-01') to cut query cost. With -o json the top 3 hits also carry\n" +
			"their columns and types; --include-schema adds the schema of every result.\n\n" +
			"For contract-specific lookup, use 'dune dataset search-by-contract' instead.\n\n" +
			"Examples:\n" +
			"  dune dataset search --query \"uniswap swaps\"\n" +
			"  dune dataset search --query \"transfers\" --categories decoded --blockchains ethereum\n" +
			"  dune dataset search --query \"dex trades\" --categories spell --include-schema --output json\n" +
			"  dune dataset search --query \"*\" --dataset-types uploaded_table\n" +
			"  dune dataset search --query \"*\" --owner-scope me",
		RunE: runSearch,
	}

	cmd.Flags().String("query", "", "natural-language search intent or entity hints (e.g. 'uniswap v3 swaps'); use '*' to browse without keyword bias")
	cmd.Flags().StringArray("categories", nil, "filter by table family: canonical (chain primitives), decoded (ABI-level events/calls), spell (curated datasets), community (user-contributed)")
	cmd.Flags().StringArray("blockchains", nil, "chain scope to reduce ambiguity and improve ranking (e.g. ethereum, solana, arbitrum)")
	cmd.Flags().StringArray("dataset-types", nil,
		"fine-grained dataset type filter: dune_table, decoded_table, spell, uploaded_table, transformation_table, transformation_view")
	cmd.Flags().StringArray("schemas", nil, "schema/namespace constraint for high precision (e.g. dex, uniswap_v3_ethereum)")
	cmd.Flags().String("owner-scope", "", "ownership filter: all (default), me, or team; does NOT automatically include private datasets")
	cmd.Flags().Bool("include-private", false, "widen results to include private datasets visible to the authenticated user/team alongside public ones")
	cmd.Flags().Bool("include-schema", false,
		"include column-level schema (name, type, nullable) for every result; the top 3 hits always include their columns")
	cmd.Flags().Bool("include-metadata", false, "include category-specific metadata (page_rank_score, description, abi_type, contract_name, project_name, etc.)")
	cmd.Flags().Int32("limit", 20, "number of results per page; use 5-15 for quick checks, 20-50 for deeper exploration")
	cmd.Flags().Int32("offset", 0, "pagination offset; combine with limit to page through large result sets")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

// searchResult is a hit from the Dune API's dataset search, plus the partition
// columns and top-hit columns that only the Dune MCP's searchTables returns.
type searchResult struct {
	models.SearchDatasetResult
	PartitionColumns []string        `json:"partition_columns,omitempty"`
	Columns          json.RawMessage `json:"columns,omitempty"`
}

type searchResponse struct {
	Total      int32                           `json:"total"`
	Results    []searchResult                  `json:"results"`
	Pagination models.SearchDatasetsPagination `json:"pagination"`
}

func runSearch(cmd *cobra.Command, _ []string) error {
	fs := cmd.Flags()
	req := models.SearchDatasetsRequest{}
	if fs.Changed("query") {
		v, _ := fs.GetString("query")
		req.Query = &v
	}
	req.Categories, _ = fs.GetStringArray("categories")
	req.Blockchains, _ = fs.GetStringArray("blockchains")
	req.DatasetTypes, _ = fs.GetStringArray("dataset-types")
	req.Schemas, _ = fs.GetStringArray("schemas")
	if fs.Changed("owner-scope") {
		v, _ := fs.GetString("owner-scope")
		req.OwnerScope = &v
	}
	if fs.Changed("include-private") {
		v, _ := fs.GetBool("include-private")
		req.IncludePrivate = &v
	}
	if fs.Changed("include-schema") {
		v, _ := fs.GetBool("include-schema")
		req.IncludeSchema = &v
	}
	if fs.Changed("include-metadata") {
		v, _ := fs.GetBool("include-metadata")
		req.IncludeMetadata = &v
	}
	if fs.Changed("limit") {
		v, _ := fs.GetInt32("limit")
		req.Limit = &v
	}
	if fs.Changed("offset") {
		v, _ := fs.GetInt32("offset")
		req.Offset = &v
	}

	found, err := cmdutil.ClientFromCmd(cmd).SearchDatasets(req)
	if err != nil {
		return err
	}
	// Non-nil, so no hits print "results": [] as before rather than null.
	resp := searchResponse{Total: found.Total, Pagination: found.Pagination,
		Results: make([]searchResult, 0, len(found.Results))}
	for _, r := range found.Results {
		resp.Results = append(resp.Results, searchResult{SearchDatasetResult: r})
	}
	addMCPColumns(cmd, resp.Results)

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		columns := []string{"FULL_NAME", "CATEGORY", "DATASET_TYPE", "PARTITIONED_BY", "BLOCKCHAINS"}
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
				strings.Join(r.PartitionColumns, ", "),
				strings.Join(r.Blockchains, ", "),
			}
		}
		output.PrintTable(w, columns, rows)
		fmt.Fprintf(w, "\n%d of %d results\n", len(resp.Results), resp.Total)
		return nil
	}
}

// addMCPColumns adds the partition columns and columns that the Dune MCP's
// searchTables reports for the same search, matched by table name. They are
// extras: when the MCP call fails the results keep only the REST fields.
func addMCPColumns(cmd *cobra.Command, results []searchResult) {
	if len(results) == 0 {
		return
	}
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"query":         "query",
		"categories":    "categories",
		"blockchains":   "blockchains",
		"dataset-types": "datasetTypes",
		"schemas":       "schemas",
		"limit":         "limit",
		"offset":        "offset",
	})
	var res struct {
		Results []struct {
			FullName         string          `json:"full_name"`
			PartitionColumns []string        `json:"partition_columns"`
			Columns          json.RawMessage `json:"columns"`
		} `json:"results"`
	}
	if _, err := cmdutil.CallTool(cmd, "searchTables", args, &res); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not read partition columns: %v\n", err)
		return
	}
	for _, extra := range res.Results {
		for i := range results {
			if results[i].FullName == extra.FullName {
				results[i].PartitionColumns = extra.PartitionColumns
				results[i].Columns = extra.Columns
			}
		}
	}
}
