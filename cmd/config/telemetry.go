package config

import (
	"fmt"
	"os"

	"github.com/duneanalytics/cli/authconfig"
	"github.com/spf13/cobra"
)

func newTelemetryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Manage anonymous usage telemetry",
	}
	cmd.AddCommand(
		newTelemetryEnableCmd(),
		newTelemetryDisableCmd(),
		newTelemetryStatusCmd(),
	)
	return cmd
}

func newTelemetryEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "enable",
		Short:       "Enable anonymous usage telemetry",
		Annotations: map[string]string{"skipAuth": "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return setTelemetry(cmd, true)
		},
	}
}

func newTelemetryDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "disable",
		Short:       "Disable anonymous usage telemetry",
		Annotations: map[string]string{"skipAuth": "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return setTelemetry(cmd, false)
		},
	}
}

func newTelemetryStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "status",
		Short:       "Show current telemetry status",
		Annotations: map[string]string{"skipAuth": "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			enabled := IsTelemetryEnabled()
			if enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "Telemetry is enabled.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Telemetry is disabled.")
			}
			return nil
		},
	}
}

func setTelemetry(cmd *cobra.Command, enabled bool) error {
	cfg, err := authconfig.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg == nil {
		cfg = &authconfig.Config{}
	}
	cfg.Telemetry = &enabled
	if err := authconfig.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	if enabled {
		fmt.Fprintln(cmd.OutOrStdout(), "Telemetry enabled.")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Telemetry disabled.")
	}
	return nil
}

// IsTelemetryEnabled checks env vars and config to determine if telemetry is on.
func IsTelemetryEnabled() bool {
	if os.Getenv("DUNE_NO_TELEMETRY") == "1" {
		return false
	}
	if os.Getenv("CI") != "" {
		return false
	}
	cfg, err := authconfig.Load()
	if err != nil || cfg == nil {
		return true // default: enabled
	}
	if cfg.Telemetry == nil {
		return true // nil = default opt-in
	}
	return *cfg.Telemetry
}
