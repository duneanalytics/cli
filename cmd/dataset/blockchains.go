package dataset

import (
	"fmt"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newBlockchainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blockchains",
		Short: "List indexed blockchains with their table counts",
		Long: "List the blockchain names in the Dune catalog with how many tables each has,\n" +
			"optionally narrowed to a category or a topic. Use it before 'dune dataset search\n" +
			"--blockchains' when you are unsure of a chain's name, or to see which chains cover\n" +
			"a topic.\n\n" +
			"Examples:\n" +
			"  dune dataset blockchains\n" +
			"  dune dataset blockchains --categories decoded\n" +
			"  dune dataset blockchains --query lending --limit 10 -o json",
		Args: cobra.NoArgs,
		RunE: runBlockchains,
	}
	cmd.Flags().StringArray("categories", nil,
		"count only tables in these categories: canonical, decoded, spell, community")
	cmd.Flags().String("query", "", "count only tables matching this topic, e.g. 'lending'; '*' counts all")
	cmd.Flags().Int("limit", 200, "maximum number of blockchains to return")
	cmd.Flags().Int("offset", 0, "number of blockchains to skip, for the next page")
	output.AddFormatFlag(cmd, "text")
	return cmd
}

func runBlockchains(cmd *cobra.Command, _ []string) error {
	args := cmdutil.ToolArgs(cmd, map[string]string{
		"categories": "categories",
		"query":      "query",
		"limit":      "limit",
		"offset":     "offset",
	})
	var res struct {
		Total   int `json:"total"`
		Results []struct {
			Blockchain string `json:"blockchain"`
			TableCount int    `json:"table_count"`
		} `json:"results"`
	}
	raw, err := cmdutil.CallTool(cmd, "listBlockchains", args, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(cmd, raw)
	}
	rows := make([][]string, len(res.Results))
	for i, r := range res.Results {
		rows[i] = []string{r.Blockchain, fmt.Sprintf("%d", r.TableCount)}
	}
	output.PrintTable(w, []string{"BLOCKCHAIN", "TABLES"}, rows)
	fmt.Fprintf(w, "\n%d of %d blockchains\n", len(res.Results), res.Total)
	return nil
}
