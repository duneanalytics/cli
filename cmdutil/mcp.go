package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/duneanalytics/cli/dunemcp"
	"github.com/spf13/cobra"
)

type mcpKey struct{}

// SetMCP stores a Dune MCP caller in the command's context.
func SetMCP(cmd *cobra.Command, c dunemcp.Caller) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, mcpKey{}, c))
}

// MCPFromCmd extracts the Dune MCP caller stored in the command's context.
func MCPFromCmd(cmd *cobra.Command) dunemcp.Caller {
	return cmd.Context().Value(mcpKey{}).(dunemcp.Caller)
}

// ToolArgs returns Dune MCP tool arguments for the flags the user set, so the
// server applies its own defaults to the rest. names maps flag names to the
// argument each one fills.
func ToolArgs(cmd *cobra.Command, names map[string]string) map[string]any {
	fs := cmd.Flags()
	args := map[string]any{}
	for name, arg := range names {
		if !fs.Changed(name) {
			continue
		}
		var v any
		switch fs.Lookup(name).Value.Type() {
		case "bool":
			v, _ = fs.GetBool(name)
		case "int":
			v, _ = fs.GetInt(name)
		case "int32":
			v, _ = fs.GetInt32(name)
		case "stringSlice":
			v, _ = fs.GetStringSlice(name)
		case "stringArray":
			v, _ = fs.GetStringArray(name)
		case "int64Slice":
			v, _ = fs.GetInt64Slice(name)
		default:
			v, _ = fs.GetString(name)
		}
		args[arg] = v
	}
	return args
}

// JSONFlag decodes a flag that holds JSON. A value that starts with @ names a
// file to read the JSON from, and @- reads standard input.
func JSONFlag(cmd *cobra.Command, name string, v any) error {
	s, _ := cmd.Flags().GetString(name)
	data := []byte(s)
	if path, ok := strings.CutPrefix(s, "@"); ok {
		var err error
		if path == "-" {
			data, err = io.ReadAll(cmd.InOrStdin())
		} else {
			data, err = os.ReadFile(path)
		}
		if err != nil {
			return fmt.Errorf("reading --%s: %w", name, err)
		}
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("invalid --%s JSON: %w", name, err)
	}
	return nil
}

// CallTool calls a Dune MCP tool and, when view is not nil, decodes the result
// into it. It returns the result as the tool sent it, for JSON output.
func CallTool(cmd *cobra.Command, tool string, args map[string]any, view any) (json.RawMessage, error) {
	raw, err := MCPFromCmd(cmd).Call(cmd.Context(), tool, args)
	if err != nil {
		return nil, err
	}
	if view != nil {
		if err := json.Unmarshal(raw, view); err != nil {
			return nil, fmt.Errorf("%s: unexpected result: %w", tool, err)
		}
	}
	return raw, nil
}

// ParseID parses a numeric query, dashboard, or visualization ID given as a
// positional argument.
func ParseID(kind, arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s ID %q: must be a positive integer", kind, arg)
	}
	return id, nil
}
