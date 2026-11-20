package docs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duneanalytics/cli/cmd/docs"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// searchArgs mirrors the tool input for search_dune_docs.
type searchArgs struct {
	Query            string `json:"query"`
	APIReferenceOnly bool   `json:"apiReferenceOnly,omitempty"`
	CodeOnly         bool   `json:"codeOnly,omitempty"`
}

func startMCPServer(t *testing.T, results []string) *httptest.Server {
	t.Helper()
	var content []mcp.Content
	for _, r := range results {
		content = append(content, &mcp.TextContent{Text: r})
	}
	return startMCPServerResult(t, &mcp.CallToolResult{Content: content})
}

// startMCPServerResult starts a docs server whose search tool returns result.
func startMCPServerResult(t *testing.T, result *mcp.CallToolResult) *httptest.Server {
	t.Helper()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-docs",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_dune_docs",
		Description: "Search Dune docs",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, any, error) {
		return result, nil, nil
	})

	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return ts
}

func newTestRoot() (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{Use: "dune"}
	root.SetContext(context.Background())
	root.AddCommand(docs.NewDocsCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)
	return root, &buf
}

func TestSearchTextOutput(t *testing.T) {
	ts := startMCPServer(t, []string{
		"Title: Decoded Tables\nLink: https://docs.dune.com/decoded\nPage: decoded\nContent: How decoded tables work",
		"Title: Raw Tables\nLink: https://docs.dune.com/raw\nPage: raw\nContent: Raw blockchain data",
	})

	root, buf := newTestRoot()
	root.SetArgs([]string{"docs", "search", "--query", "decoded tables", "--mcp-endpoint", ts.URL})

	require.NoError(t, root.Execute())

	out := buf.String()
	assert.Contains(t, out, "Decoded Tables")
	assert.Contains(t, out, "https://docs.dune.com/decoded")
	assert.Contains(t, out, "Raw Tables")
	assert.Contains(t, out, "2 result(s)")
}

func TestSearchJSONOutput(t *testing.T) {
	ts := startMCPServer(t, []string{
		"Title: API Pagination\nLink: https://docs.dune.com/api/pagination\nPage: api/pagination\n" +
			"Content: ## Pagination\n\nUse `limit` and `offset`.\nLink: not a header here",
	})

	root, buf := newTestRoot()
	root.SetArgs([]string{"docs", "search", "--query", "pagination", "-o", "json", "--mcp-endpoint", ts.URL})

	require.NoError(t, root.Execute())

	var resp struct {
		Query   string `json:"query"`
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))
	assert.Equal(t, "pagination", resp.Query)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, "API Pagination", resp.Results[0].Title)
	assert.Equal(t, "https://docs.dune.com/api/pagination", resp.Results[0].URL)
	// Content runs to the end of the block, even when a content line looks like a header.
	assert.Equal(t, "## Pagination\n\nUse `limit` and `offset`.\nLink: not a header here", resp.Results[0].Description)
}

func TestSearchEmptyResults(t *testing.T) {
	ts := startMCPServer(t, nil)

	root, buf := newTestRoot()
	root.SetArgs([]string{"docs", "search", "--query", "nonexistent topic xyz", "--mcp-endpoint", ts.URL})

	require.NoError(t, root.Execute())
	assert.Contains(t, buf.String(), "No results found.")
}

func TestSearchToolError(t *testing.T) {
	ts := startMCPServerResult(t, &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: "Search index unavailable."}},
	})

	root, buf := newTestRoot()
	root.SetArgs([]string{"docs", "search", "--query", "decoded tables", "--mcp-endpoint", ts.URL})

	assert.EqualError(t, root.Execute(), "searching docs: Search index unavailable")
	assert.NotContains(t, buf.String(), "result(s)")
}

func TestSearchRequiresQuery(t *testing.T) {
	root, _ := newTestRoot()
	root.SetArgs([]string{"docs", "search"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}
