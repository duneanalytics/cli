package evm

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/output"
)

// NewCollectiblesCmd returns the `sim evm collectibles` command.
func NewCollectiblesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collectibles <address>",
		Short: "Get NFT collectibles (ERC721/ERC1155) held by a wallet address",
		Long: "Return ERC721 and ERC1155 collectibles (NFTs) held by the given wallet\n" +
			"address across supported EVM chains. Results include token metadata, images,\n" +
			"and acquisition timestamps. Spam filtering is enabled by default to hide\n" +
			"airdropped junk NFTs.\n\n" +
			"Response fields per collectible:\n" +
			"  - contract_address: NFT collection contract\n" +
			"  - token_id: unique token identifier within the collection\n" +
			"  - token_standard: 'erc721' or 'erc1155'\n" +
			"  - chain, chain_id: network name and numeric ID\n" +
			"  - name, symbol, description: collection metadata\n" +
			"  - image_url: token image (may be IPFS, HTTP, or data URI)\n" +
			"  - balance: quantity held (always '1' for ERC721, may be >1 for ERC1155)\n" +
			"  - last_acquired: timestamp of most recent acquisition\n" +
			"  - is_spam, spam_score: spam classification (visible with --show-spam-scores)\n\n" +
			"Spam filtering uses a scoring model based on collection traits (holder count,\n" +
			"transfer patterns, metadata quality). Disable with --filter-spam=false to see\n" +
			"all NFTs including suspected spam.\n\n" +
			"By default, queries all chains tagged 'default'. Run 'dune sim evm\n" +
			"supported-chains' to see which chains support collectibles.\n\n" +
			"Examples:\n" +
			"  dune sim evm collectibles 0xd8da6bf26964af9d7eed9e03e53415d37aa96045\n" +
			"  dune sim evm collectibles 0xd8da... --chain-ids 1\n" +
			"  dune sim evm collectibles 0xd8da... --filter-spam=false --show-spam-scores -o json\n" +
			"  dune sim evm collectibles 0xd8da... --limit 100 -o json",
		Args: cobra.ExactArgs(1),
		RunE: runCollectibles,
	}

	cmd.Flags().String("chain-ids", "", "Restrict to specific chains by numeric ID or tag name (comma-separated, e.g. '1,8453' or 'default'); defaults to all chains tagged 'default'")
	cmd.Flags().Bool("filter-spam", true, "Hide collectibles identified as spam by the scoring model (default: true); set --filter-spam=false to include all NFTs")
	cmd.Flags().Bool("show-spam-scores", false, "Include spam classification details in the response: is_spam flag, numeric spam_score, and per-feature explanations with weights")
	cmd.Flags().Int("limit", 0, "Maximum number of collectibles to return per page (1-2500, default: 250)")
	cmd.Flags().String("offset", "", "Pagination cursor returned as next_offset in a previous response; use to fetch the next page of results")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type collectiblesResponse struct {
	Address      string             `json:"address"`
	Entries      []collectibleEntry `json:"entries"`
	Warnings     []warningEntry     `json:"warnings,omitempty"`
	NextOffset   string             `json:"next_offset,omitempty"`
	RequestTime  string             `json:"request_time,omitempty"`
	ResponseTime string             `json:"response_time,omitempty"`
}

type collectibleEntry struct {
	ContractAddress string               `json:"contract_address"`
	TokenStandard   string               `json:"token_standard"`
	TokenID         string               `json:"token_id"`
	Chain           string               `json:"chain"`
	ChainID         int64                `json:"chain_id"`
	Name            string               `json:"name,omitempty"`
	Symbol          string               `json:"symbol,omitempty"`
	Description     string               `json:"description,omitempty"`
	ImageURL        string               `json:"image_url,omitempty"`
	LastSalePrice   string               `json:"last_sale_price,omitempty"`
	Metadata        *collectibleMetadata `json:"metadata,omitempty"`
	Balance         string               `json:"balance"`
	LastAcquired    string               `json:"last_acquired"`
	IsSpam          bool                 `json:"is_spam"`
	SpamScore       int                  `json:"spam_score,omitempty"`
	Explanations    []spamExplanation    `json:"explanations,omitempty"`
}

type collectibleMetadata struct {
	URI        string                 `json:"uri,omitempty"`
	Attributes []collectibleAttribute `json:"attributes,omitempty"`
}

type collectibleAttribute struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Format string `json:"format,omitempty"`
}

type spamExplanation struct {
	Feature       string          `json:"feature"`
	Value         json.RawMessage `json:"value,omitempty"`
	FeatureScore  int             `json:"feature_score,omitempty"`
	FeatureWeight float64         `json:"feature_weight,omitempty"`
}

func runCollectibles(cmd *cobra.Command, args []string) error {
	client := SimClientFromCmd(cmd)

	address := args[0]
	params := url.Values{}

	if v, _ := cmd.Flags().GetString("chain-ids"); v != "" {
		params.Set("chain_ids", v)
	}
	// filter_spam defaults to true on the API side, so only send when explicitly false.
	if v, _ := cmd.Flags().GetBool("filter-spam"); !v {
		params.Set("filter_spam", "false")
	}
	if v, _ := cmd.Flags().GetBool("show-spam-scores"); v {
		params.Set("show_spam_scores", "true")
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, _ := cmd.Flags().GetString("offset"); v != "" {
		params.Set("offset", v)
	}

	data, err := client.Get(cmd.Context(), "/v1/evm/collectibles/"+address, params)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	switch output.FormatFromCmd(cmd) {
	case output.FormatJSON:
		var raw json.RawMessage = data
		return output.PrintJSON(w, raw)
	default:
		var resp collectiblesResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		// Print warnings to stderr.
		printWarnings(cmd, resp.Warnings)

		showSpam, _ := cmd.Flags().GetBool("show-spam-scores")

		columns := []string{"CHAIN", "NAME", "SYMBOL", "TOKEN_ID", "STANDARD", "BALANCE"}
		if showSpam {
			columns = append(columns, "SPAM", "SPAM_SCORE")
		}
		rows := make([][]string, len(resp.Entries))
		for i, e := range resp.Entries {
			row := []string{
				e.Chain,
				e.Name,
				e.Symbol,
				e.TokenID,
				e.TokenStandard,
				e.Balance,
			}
			if showSpam {
				spam := "N"
				if e.IsSpam {
					spam = "Y"
				}
				row = append(row, spam, fmt.Sprintf("%d", e.SpamScore))
			}
			rows[i] = row
		}
		output.PrintTable(w, columns, rows)

		if resp.NextOffset != "" {
			fmt.Fprintf(w, "\nNext offset: %s\n", resp.NextOffset)
		}
		return nil
	}
}
