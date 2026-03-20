package sim

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/duneapi-client-go/config"
	"github.com/duneanalytics/duneapi-client-go/dune"

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

	// Resolve customer identity for analytics (best-effort, never blocks the CLI).
	// We try to resolve the Dune API key (not the Sim API key) to identify the user.
	if tr := cmdutil.TrackerFromCmd(cmd); tr != nil {
		if env := resolveDuneEnv(); env != nil {
			client := dune.NewDuneClient(env)
			if customerID := resolveCustomerIDForSim(client, env.APIKey); customerID != "" {
				tr.SetUserID(customerID)
			}
		}
	}

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

// resolveDuneEnv attempts to resolve the Dune API key from environment variables
// or the config file, returning a Dune environment configuration.
// Returns nil if no Dune API key is available (this is not an error for sim commands).
func resolveDuneEnv() *config.Env {
	// Try environment variable first.
	env, err := config.FromEnvVars()
	if err == nil {
		return env
	}

	// Try config file.
	cfg, err := authconfig.Load()
	if err != nil || cfg == nil {
		return nil
	}

	key := strings.TrimSpace(cfg.APIKey)
	if key == "" {
		return nil
	}

	return config.FromAPIKey(key)
}

// resolveCustomerIDForSim is a wrapper around cli.ResolveCustomerID that
// can be called from the sim package. It must be declared here to avoid
// a circular import between cli and cmd/sim.
func resolveCustomerIDForSim(client dune.DuneClient, apiKey string) string {
	keyHash := authconfig.HashAPIKey(apiKey)

	// Try the cache first.
	cached, err := authconfig.LoadIdentity()
	if err == nil && cached != nil && cached.APIKeyHash == keyHash && cached.CustomerID != "" {
		return cached.CustomerID
	}

	// Cache miss or stale — call the API.
	resp, err := client.WhoAmI()
	if err != nil || resp == nil || resp.CustomerID == "" {
		return ""
	}

	// Persist for next time (best-effort).
	_ = authconfig.SaveIdentity(&authconfig.UserIdentity{
		CustomerID: resp.CustomerID,
		APIKeyHash: keyHash,
	})

	return resp.CustomerID
}
