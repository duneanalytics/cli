package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/duneanalytics/cli/cmd/execution"
	"github.com/duneanalytics/cli/cmd/query"
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
		var env *config.Env

		switch {
		case apiKeyFlag != "":
			env = config.FromAPIKey(apiKeyFlag)
		default:
			var err error
			env, err = config.FromEnvVars()
			if err != nil {
				return fmt.Errorf("missing API key: set DUNE_API_KEY or pass --api-key")
			}
		}

		client := dune.NewDuneClient(env)
		cmdutil.SetClient(cmd, client)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Dune API key (overrides DUNE_API_KEY env var)")
	rootCmd.AddCommand(query.NewQueryCmd())
	rootCmd.AddCommand(execution.NewExecutionCmd())
}

// Execute runs the root command via Fang.
func Execute() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
