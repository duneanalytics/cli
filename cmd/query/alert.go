package query

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

// alertChannels maps each channel flag to the channel type setQueryAlert takes.
var alertChannels = map[string]string{
	"slack-webhook":   "slack",
	"discord-webhook": "discord",
	"webhook":         "webhook",
}

func newAlertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert <query-id>",
		Short: "Notify you each time a saved query's scheduled run returns rows",
		Long: `Put a saved query on a schedule and send a notification each time a run
returns rows. This sets the query's schedule (creating it or replacing its
frequency and tier) and creates or replaces the alert for the chosen channel.

Write the query so it returns rows only when the event happened in the latest
period, such as the last hour for an hourly schedule.

Choose one channel:
  --email a@x.com --email b@x.com   up to 5 addresses
  --slack-webhook <url>             Slack incoming webhook
  --discord-webhook <url>           Discord webhook
  --webhook <url>                   HTTPS URL that receives a JSON POST per fire

- A query has one alert per channel; this replaces that channel's alert.
- Every scheduled run is charged credits at --performance, so use the smallest
  tier and the lowest frequency that catch the event in time.
- You need to own the query or be an editor or admin of the team that owns it.
  The owner's plan limits how many queries can have alerts and schedules.
- Recipients receive the query name, a link to it and its charts.

Supported cron expressions (5-part, UTC):
  "*/15 * * * *", "*/20 * * * *", "*/30 * * * *"  every 15, 20 or 30 minutes
  "0 * * * *"                                    hourly
  "0 */N * * *" with N in 2, 3, 4, 6, 8, 12      every N hours
  "M H * * *"                                    daily at H:M
  "M H * * D"                                    weekly on day D (0 = Sunday)
where M is 0 or 30 and H is 0-23.

To stop the schedule, use 'dune schedule delete --query-id <id>'.

Examples:
  dune query alert 12345 --cron "0 * * * *" --performance small --email me@example.com
  dune query alert 12345 --cron "*/30 * * * *" --performance small --slack-webhook "$SLACK_URL"`,
		Args: cobra.ExactArgs(1),
		RunE: runAlert,
	}

	cmd.Flags().String("cron", "", "when to run, as a supported 5-part UTC cron expression (required)")
	cmd.Flags().String("performance", "", `engine tier for each run: "small", "medium", or "large" (required)`)
	cmd.Flags().StringArray("email", nil, "address to notify; repeat for up to 5")
	cmd.Flags().String("slack-webhook", "", "Slack incoming webhook URL")
	cmd.Flags().String("discord-webhook", "", "Discord webhook URL")
	cmd.Flags().String("webhook", "", "HTTPS URL that receives a JSON POST for each fire")
	_ = cmd.MarkFlagRequired("cron")
	_ = cmd.MarkFlagRequired("performance")
	cmd.MarkFlagsOneRequired("email", "slack-webhook", "discord-webhook", "webhook")
	cmd.MarkFlagsMutuallyExclusive("email", "slack-webhook", "discord-webhook", "webhook")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

func runAlert(cmd *cobra.Command, args []string) error {
	queryID, err := strconv.Atoi(args[0])
	if err != nil || queryID <= 0 {
		return fmt.Errorf("invalid query ID %q: must be a positive integer", args[0])
	}

	var channel map[string]any
	if cmd.Flags().Changed("email") {
		emails, _ := cmd.Flags().GetStringArray("email")
		channel = map[string]any{"type": "email", "emails": emails}
	}
	for flag, typ := range alertChannels {
		if cmd.Flags().Changed(flag) {
			url, _ := cmd.Flags().GetString(flag)
			if url == "" {
				return fmt.Errorf("--%s must not be empty", flag)
			}
			channel = map[string]any{"type": typ, "url": url}
		}
	}
	if channel == nil {
		return errors.New("choose a channel: --email, --slack-webhook, --discord-webhook, or --webhook")
	}

	toolArgs := cmdutil.ToolArgs(cmd, map[string]string{"cron": "cron_expression", "performance": "performance"})
	toolArgs["query_id"] = queryID
	toolArgs["channel"] = channel

	// The result has no webhook URL, so it is safe to print as it is.
	var res struct {
		AlertID           string `json:"alert_id"`
		Channel           string `json:"channel"`
		CronExpression    string `json:"cron_expression"`
		NextExecutionTime string `json:"next_execution_time"`
	}
	raw, err := cmdutil.CallTool(cmd, "setQueryAlert", toolArgs, &res)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		return output.PrintRawJSON(w, raw)
	}
	fmt.Fprintf(w, "Set %s alert %s on query %d\n", res.Channel, res.AlertID, queryID)
	fmt.Fprintf(w, "Schedule: %s (UTC), next run %s\n", res.CronExpression, res.NextExecutionTime)
	return nil
}
