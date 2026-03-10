package evm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	// This endpoint is public (no auth required). Use the context client
	// if available, otherwise create a bare HTTP client.
	client := SimClientFromCmd(cmd)
	if client == nil {
		client = newBareSimClient()
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

// bareSimClient is a minimal HTTP client for public Sim API endpoints
// that don't require authentication.
type bareSimClient struct {
	httpClient *http.Client
}

func newBareSimClient() *bareSimClient {
	return &bareSimClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *bareSimClient) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u, err := url.Parse("https://api.sim.dune.com" + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Sim API error (HTTP %d)", resp.StatusCode)
	}

	return body, nil
}
