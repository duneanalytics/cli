package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/duneanalytics/cli/authconfig"
	"github.com/duneanalytics/cli/cmd/auth"
	"github.com/duneanalytics/cli/cmd/dataset"
	"github.com/duneanalytics/cli/cmd/docs"
	"github.com/duneanalytics/cli/cmd/execution"
	"github.com/duneanalytics/cli/cmd/query"
	"github.com/duneanalytics/cli/cmd/usage"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/duneapi-client-go/config"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/spf13/cobra"
)

var apiKeyFlag string

var rootCmd = &cobra.Command{
	Use:   "dune",
	Short: "Dune CLI — interact with the Dune Analytics API",
	Long: "A command-line interface for interacting with the Dune Analytics API.\n" +
		"Manage queries, execute them, and retrieve results.",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if cmd.Annotations["skipAuth"] == "true" {
			return nil
		}

		var env *config.Env

		switch {
		case apiKeyFlag != "":
			env = config.FromAPIKey(apiKeyFlag)
		default:
			var err error
			env, err = config.FromEnvVars()
			if err != nil {
				cfg, cfgErr := authconfig.Load()
				if cfgErr != nil {
					return fmt.Errorf("reading auth config: %w", cfgErr)
				}
				if cfg != nil {
					key := strings.TrimSpace(cfg.APIKey)
					if key == "" {
						return fmt.Errorf("empty API key in config: run dune auth --api-key <key>")
					}
					env = config.FromAPIKey(key)
				} else {
					return fmt.Errorf("missing API key: set DUNE_API_KEY, pass --api-key, or run dune auth")
				}
			}
		}

		client := dune.NewDuneClient(env)
		cmdutil.SetClient(cmd, client)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Dune API key (overrides DUNE_API_KEY env var)")
	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(dataset.NewDatasetCmd())
	rootCmd.AddCommand(docs.NewDocsCmd())
	rootCmd.AddCommand(query.NewQueryCmd())
	rootCmd.AddCommand(execution.NewExecutionCmd())
	rootCmd.AddCommand(usage.NewUsageCmd())
}

// Execute runs the root command via Fang.
func Execute(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)

	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
