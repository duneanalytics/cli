package sim

import (
	"fmt"
	"os"
	"strings"
	"time"

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
		Short: "Query real-time blockchain data via the Dune Sim API",
		Long: "Access real-time, indexed blockchain data through the Dune Sim API. Unlike\n" +
			"'dune query run' which executes SQL against historical data warehouses, Sim API\n" +
			"endpoints return pre-indexed, low-latency responses for common wallet and token\n" +
			"lookups.\n\n" +
			"Available subcommands:\n" +
			"  evm  - Query EVM chains: balances, activity, transactions, collectibles,\n" +
			"         token-info, token-holders, defi-positions, supported-chains\n" +
			"  svm  - Query SVM chains (Solana, Eclipse): balances, transactions\n" +
			"  auth - Save your Sim API key to the local config file\n\n" +
			"Authentication:\n" +
			"  Most commands require a Sim API key. The key is resolved in priority order:\n" +
			"  1. --sim-api-key flag\n" +
			"  2. DUNE_SIM_API_KEY environment variable\n" +
			"  3. Saved key in ~/.config/dune/config.yaml (set via 'dune sim auth')\n\n" +
			"  Exception: 'dune sim evm supported-chains' is a public endpoint and does\n" +
			"  not require authentication.\n\n" +
			"Each API call consumes compute units based on the number of chains queried\n" +
			"and the complexity of the request. Use -o json on any command to get the\n" +
			"full structured response for programmatic consumption.",
		Annotations:       map[string]string{"skipAuth": "true"},
		PersistentPreRunE: simPreRun,
	}

	cmd.PersistentFlags().StringVar(
		&simAPIKeyFlag, "sim-api-key", "",
		"Sim API key for authentication (overrides DUNE_SIM_API_KEY env var and saved config); keys are prefixed 'sim_'",
	)

	cmd.AddCommand(NewAuthCmd())
	cmd.AddCommand(evm.NewEvmCmd())
	cmd.AddCommand(svm.NewSvmCmd())

	return cmd
}

// simPreRun resolves the Sim API key and stores a SimClient in the command context.
// Commands annotated with "skipSimAuth": "true" bypass this step.
func simPreRun(cmd *cobra.Command, _ []string) error {
	// The sim command's PersistentPreRunE overrides the root command's hook
	// (cobra does not chain PersistentPreRunE without EnableTraverseRunHooks).
	// Record the start time here so the root's PersistentPostRunE computes a
	// correct duration for telemetry.
	cmdutil.SetStartTime(cmd, time.Now())

	// Commands like `sim evm supported-chains` that hit public endpoints
	// don't require an API key. Provide a bare (unauthenticated) client so
	// they can still use the shared HTTP infrastructure and error handling.
	if cmd.Annotations["skipSimAuth"] == "true" {
		cmdutil.SetSimClient(cmd, NewBareSimClient())
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
