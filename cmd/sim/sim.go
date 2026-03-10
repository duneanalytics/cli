package sim

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/duneanalytics/cli/cmd/sim/evm"
	"github.com/duneanalytics/cli/cmd/sim/svm"
	"github.com/duneanalytics/cli/cmdutil"
)

var simAPIKeyFlag string

// NewSimCmd returns the `sim` parent command.
func NewSimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sim",
		Short: "Query real-time blockchain data via the Sim API",
		Long: "Access real-time blockchain data including balances, activity, transactions,\n" +
			"collectibles, token info, token holders, and DeFi positions across EVM and SVM chains.\n\n" +
			"Authenticate with a Sim API key via --sim-api-key, the DUNE_SIM_API_KEY environment\n" +
			"variable, or by running `dune sim auth`.",
		Annotations:       map[string]string{"skipAuth": "true"},
		PersistentPreRunE: simPreRun,
	}

	cmd.PersistentFlags().StringVar(
		&simAPIKeyFlag, "sim-api-key", "",
		"Sim API key (overrides DUNE_SIM_API_KEY env var)",
	)

	cmd.AddCommand(evm.NewEvmCmd())
	cmd.AddCommand(svm.NewSvmCmd())

	return cmd
}

// simPreRun resolves the Sim API key and stores a SimClient in the command context.
// Commands annotated with "skipSimAuth": "true" bypass this step.
func simPreRun(cmd *cobra.Command, _ []string) error {
	// Allow commands like `sim auth` to skip sim client creation.
	if cmd.Annotations["skipSimAuth"] == "true" {
		return nil
	}

	apiKey := resolveSimAPIKey()
	if apiKey == "" {
		return fmt.Errorf(
			"missing Sim API key: set DUNE_SIM_API_KEY, pass --sim-api-key, or run `dune sim auth`",
		)
	}

	client := NewSimClient(apiKey)
	cmdutil.SetSimClient(cmd, client)

	return nil
}

// resolveSimAPIKey resolves the Sim API key from (in priority order):
// 1. --sim-api-key flag
// 2. DUNE_SIM_API_KEY environment variable
// 3. sim_api_key from ~/.config/dune/config.yaml
func resolveSimAPIKey() string {
	// 1. Flag
	if simAPIKeyFlag != "" {
		return strings.TrimSpace(simAPIKeyFlag)
	}

	// 2. Environment variable
	if key := os.Getenv("DUNE_SIM_API_KEY"); key != "" {
		return strings.TrimSpace(key)
	}

	// 3. Config file
	cfg, err := authconfig.Load()
	if err != nil || cfg == nil {
		return ""
	}
	return strings.TrimSpace(cfg.SimAPIKey)
}

// SimClientFromCmd is a convenience helper that extracts and type-asserts the
// SimClient from the command context.
func SimClientFromCmd(cmd *cobra.Command) *SimClient {
	v := cmdutil.SimClientFromCmd(cmd)
	if v == nil {
		return nil
	}
	return v.(*SimClient)
}
