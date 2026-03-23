package sim

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/duneanalytics/cli/authconfig"
)

// NewAuthCmd returns the `sim auth` command.
func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Save your Sim API key to the local configuration file",
		Long: "Persist your Sim API key to ~/.config/dune/config.yaml so subsequent\n" +
			"'dune sim' commands authenticate automatically without requiring\n" +
			"--sim-api-key or the DUNE_SIM_API_KEY environment variable.\n\n" +
			"The key can be provided via:\n" +
			"  1. --api-key flag\n" +
			"  2. DUNE_SIM_API_KEY environment variable\n" +
			"  3. Interactive prompt (if neither of the above is set)\n\n" +
			"The saved key is used as the lowest-priority fallback; --sim-api-key and\n" +
			"DUNE_SIM_API_KEY always take precedence when set.\n\n" +
			"Examples:\n" +
			"  dune sim auth\n" +
			"  dune sim auth --api-key sim_abc123...",
		Annotations: map[string]string{"skipSimAuth": "true"},
		RunE:        runSimAuth,
	}

	cmd.Flags().String("api-key", "", "Sim API key to save (prefixed 'sim_'); if omitted, reads from DUNE_SIM_API_KEY or prompts interactively")

	return cmd
}

func runSimAuth(cmd *cobra.Command, _ []string) error {
	key, _ := cmd.Flags().GetString("api-key")

	if key == "" {
		key = os.Getenv("DUNE_SIM_API_KEY")
	}

	if key == "" {
		fmt.Fprint(cmd.ErrOrStderr(), "Enter your Sim API key: ")
		scanner := bufio.NewScanner(cmd.InOrStdin())
		if scanner.Scan() {
			key = strings.TrimSpace(scanner.Text())
		}
	}

	if key == "" {
		return fmt.Errorf("no API key provided")
	}

	cfg, err := authconfig.Load()
	if err != nil {
		return fmt.Errorf("loading existing config: %w", err)
	}
	if cfg == nil {
		cfg = &authconfig.Config{}
	}
	cfg.SimAPIKey = key
	if err := authconfig.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	p, _ := authconfig.Path()
	fmt.Fprintf(cmd.OutOrStdout(), "Sim API key saved to %s\n", p)
	return nil
}
