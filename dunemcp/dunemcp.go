// Package dunemcp calls tools on the Dune MCP server with the CLI's API key.
//
// The CLI uses it for capabilities that the Dune MCP server offers and the REST
// API does not, so those commands behave exactly like the MCP tools.
package dunemcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Path is where the Dune API host serves the MCP server.
const Path = "/mcp/v1"

// Caller calls tools on the Dune MCP server.
type Caller interface {
	// Call runs a tool and returns its structured result as JSON.
	Call(ctx context.Context, tool string, args map[string]any) (json.RawMessage, error)
	// CallWithImages runs a tool like Call and also returns the images in its
	// result, decoded.
	CallWithImages(ctx context.Context, tool string, args map[string]any) (json.RawMessage, [][]byte, error)
}

// Client is a Caller that opens one MCP session per call.
type Client struct {
	endpoint string
	headers  map[string]string
	version  string
}

// New returns a Client for the Dune API at host. The API key and headers are
// sent on every request.
func New(host, apiKey string, headers map[string]string, version string) *Client {
	h := map[string]string{"X-Dune-Api-Key": apiKey}
	for k, v := range headers {
		h[k] = v
	}
	return &Client{
		endpoint: strings.TrimRight(host, "/") + Path,
		headers:  h,
		version:  version,
	}
}

// Call runs tool with args. A tool that reports an error returns it as a Go error
// carrying the server's message.
func (c *Client) Call(ctx context.Context, tool string, args map[string]any) (json.RawMessage, error) {
	raw, _, err := c.CallWithImages(ctx, tool, args)
	return raw, err
}

// CallWithImages runs tool with args like Call and also returns the images the
// tool sent.
func (c *Client) CallWithImages(ctx context.Context, tool string, args map[string]any) (json.RawMessage, [][]byte, error) {
	session, err := c.connect(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", tool, err)
	}
	raw, err := Result(tool, res)
	if err != nil {
		return nil, nil, err
	}
	return raw, Images(res), nil
}

func (c *Client) connect(ctx context.Context) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "dune-cli", Version: c.version}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             c.endpoint,
		HTTPClient:           &http.Client{Transport: headerTransport{headers: c.headers}},
		DisableStandaloneSSE: true,
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to the Dune MCP server at %s: %w", c.endpoint, err)
	}
	return session, nil
}

// Result returns the structured content of a tool result as JSON, or the error
// the tool reported. A result without structured content falls back to its
// first text block, which Dune tools fill with the same JSON.
func Result(tool string, res *mcp.CallToolResult) (json.RawMessage, error) {
	if err := ToolError(tool, res); err != nil {
		return nil, err
	}
	if res.StructuredContent != nil {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false) // keep & in chart URLs as it is
		if err := enc.Encode(res.StructuredContent); err != nil {
			return nil, err
		}
		return bytes.TrimRight(buf.Bytes(), "\n"), nil
	}
	if t := text(res); json.Valid([]byte(t)) {
		return json.RawMessage(t), nil
	}
	return nil, errors.New(tool + ": the tool returned no structured result")
}

// ToolError returns the error a tool result reports, or nil when it reports none.
func ToolError(tool string, res *mcp.CallToolResult) error {
	if !res.IsError {
		return nil
	}
	// Go error strings end without punctuation; the CLI's error printer adds a period.
	msg := strings.TrimSuffix(strings.TrimSpace(text(res)), ".")
	if msg == "" {
		msg = tool + " failed without a message"
	}
	return errors.New(msg)
}

// Images returns the decoded data of the image blocks in a tool result.
func Images(res *mcp.CallToolResult) [][]byte {
	var images [][]byte
	for _, c := range res.Content {
		if img, ok := c.(*mcp.ImageContent); ok {
			images = append(images, img.Data)
		}
	}
	return images
}

func text(res *mcp.CallToolResult) string {
	var parts []string
	for _, c := range res.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, t.Text)
		}
	}
	return strings.Join(parts, "\n")
}

type headerTransport struct {
	headers map[string]string
}

func (t headerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	for k, v := range t.headers {
		r.Header.Set(k, v)
	}
	return http.DefaultTransport.RoundTrip(r)
}
