package cmdutil

import (
	"context"

	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/spf13/cobra"
)

type clientKey struct{}

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
