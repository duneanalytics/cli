package dataset

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newSizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "size <table>",
		Short: "Show the storage size of a table and the tables under it",
		Long: "Show the total storage size of a Dune table or view and of each physical table\n" +
			"under it. Views and spells such as dex.trades can sit on many tables; their sizes\n" +
			"are added up. Use it to estimate how much a full scan reads.\n\n" +
			"It runs an EXPLAIN on the query engine to find the underlying tables, so each call\n" +
			"starts a new EXPLAIN execution; it changes no data. It does not return row counts\n" +
			"or columns; use 'dune dataset search --include-schema' for those.\n\n" +
			"Examples:\n" +
			"  dune dataset size dex.trades\n" +
			"  dune dataset size dune.my_team.result_token_summary -o json",
		Args: cobra.ExactArgs(1),
		RunE: runSize,
	}
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runSize(cmd *cobra.Command, args []string) error {
	var res struct {
		TableName        string `json:"table_name"`
		TotalSizeBytes   *int64 `json:"total_size_bytes"`
		UnderlyingTables []struct {
			TableName string `json:"table_name"`
			SizeBytes *int64 `json:"size_bytes"`
			Error     string `json:"error"`
		} `json:"underlying_tables"`
		Warnings []string `json:"warnings"`
	}
	raw, err := cmdutil.CallTool(cmd, "getTableSize", map[string]any{"table_name": args[0]}, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(cmd, raw)
	}
	if res.TotalSizeBytes != nil {
		fmt.Fprintf(w, "%s: %s\n", res.TableName, output.FormatBytes(*res.TotalSizeBytes))
	} else {
		fmt.Fprintf(w, "%s: size unknown\n", res.TableName)
	}
	if len(res.UnderlyingTables) > 0 {
		fmt.Fprintln(w)
		rows := make([][]string, len(res.UnderlyingTables))
		for i, t := range res.UnderlyingTables {
			size := t.Error
			if t.SizeBytes != nil {
				size = output.FormatBytes(*t.SizeBytes)
			}
			rows[i] = []string{t.TableName, size}
		}
		output.PrintTable(w, []string{"UNDERLYING_TABLE", "SIZE"}, rows)
	}
	for _, warning := range res.Warnings {
		fmt.Fprintf(w, "\nWarning: %s\n", warning)
	}
	return nil
}
