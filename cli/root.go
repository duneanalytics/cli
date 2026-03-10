package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/duneanalytics/duneapi-client-go/config"
	"github.com/duneanalytics/duneapi-client-go/dune"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/duneanalytics/cli/cmd/auth"
	duneconfig "github.com/duneanalytics/cli/cmd/config"
	"github.com/duneanalytics/cli/cmd/dataset"
	"github.com/duneanalytics/cli/cmd/docs"
	"github.com/duneanalytics/cli/cmd/execution"
	"github.com/duneanalytics/cli/cmd/query"
	"github.com/duneanalytics/cli/cmd/usage"
	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/tracking"
)

var apiKeyFlag string

var rootCmd = &cobra.Command{
	Use:   "dune",
	Short: "Dune CLI — query, explore, and manage blockchain data on Dune Analytics",
	Long: "A command-line interface for the Dune Analytics platform.\n\n" +
		"Discover datasets across the Dune catalog, execute SQL queries (DuneSQL dialect),\n" +
		"retrieve execution results, and manage your saved queries — all from the terminal.\n\n" +
		"Capabilities:\n" +
		"  - Search datasets by keyword, contract address, category, or blockchain\n" +
		"  - Create, update, archive, and retrieve saved DuneSQL queries\n" +
		"  - Execute saved queries or raw DuneSQL and display results\n" +
		"  - Browse Dune documentation for DuneSQL syntax, API references, and guides\n" +
		"  - Monitor credit usage, storage consumption, and billing periods\n\n" +
		"Authenticate with an API key via --api-key, the DUNE_API_KEY environment variable,\n" +
		"or by running `dune auth`.",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		cmdutil.SetStartTime(cmd, time.Now())

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
	PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
		tr := cmdutil.TrackerFromCmd(cmd)
		if tr == nil {
			return nil
		}
		start := cmdutil.StartTimeFromCmd(cmd)
		durationMs := time.Since(start).Milliseconds()

		commandPath := cmd.CommandPath()
		// Strip the root command name for cleaner paths.
		if parts := strings.SplitN(commandPath, " ", 2); len(parts) == 2 {
			commandPath = parts[1]
		}

		tr.Track(commandPath, tracking.StatusSuccess, "", durationMs)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Dune API key (overrides DUNE_API_KEY env var)")
	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(duneconfig.NewConfigCmd())
	rootCmd.AddCommand(dataset.NewDatasetCmd())
	rootCmd.AddCommand(docs.NewDocsCmd())
	rootCmd.AddCommand(query.NewQueryCmd())
	rootCmd.AddCommand(execution.NewExecutionCmd())
	rootCmd.AddCommand(usage.NewUsageCmd())
}

// Execute runs the root command via Fang.
func Execute(version, commit, date, amplitudeKey string) {
	versionStr := fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)

	telemetryEnabled := duneconfig.IsTelemetryEnabled()
	configDir, _ := authconfig.Dir()
	tracker := tracking.New(tracking.Config{
		AmplitudeKey: amplitudeKey,
		CLIVersion:   version,
		ConfigDir:    configDir,
		Enabled:      telemetryEnabled,
	})
	defer tracker.Shutdown()

	rootCmd.SetContext(context.Background())
	cmdutil.SetTracker(rootCmd, tracker)

	if err := fang.Execute(rootCmd.Context(), rootCmd,
		fang.WithVersion(versionStr),
	); err != nil {
		// Build best-effort command path from os.Args (strip flags).
		commandPath := commandPathFromArgs(os.Args)
		tracker.Track(commandPath, tracking.StatusError, err.Error(), 0)
		// Flush the event before exiting — os.Exit does not run deferred funcs,
		// so defer tracker.Shutdown() above would never fire.
		tracker.Shutdown()

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// commandPathFromArgs extracts the subcommand path from os.Args, stripping
// the binary name and any flags so the tracked path is e.g. "query list"
// rather than "query list --limit 10".
func commandPathFromArgs(args []string) string {
	var parts []string
	for _, a := range args[1:] { // skip binary name
		if strings.HasPrefix(a, "-") {
			break
		}
		parts = append(parts, a)
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, " ")
}