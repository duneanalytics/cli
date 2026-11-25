package cmdutil

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/duneanalytics/duneapi-client-go/dune"
	"github.com/stretchr/testify/assert"
)

func TestPlanHint(t *testing.T) {
	creditsHint := "Credits used this billing period: dune usage\n" +
		"Upgrade your plan: https://dune.com/pricing\n" +
		"Change your extra credits limit: https://dune.com/settings/plan (for a team, open the team at https://dune.com/settings/teams)"

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "REST credits refusal",
			err: &dune.APIError{
				StatusCode: http.StatusPaymentRequired,
				Message:    "This request needs more credits than your Dune plan has left for this billing period.",
			},
			want: creditsHint,
		},
		{
			name: "MCP tool error at the extra credits limit, wrapped by a command",
			err: fmt.Errorf("running query 7: %w", errors.New(
				"This Dune account has used all the credits its plan includes for this billing period and has reached the extra credits limit it set (500 extra credits)",
			)),
			want: creditsHint,
		},
		{
			name: "run stopped at the per-query cap",
			err:  errors.New("Query execution has exceeded the user defined maximum amount of resources"),
			want: creditsHint,
		},
		{
			name: "links the REST message already gives are not repeated",
			err: &dune.APIError{
				StatusCode: http.StatusPaymentRequired,
				Message:    "Raise the limit in your plan settings at https://dune.com/settings/plan, or upgrade your plan at https://dune.com/pricing.",
			},
			want: "Credits used this billing period: dune usage",
		},
		{
			name: "rate limit",
			err:  &dune.APIError{StatusCode: http.StatusTooManyRequests, Message: "Too many requests"},
			want: "Plans with higher rate limits: https://dune.com/pricing",
		},
		{
			name: "unpaid invoices are a payment problem, not a credits one",
			err: &dune.APIError{
				StatusCode: http.StatusPaymentRequired,
				Message:    "You have unpaid outstanding invoices.",
			},
			want: "Update your payment details in your plan settings: https://dune.com/settings/plan",
		},
		{
			name: "unrelated error",
			err:  &dune.APIError{StatusCode: http.StatusNotFound, Message: "Query not found"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, PlanHint(tt.err))
		})
	}
}
