package matview

import (
	"fmt"
	"regexp"
	"strings"

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

// bareName returns the table segment of a fully-qualified SQL name
// (e.g. "dune.my_team.result_x" -> "result_x"). The upsert endpoint takes the bare name, while
// get/refresh/delete take the fully-qualified name.
func bareName(fqName string) string {
	if i := strings.LastIndex(fqName, "."); i >= 0 {
		return fqName[i+1:]
	}
	return fqName
}
