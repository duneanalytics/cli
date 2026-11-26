// Package dunemcptest provides a fake Dune MCP server for command tests.
package dunemcptest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/spf13/cobra"
)

// Fake is a dunemcp.Caller that returns canned tool results and records calls.
type Fake struct {
	// Results maps a tool name to the JSON result it returns.
	Results map[string]string
	// Images maps a tool name to the images it returns with its result.
	Images map[string][][]byte
	// Errors maps a tool name to the error it returns.
	Errors map[string]error
	Calls  []Call
}

// Call is one recorded tool call, with its arguments as they go on the wire.
type Call struct {
	Tool string
	Args string
}

// Call implements dunemcp.Caller.
func (f *Fake) Call(_ context.Context, tool string, args map[string]any) (json.RawMessage, error) {
	b, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	f.Calls = append(f.Calls, Call{Tool: tool, Args: string(b)})
	if err := f.Errors[tool]; err != nil {
		return nil, err
	}
	res, ok := f.Results[tool]
	if !ok {
		return nil, fmt.Errorf("unexpected call to %s", tool)
	}
	return json.RawMessage(res), nil
}

// CallWithImages implements dunemcp.Caller.
func (f *Fake) CallWithImages(ctx context.Context, tool string, args map[string]any) (json.RawMessage, [][]byte, error) {
	raw, err := f.Call(ctx, tool, args)
	if err != nil {
		return nil, nil, err
	}
	return raw, f.Images[tool], nil
}

// Last returns the most recent call. It panics when there was none.
func (f *Fake) Last() Call { return f.Calls[len(f.Calls)-1] }

// NewRoot returns a root command holding cmd, with fake as its Dune MCP server,
// and the buffer that collects its output.
func NewRoot(cmd *cobra.Command, fake *Fake) (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{
		Use: "dune",
		PersistentPreRun: func(c *cobra.Command, _ []string) {
			cmdutil.SetMCP(c, fake)
		},
	}
	root.SetContext(context.Background())
	root.AddCommand(cmd)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&bytes.Buffer{})
	return root, &buf
}
