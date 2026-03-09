package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/spf13/cobra"
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

	if key == "" {
		fmt.Fprint(cmd.ErrOrStderr(), "Enter your Dune API key: ")
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
	cfg.APIKey = key
	if err := authconfig.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	p, _ := authconfig.Path()
	fmt.Fprintf(cmd.OutOrStdout(), "API key saved to %s\n", p)
	return nil
}
