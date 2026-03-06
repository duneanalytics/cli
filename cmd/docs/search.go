package docs

import (
	"encoding/json"
	"fmt"

	"github.com/duneanalytics/cli/output"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

const defaultMCPEndpoint = "https://docs.dune.com/mcp"

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search the Dune documentation for guides, API references, and code examples",
		Long: "Search across all Dune documentation pages including guides, API references,\n" +
			"DuneSQL syntax, and code examples. Does not require authentication.",
		Annotations: map[string]string{"skipAuth": "true"},
		RunE:        runSearch,
	}

	cmd.Flags().String("query", "", "search query text, e.g. 'DuneSQL date functions' or 'API authentication' (required)")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Bool("api-reference-only", false, "prioritize API reference pages over conceptual guides")
	cmd.Flags().Bool("code-only", false, "prioritize pages with executable examples and code snippets")
	cmd.Flags().String("mcp-endpoint", defaultMCPEndpoint, "MCP server endpoint URL")
	_ = cmd.Flags().MarkHidden("mcp-endpoint")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

// docResult represents a single documentation search result.
type docResult struct {
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// searchResponse is the JSON output envelope.
type searchResponse struct {
	Query   string      `json:"query"`
	Results []docResult `json:"results"`
}

func runSearch(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	query, _ := cmd.Flags().GetString("query")
	apiRefOnly, _ := cmd.Flags().GetBool("api-reference-only")
	codeOnly, _ := cmd.Flags().GetBool("code-only")
	endpoint, _ := cmd.Flags().GetString("mcp-endpoint")

	// Build MCP tool arguments.
	args := map[string]any{"query": query}
	if apiRefOnly {
		args["apiReferenceOnly"] = true
	}
	if codeOnly {
		args["codeOnly"] = true
	}

	// Connect to the Mintlify MCP server.
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "dune-cli",
		Version: "1.0.0",
	}, nil)

	transport := &mcp.StreamableClientTransport{Endpoint: endpoint}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("connecting to docs server: %w", err)
	}
	defer session.Close()

	// Call the SearchDuneDocs tool.
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "SearchDuneDocs",
		Arguments: args,
	})
	if err != nil {
		return fmt.Errorf("searching docs: %w", err)
	}

	// Parse results from content blocks.
	results := parseResults(result)

	// Output.
	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, searchResponse{
			Query:   query,
			Results: results,
		})
	default:
		if len(results) == 0 {
			fmt.Fprintln(w, "No results found.")
			return nil
		}
		for i, r := range results {
			if i > 0 {
				fmt.Fprintln(w)
			}
			if r.Title != "" {
				fmt.Fprintf(w, "  %s\n", r.Title)
			}
			if r.URL != "" {
				fmt.Fprintf(w, "  %s\n", r.URL)
			}
			if r.Description != "" {
				fmt.Fprintf(w, "  %s\n", r.Description)
			}
		}
		fmt.Fprintf(w, "\n%d result(s)\n", len(results))
		return nil
	}
}

// parseResults extracts doc results from MCP tool response content blocks.
func parseResults(result *mcp.CallToolResult) []docResult {
	var results []docResult
	for _, c := range result.Content {
		tc, ok := c.(*mcp.TextContent)
		if !ok {
			continue
		}
		// Try to parse as JSON (single result or array).
		var single docResult
		if err := json.Unmarshal([]byte(tc.Text), &single); err == nil {
			results = append(results, single)
			continue
		}
		var arr []docResult
		if err := json.Unmarshal([]byte(tc.Text), &arr); err == nil {
			results = append(results, arr...)
			continue
		}
		// Fallback: treat raw text as a description-only result.
		results = append(results, docResult{Description: tc.Text})
	}
	return results
}
