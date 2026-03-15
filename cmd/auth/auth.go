package auth

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/duneanalytics/cli/output"
	"github.com/duneanalytics/duneapi-client-go/config"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/duneanalytics/duneapi-client-go/models"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	queryStateCompleted = "QUERY_STATE_COMPLETED"

	// sampleQueryMaxRetries is the error-retry ceiling for the sample query poll loop.
	// Matches the CLI default: 300s timeout / 2s poll interval = 150.
	sampleQueryMaxRetries = 150
)

// NewAuthCmd returns the `auth` command.
func NewAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "auth",
		Short:       "Authenticate with the Dune API",
		Long:        "Save your Dune API key to ~/.config/dune/config.yaml so you don't need to pass it every time.",
		Annotations: map[string]string{"skipAuth": "true"},
		RunE:        runAuth,
	}
}

func runAuth(cmd *cobra.Command, _ []string) error {
	key, _ := cmd.Flags().GetString("api-key")

	if key == "" {
		key = os.Getenv("DUNE_API_KEY")
	}

	// stdinReader is shared across all interactive prompts so a single
	// buffered reader is used for the underlying stream.
	var stdinReader *bufio.Reader

	if key == "" {
		if stdin, ok := cmd.InOrStdin().(*os.File); ok && !term.IsTerminal(int(stdin.Fd())) {
			return fmt.Errorf("no API key provided; pass --api-key, set DUNE_API_KEY, or run dune auth in an interactive terminal")
		}

		stdinReader = bufio.NewReader(cmd.InOrStdin())

		fmt.Fprint(cmd.ErrOrStderr(), "Enter your Dune API key: ")
		line, _ := stdinReader.ReadString('\n')
		key = strings.TrimSpace(line)
	}

	if key == "" {
		return fmt.Errorf("no API key provided; pass --api-key, set DUNE_API_KEY, or run dune auth in an interactive terminal")
	}

	cfg, err := authconfig.Load()
	if err != nil {
		return fmt.Errorf("loading existing config: %w", err)
	}
	if cfg == nil {
		cfg = &authconfig.Config{}
	}
	cfg.APIKey = key
	if err := authconfig.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	p, _ := authconfig.Path()
	fmt.Fprintf(cmd.OutOrStdout(), "API key saved to %s\n", p)

	if err := maybeSampleQuery(cmd, key, stdinReader); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Sample query failed: %v\n", err)
	}

	return nil
}

const sampleSQL = "SELECT number, time, gas_used, base_fee_per_gas FROM ethereum.blocks ORDER BY number DESC LIMIT 5"

func maybeSampleQuery(cmd *cobra.Command, apiKey string, reader *bufio.Reader) error {
	// Only prompt in interactive terminals.
	if reader == nil {
		// Non-interactive path (key was provided via flag or env var).
		// Create a reader to check for a terminal; if not interactive, skip.
		stdin, ok := cmd.InOrStdin().(*os.File)
		if !ok || !term.IsTerminal(int(stdin.Fd())) {
			return nil
		}
		reader = bufio.NewReader(cmd.InOrStdin())
	}

	fmt.Fprint(cmd.ErrOrStderr(), "\nWould you like to run a sample query to verify your setup? [Y/n] ")
	line, _ := reader.ReadString('\n')
	answer := strings.TrimSpace(strings.ToLower(line))
	if answer != "" && answer != "y" && answer != "yes" {
		return nil
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "\nRunning: %s\n\n", sampleSQL)

	client := dune.NewDuneClient(config.FromAPIKey(apiKey))
	exec, err := client.RunSQL(models.ExecuteSQLRequest{
		SQL:         sampleSQL,
		Performance: "medium",
	})
	if err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	resp, err := exec.WaitGetResults(2*time.Second, sampleQueryMaxRetries)
	if err != nil {
		return fmt.Errorf("waiting for results: %w", err)
	}

	if resp.State != queryStateCompleted {
		msg := fmt.Sprintf("query finished with state %s", resp.State)
		if resp.Error != nil {
			msg += fmt.Sprintf(": %s", resp.Error.Message)
		}
		return errors.New(msg)
	}

	w := cmd.OutOrStdout()
	columns := resp.Result.Metadata.ColumnNames
	rows := output.ResultRowsToStrings(resp.Result.Rows, columns)
	output.PrintTable(w, columns, rows)
	fmt.Fprintf(w, "\n%d rows\n", len(resp.Result.Rows))
	fmt.Fprintln(w, "\nYou're all set! Run `dune query run-sql --sql \"YOUR SQL\"` to execute your own queries.")

	return nil
}
