package dunemcp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/duneanalytics/cli/dunemcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type queryArgs struct {
	QueryID int `json:"query_id"`
}

type queryOut struct {
	QueryID int    `json:"query_id"`
	Name    string `json:"name"`
}

// TestClientCallsTheDuneMCPServer runs the client against a real MCP server over
// streamable HTTP, as the CLI does against api.dune.com.
func TestClientCallsTheDuneMCPServer(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "dune", Version: "test"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "getDuneQuery"},
		func(_ context.Context, _ *mcp.CallToolRequest, in queryArgs) (*mcp.CallToolResult, queryOut, error) {
			return nil, queryOut{QueryID: in.QueryID, Name: "Daily volume"}, nil
		})
	mcp.AddTool(server, &mcp.Tool{Name: "getDuneFolder"},
		func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
			res := &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: "Folder not found"}}}
			return res, nil, nil
		})
	mcp.AddTool(server, &mcp.Tool{Name: "getDashboardImage"},
		func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, queryOut, error) {
			res := &mcp.CallToolResult{Content: []mcp.Content{&mcp.ImageContent{Data: []byte{0xff, 0xd8, 0x00}, MIMEType: "image/jpeg"}}}
			return res, queryOut{Name: "dash"}, nil
		})

	var mu sync.Mutex
	var paths, keys, origins []string
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		keys = append(keys, r.Header.Get("X-Dune-Api-Key"))
		origins = append(origins, r.Header.Get("Dune-Origin"))
		mu.Unlock()
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(ts.Close)

	client := dunemcp.New(ts.URL+"/", "secret-key", map[string]string{"Dune-Origin": "CLI"}, "test")

	raw, err := client.Call(context.Background(), "getDuneQuery", map[string]any{"query_id": 42})
	require.NoError(t, err)
	assert.JSONEq(t, `{"query_id":42,"name":"Daily volume"}`, string(raw))

	_, err = client.Call(context.Background(), "getDuneFolder", map[string]any{})
	assert.EqualError(t, err, "Folder not found")

	// Images arrive base64-encoded and come back as their raw bytes.
	raw, images, err := client.CallWithImages(context.Background(), "getDashboardImage", map[string]any{})
	require.NoError(t, err)
	assert.JSONEq(t, `{"query_id":0,"name":"dash"}`, string(raw))
	assert.Equal(t, [][]byte{{0xff, 0xd8, 0x00}}, images)

	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, paths)
	for i := range paths {
		assert.Equal(t, "/mcp/v1", paths[i])
		assert.Equal(t, "secret-key", keys[i])
		assert.Equal(t, "CLI", origins[i])
	}
}

func TestResultFallsBackToJSONText(t *testing.T) {
	raw, err := dunemcp.Result("t", &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: `{"a":1}`}}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"a":1}`, string(raw))

	_, err = dunemcp.Result("t", &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "not json"}}})
	assert.EqualError(t, err, "t: the tool returned no structured result")

	_, err = dunemcp.Result("t", &mcp.CallToolResult{IsError: true})
	assert.EqualError(t, err, "t failed without a message")

	_, err = dunemcp.Result("t", &mcp.CallToolResult{IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: "No folder named \"x\" for this account.\n"}}})
	assert.EqualError(t, err, `No folder named "x" for this account`)
}

func TestResultKeepsURLsReadable(t *testing.T) {
	raw, err := dunemcp.Result("t", &mcp.CallToolResult{
		StructuredContent: map[string]any{"chartJpegUrl": "https://x/shot?url=y&cache=0"}})
	require.NoError(t, err)
	assert.Equal(t, `{"chartJpegUrl":"https://x/shot?url=y&cache=0"}`, string(raw))
}
