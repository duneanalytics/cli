package dataset

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
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
			"Filter by category, blockchain, schema, or dataset type. Each result lists the\n" +
			"columns the table is partitioned by: filter on them in SQL (e.g. WHERE block_month\n" +
			">= DATE '2026-01-01') to cut query cost. With -o json the top 3 hits also carry\n" +
			"their columns and types; --include-schema adds the schema of every result.\n\n" +
			"For contract-specific lookup, use 'dune dataset search-by-contract' instead.\n\n" +
			"Examples:\n" +
			"  dune dataset search --query \"uniswap swaps\"\n" +
			"  dune dataset search --query \"transfers\" --categories decoded --blockchains ethereum\n" +
			"  dune dataset search --query \"dex trades\" --categories spell --include-schema --output json\n" +
			"  dune dataset search --query \"*\" --dataset-types uploaded_table",
		RunE: runSearch,
	}

	cmd.Flags().String("query", "", "natural-language search intent or entity hints (e.g. 'uniswap v3 swaps'); use '*' to browse without keyword bias")
	cmd.Flags().StringArray("categories", nil, "filter by table family: canonical (chain primitives), decoded (ABI-level events/calls), spell (curated datasets), community (user-contributed)")
	cmd.Flags().StringArray("blockchains", nil, "chain scope to reduce ambiguity and improve ranking (e.g. ethereum, solana, arbitrum)")
	cmd.Flags().StringArray("dataset-types", nil,
		"fine-grained dataset type filter: dune_table, decoded_table, spell, uploaded_table, transformation_table, transformation_view")
	cmd.Flags().StringArray("schemas", nil, "schema/namespace constraint for high precision (e.g. dex, uniswap_v3_ethereum)")
	cmd.Flags().String("owner-scope", "", "no longer used")
	cmd.Flags().Bool("include-private", false, "no longer used")
	_ = cmd.Flags().MarkDeprecated("owner-scope", "it has no effect and will be removed")
	_ = cmd.Flags().MarkDeprecated("include-private", "it has no effect and will be removed")
	cmd.Flags().Bool("include-schema", false,
		"include column-level schema (name, type, nullable) for every result; the top 3 hits always include their columns")
	cmd.Flags().Bool("include-metadata", false, "include category-specific metadata (page_rank_score, description, abi_type, contract_name, project_name, etc.)")
	cmd.Flags().Int32("limit", 20, "number of results per page; use 5-15 for quick checks, 20-50 for deeper exploration")
	cmd.Flags().Int32("offset", 0, "pagination offset; combine with limit to page through large result sets")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

// searchResult is a search hit in the shape the Dune API's dataset search uses,
// plus the partition and column details only the Dune MCP's searchTables returns.
type searchResult struct {
	FullName         string          `json:"full_name"`
	Category         string          `json:"category"`
	DatasetType      string          `json:"dataset_type,omitempty"`
	Blockchains      []string        `json:"blockchains,omitempty"`
	PartitionColumns []string        `json:"partition_columns,omitempty"`
	Visibility       string          `json:"visibility,omitempty"`
	Description      string          `json:"description,omitempty"`
	Columns          json.RawMessage `json:"columns,omitempty"`
	Schema           json.RawMessage `json:"schema,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
}

type searchResponse struct {
	Total      int            `json:"total"`
	Results    []searchResult `json:"results"`
	Pagination struct {
		Limit      int  `json:"limit"`
		Offset     int  `json:"offset"`
		NextOffset *int `json:"next_offset,omitempty"`
		HasMore    bool `json:"has_more"`
	} `json:"pagination"`
}

func runSearch(cmd *cobra.Command, _ []string) error {
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"query":            "query",
		"categories":       "categories",
		"blockchains":      "blockchains",
		"dataset-types":    "datasetTypes",
		"schemas":          "schemas",
		"include-schema":   "includeSchema",
		"include-metadata": "includeMetadata",
		"limit":            "limit",
		"offset":           "offset",
	})
	var res struct {
		Total      int            `json:"total"`
		Results    []searchResult `json:"results"`
		Pagination struct {
			Limit      int  `json:"limit"`
			Offset     int  `json:"offset"`
			NextOffset *int `json:"nextOffset"`
			HasMore    bool `json:"hasMore"`
		} `json:"pagination"`
	}
	if _, err := cmdutil.CallTool(cmd, "searchTables", args, &res); err != nil {
		return err
	}
	resp := searchResponse{Total: res.Total, Results: res.Results}
	resp.Pagination.Limit = res.Pagination.Limit
	resp.Pagination.Offset = res.Pagination.Offset
	resp.Pagination.NextOffset = res.Pagination.NextOffset
	resp.Pagination.HasMore = res.Pagination.HasMore

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		columns := []string{"FULL_NAME", "CATEGORY", "DATASET_TYPE", "PARTITIONED_BY", "BLOCKCHAINS"}
		rows := make([][]string, len(resp.Results))
		for i, r := range resp.Results {
			rows[i] = []string{
				r.FullName,
				r.Category,
				r.DatasetType,
				strings.Join(r.PartitionColumns, ", "),
				strings.Join(r.Blockchains, ", "),
			}
		}
		output.PrintTable(w, columns, rows)
		fmt.Fprintf(w, "\n%d of %d results\n", len(resp.Results), resp.Total)
		return nil
	}
}
