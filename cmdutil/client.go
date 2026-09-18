package cmdutil

import (
	"context"
	"time"

	"github.com/duneanalytics/cli/tracking"
	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/spf13/cobra"
)

type clientKey struct{}
type trackerKey struct{}
type startTimeKey struct{}

// SetClient stores a DuneClient in the command's context.
func SetClient(cmd *cobra.Command, client dune.DuneClient) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, clientKey{}, client))
}

// ClientFromCmd extracts the DuneClient stored in the command's context.
func ClientFromCmd(cmd *cobra.Command) dune.DuneClient {
	return cmd.Context().Value(clientKey{}).(dune.DuneClient)
}

// SetTracker stores a Tracker in the command's context.
func SetTracker(cmd *cobra.Command, t *tracking.Tracker) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, trackerKey{}, t))
}

// TrackerFromCmd extracts the Tracker stored in the command's context.
func TrackerFromCmd(cmd *cobra.Command) *tracking.Tracker {
	v, _ := cmd.Context().Value(trackerKey{}).(*tracking.Tracker)
	return v
}

// SetStartTime stores the command start time in the command's context.
func SetStartTime(cmd *cobra.Command, t time.Time) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, startTimeKey{}, t))
}

// StartTimeFromCmd extracts the command start time from the command's context.
func StartTimeFromCmd(cmd *cobra.Command) time.Time {
	v, _ := cmd.Context().Value(startTimeKey{}).(time.Time)
	return v
}
