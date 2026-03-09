package config

import "github.com/spf13/cobra"

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "Manage CLI configuration",
		Annotations: map[string]string{"skipAuth": "true"},
	}
	cmd.AddCommand(newTelemetryCmd())
	return cmd
}
