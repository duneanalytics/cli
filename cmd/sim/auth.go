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
		Short: "Authenticate with the Sim API",
		Long: "Save your Sim API key so you don't need to pass --sim-api-key or set DUNE_SIM_API_KEY every time.",
		Annotations: map[string]string{"skipSimAuth": "true"},
		RunE:        runSimAuth,
	}

	cmd.Flags().String("api-key", "", "Sim API key to save")

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
