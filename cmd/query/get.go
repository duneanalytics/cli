package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <query-id>",
		Short: "Get a saved query",
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	queryID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid query ID %q: must be an integer", args[0])
	}

	client := cmdutil.ClientFromCmd(cmd)

	resp, err := client.GetQuery(queryID)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		return output.PrintJSON(w, resp)
	default:
		fmt.Fprintf(w, "ID:          %d\n", resp.QueryID)
		fmt.Fprintf(w, "Name:        %s\n", resp.Name)
		if resp.Description != "" {
			fmt.Fprintf(w, "Description: %s\n", resp.Description)
		}
		fmt.Fprintf(w, "Owner:       %s\n", resp.Owner)
		fmt.Fprintf(w, "Engine:      %s\n", resp.QueryEngine)
		fmt.Fprintf(w, "Version:     %d\n", resp.Version)
		fmt.Fprintf(w, "Private:     %t\n", resp.IsPrivate)
		fmt.Fprintf(w, "Archived:    %t\n", resp.IsArchived)
		if len(resp.Tags) > 0 {
			fmt.Fprintf(w, "Tags:        %s\n", strings.Join(resp.Tags, ", "))
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "SQL:")
		for _, line := range strings.Split(resp.QuerySQL, "\n") {
			fmt.Fprintf(w, "  %s\n", line)
		}
		return nil
	}
}
