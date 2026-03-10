package evm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewSupportedChainsCmd returns the `sim evm supported-chains` command.
func NewSupportedChainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supported-chains",
		Short: "List supported EVM chains and their endpoint availability",
		Long: "Display all EVM chains supported by the Sim API and which endpoints\n" +
			"(balances, activity, transactions, etc.) are available for each chain.\n\n" +
			"This endpoint is public and does not require a Sim API key.\n\n" +
			"Examples:\n" +
			"  dune sim evm supported-chains\n" +
			"  dune sim evm supported-chains -o json",
		Annotations: map[string]string{"skipSimAuth": "true"},
		RunE:        runSupportedChains,
	}

	output.AddFormatFlag(cmd, "text")

	return cmd
}

type supportedChainsResponse struct {
	Chains []chainEntry `json:"chains"`
}

type chainEntry struct {
	Name          string          `json:"name"`
	ChainID       json.Number     `json:"chain_id"`
	Tags          []string        `json:"tags"`
	Balances      endpointSupport `json:"balances"`
	Activity      endpointSupport `json:"activity"`
	Transactions  endpointSupport `json:"transactions"`
	TokenInfo     endpointSupport `json:"token_info"`
	TokenHolders  endpointSupport `json:"token_holders"`
	Collectibles  endpointSupport `json:"collectibles"`
	DefiPositions endpointSupport `json:"defi_positions"`
}

type endpointSupport struct {
	Supported bool `json:"supported"`
}

func runSupportedChains(cmd *cobra.Command, _ []string) error {
	client := SimClientFromCmd(cmd)
	if client == nil {
		return fmt.Errorf("sim client not initialized")
	}

	data, err := client.Get(cmd.Context(), "/v1/evm/supported-chains", nil)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp supportedChainsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		columns := []string{
			"NAME", "CHAIN_ID", "TAGS",
			"BALANCES", "ACTIVITY", "TXS",
			"TOKEN_INFO", "HOLDERS", "COLLECTIBLES", "DEFI",
		}
		rows := make([][]string, len(resp.Chains))
		for i, c := range resp.Chains {
			rows[i] = []string{
				c.Name,
				c.ChainID.String(),
				strings.Join(c.Tags, ","),
				boolYN(c.Balances.Supported),
				boolYN(c.Activity.Supported),
				boolYN(c.Transactions.Supported),
				boolYN(c.TokenInfo.Supported),
				boolYN(c.TokenHolders.Supported),
				boolYN(c.Collectibles.Supported),
				boolYN(c.DefiPositions.Supported),
			}
		}
		output.PrintTable(w, columns, rows)
		return nil
	}
}

func boolYN(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}
