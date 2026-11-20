package dashboard_test

import (
	"testing"

	"github.com/duneanalytics/cli/cmd/dashboard"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const searchResult = `{"results":[
	{"dashboard_id":5120,"name":"DEX Metrics","slug":"dex-metrics",
	 "owner":{"type":"user","id":7,"handle":"hagaetc","name":"hagaetc"}}]}`

func TestSearch(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"searchDuneDashboards": searchResult}}
	root, buf := dunemcptest.NewRoot(dashboard.NewDashboardCmd(), fake)
	root.SetArgs([]string{"dashboard", "search", "--query", "dex", "--owned-only", "--limit", "5"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "searchDuneDashboards", fake.Last().Tool)
	assert.JSONEq(t, `{"search_string":"dex","scope":"owned_only","limit":5}`, fake.Last().Args)
	// The owner and slug are what 'dune dashboard get' needs.
	assert.Regexp(t, `5120\s+DEX Metrics\s+hagaetc\s+dex-metrics`, buf.String())
}

func TestSearchSendsOnlyTheQueryByDefault(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"searchDuneDashboards": `{"results":[]}`}}
	root, buf := dunemcptest.NewRoot(dashboard.NewDashboardCmd(), fake)
	root.SetArgs([]string{"dashboard", "search", "--query", "nothing"})
	require.NoError(t, root.Execute())

	assert.JSONEq(t, `{"search_string":"nothing"}`, fake.Last().Args)
	assert.Equal(t, "No dashboards found.\n", buf.String())
}
