package matview

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/spf13/cobra"
)

// matviewNameRe mirrors the server-side matview name validator (handlers/matviews and the
// GraphQL matviewName validator): a "result_" prefix, lowercase alphanumerics and underscores,
// and no trailing underscore.
var matviewNameRe = regexp.MustCompile(`^result_([a-z0-9_]{0,}[a-z0-9]+)?$`)

// validateMatviewName checks a bare matview name client-side for a fast, friendly failure.
// The server remains the source of truth.
func validateMatviewName(name string) error {
	if len(name) < 8 || len(name) > 128 {
		return fmt.Errorf("invalid matview name %q: must be between 8 and 128 characters", name)
	}
	if !matviewNameRe.MatchString(name) {
		return fmt.Errorf(
			"invalid matview name %q: must start with \"result_\", contain only lowercase letters, "+
				"numbers, and underscores, and not end with an underscore",
			name,
		)
	}
	return nil
}

// parsePerformance validates the --performance flag. An empty value means "use the account
// default tier".
func parsePerformance(cmd *cobra.Command) (string, error) {
	performance, _ := cmd.Flags().GetString("performance")
	switch performance {
	case "", "small", "medium", "large":
		return performance, nil
	default:
		return "", fmt.Errorf(
			"invalid performance tier %q: must be \"small\", \"medium\", or \"large\"",
			performance,
		)
	}
}

// privacyNote explains why --private is deprecated; cobra prints it whenever the flag is used.
const privacyNote = "a materialized view always has its source query's privacy; " +
	"change it with 'dune query update <query-id> --private'"

// addPrivateFlag registers the deprecated --private flag, which scripts may still pass.
func addPrivateFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("private", false, "no longer settable: "+privacyNote)
	_ = cmd.Flags().MarkDeprecated("private", privacyNote)
}

// checkPrivacy fails when --private asks for a privacy the source query does not have. The
// server ignores is_private and copies the query's privacy, so without this check
// --private on a public query would quietly create a public matview.
func checkPrivacy(cmd *cobra.Command, client dune.DuneClient, queryID int) error {
	if !cmd.Flags().Changed("private") {
		return nil
	}
	want, _ := cmd.Flags().GetBool("private")
	query, err := client.GetQuery(queryID)
	if err != nil {
		return fmt.Errorf("checking the privacy of query %d: %w", queryID, err)
	}
	if query.IsPrivate == want {
		return nil
	}
	return fmt.Errorf(
		"query %d is %s, and a materialized view always has its source query's privacy; "+
			"to make it %s, run 'dune query update %d --private=%t' first",
		queryID, privacyName(query.IsPrivate), privacyName(want), queryID, want,
	)
}

func privacyName(private bool) string {
	if private {
		return "private"
	}
	return "public"
}

// bareName returns the table segment of a fully-qualified SQL name
// (e.g. "dune.my_team.result_x" -> "result_x"). The upsert endpoint takes the bare name, while
// get/refresh/delete take the fully-qualified name.
func bareName(fqName string) string {
	if i := strings.LastIndex(fqName, "."); i >= 0 {
		return fqName[i+1:]
	}
	return fqName
}
