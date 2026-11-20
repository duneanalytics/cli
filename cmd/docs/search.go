package docs

import (
	"fmt"
	"strings"

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
			"DuneSQL syntax, table naming conventions, and code examples.\n" +
			"Does not require authentication.\n\n" +
			"Useful for looking up:\n" +
			"  - DuneSQL functions and syntax (date/time, string, aggregate, window functions)\n" +
			"  - API endpoint references and authentication\n" +
			"  - Table naming conventions and dataset structure\n" +
			"  - Dune platform concepts (decoding, Spellbook, materialized views)\n\n" +
			"Examples:\n" +
			"  dune docs search --query \"DuneSQL string functions\"\n" +
			"  dune docs search --query \"execute query API\" --api-reference-only\n" +
			"  dune docs search --query \"decoded tables\" --code-only",
		Annotations: map[string]string{"skipAuth": "true"},
		RunE:        runSearch,
	}

	cmd.Flags().String("query", "", "natural-language search query for Dune docs; e.g. 'DuneSQL date functions', 'API authentication', 'decoded tables meaning', 'query pagination' (required)")
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

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "search_dune_docs",
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
			for _, line := range strings.Split(r.Description, "\n") {
				if line != "" {
					fmt.Fprintf(w, "  %s\n", line)
				}
			}
		}
		fmt.Fprintf(w, "\n%d result(s)\n", len(results))
		return nil
	}
}

// parseResults extracts doc results from MCP tool response content blocks. The
// docs server returns one text block per page section:
//
//	Title: <title>
//	Link: <url>
//	Page: <path>
//	Content: <markdown, possibly several lines>
func parseResults(result *mcp.CallToolResult) []docResult {
	var results []docResult
	for _, c := range result.Content {
		tc, ok := c.(*mcp.TextContent)
		if !ok {
			continue
		}
		results = append(results, parseResult(tc.Text))
	}
	return results
}

func parseResult(text string) docResult {
	var r docResult
	rest := text
	for {
		line, after, _ := strings.Cut(rest, "\n")
		key, value, ok := strings.Cut(line, ": ")
		switch {
		case ok && key == "Title":
			r.Title = value
		case ok && key == "Link":
			r.URL = value
		case ok && key == "Page":
		case ok && key == "Content":
			r.Description = strings.TrimSpace(value + "\n" + after)
			return r
		default:
			// Not in the expected format: keep the whole text.
			return docResult{Description: text}
		}
		if after == "" {
			return r
		}
		rest = after
	}
}
