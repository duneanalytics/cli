package evm

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// Chain-status values returned by the API are lowercase strings.
const protocolStatusPreview = "preview"

// NewSupportedProtocolsCmd returns the `sim evm supported-protocols` command.
func NewSupportedProtocolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supported-protocols",
		Short: "List DeFi protocol families and chains supported by defi-positions",
		Long: "Display DeFi protocol families covered by the Sim defi-positions\n" +
			"endpoint, the chains each family is available on, and the sub-protocols\n" +
			"(forks) recognized under each family. Each chain entry has a status of\n" +
			"Stable or Preview.\n\n" +
			"Use this to discover which protocols and chains 'dune sim evm defi-positions'\n" +
			"can return data for.\n\n" +
			"Examples:\n" +
			"  dune sim evm supported-protocols\n" +
			"  dune sim evm supported-protocols --include-preview-chains\n" +
			"  dune sim evm supported-protocols --include-preview-protocols\n" +
			"  dune sim evm supported-protocols -o json",
		RunE: runSupportedProtocols,
	}

	cmd.Flags().Bool("include-preview-chains", false, "Include chains that are marked as preview (not yet publicly available)")
	cmd.Flags().Bool("include-preview-protocols", false, "Include protocols that are marked as preview on the requested chains")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type supportedProtocolsResponse struct {
	ProtocolFamilies []supportedProtocolFamily `json:"protocol_families"`
}

type supportedProtocolFamily struct {
	Family       string                   `json:"family"`
	Chains       []supportedProtocolChain `json:"chains"`
	SubProtocols []string                 `json:"sub_protocols"`
}

type supportedProtocolChain struct {
	ChainID   json.Number `json:"chain_id"`
	ChainName string      `json:"chain_name"`
	Status    string      `json:"status"`
}

func runSupportedProtocols(cmd *cobra.Command, _ []string) error {
	client := SimClientFromCmd(cmd)
	if client == nil {
		return fmt.Errorf("sim client not initialized")
	}

	params := url.Values{}
	if v, _ := cmd.Flags().GetBool("include-preview-chains"); v {
		params.Set("include_preview_chains", "true")
	}
	if v, _ := cmd.Flags().GetBool("include-preview-protocols"); v {
		params.Set("include_preview_protocols", "true")
	}

	data, err := client.Get(cmd.Context(), "/v1/evm/defi/supported-protocols", params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp supportedProtocolsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		columns := []string{"FAMILY", "CHAINS", "SUB_PROTOCOLS"}
		rows := make([][]string, len(resp.ProtocolFamilies))
		for i, f := range resp.ProtocolFamilies {
			rows[i] = []string{
				f.Family,
				formatProtocolChains(f.Chains),
				strings.Join(f.SubProtocols, ","),
			}
		}
		output.PrintTable(w, columns, rows)
		return nil
	}
}

// formatProtocolChains renders chains as "name(id)" with a "*" suffix for
// preview-status entries, joined by commas.
func formatProtocolChains(chains []supportedProtocolChain) string {
	parts := make([]string, len(chains))
	for i, c := range chains {
		entry := fmt.Sprintf("%s(%s)", c.ChainName, c.ChainID.String())
		if strings.EqualFold(c.Status, protocolStatusPreview) {
			entry += "*"
		}
		parts[i] = entry
	}
	return strings.Join(parts, ",")
}
