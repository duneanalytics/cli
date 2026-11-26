package cmdutil

import (
	"errors"
	"net/http"
	"strings"

	"github.com/duneanalytics/duneapi-client-go/dune"
)

const (
	pricingURL      = "https://dune.com/pricing"
	planSettingsURL = "https://dune.com/settings/plan"
	teamSettingsURL = "https://dune.com/settings/teams"
)

// PlanHint returns the next steps for an error caused by the account's plan:
// running out of credits, reaching the extra credits limit, a rate limit, or
// unpaid invoices. It returns "" for any other error.
//
// Unlike the Dune MCP server, which plugin directories keep from promoting
// upgrades, the CLI links straight to the plan settings and pricing pages.
func PlanHint(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	var apiErr *dune.APIError
	status := 0
	if errors.As(err, &apiErr) {
		status = apiErr.StatusCode
	}

	var lines []string
	// The REST API already links some of these pages; repeat only the others.
	link := func(text, url string) {
		if !strings.Contains(msg, url) {
			lines = append(lines, text+url)
		}
	}
	switch {
	// Gated on the status: a SQL error can name a column such as unpaid_balance.
	case status == http.StatusPaymentRequired && strings.Contains(msg, "unpaid"):
		link("Update your payment details in your plan settings: ", planSettingsURL)
	case status == http.StatusTooManyRequests || strings.Contains(msg, "too many requests"):
		link("Plans with higher rate limits: ", pricingURL)
	case status == http.StatusPaymentRequired || isCreditsLimit(msg):
		lines = append(lines, "Credits used this billing period: dune usage")
		link("Upgrade your plan: ", pricingURL)
		if !strings.Contains(msg, planSettingsURL) && !strings.Contains(msg, teamSettingsURL) {
			lines = append(lines, "Change your extra credits limit: "+planSettingsURL+
				" (for a team, open the team at "+teamSettingsURL+")")
		}
	}
	return strings.Join(lines, "\n")
}

// isCreditsLimit reports whether an error message from the Dune MCP server or
// the query engine says a credit limit stopped the request. The phrases are
// the exact wording of those messages, so that query errors which merely
// mention credits get no hint. REST refusals are caught by their 402 status.
func isCreditsLimit(msg string) bool {
	for _, s := range []string{
		"not enough credits",                             // core's not_enough_credits error
		"enough credits left",                            // MCP, unknown cause
		"used all the credits",                           // MCP, free plan or extra credits off or at the limit
		"used all its credits",                           // MCP, trial
		"the credits and extra credits its limits allow", // MCP, team or organization limit
		"user defined maximum amount of resources",       // query engine, per-query cap
	} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}
