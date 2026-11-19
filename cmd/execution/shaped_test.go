package execution_test

import (
	"bytes"
	"testing"

	"github.com/duneanalytics/cli/cmd/execution"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runShaped(fake *dunemcptest.Fake, args ...string) (*bytes.Buffer, error) {
	root, buf := dunemcptest.NewRoot(execution.NewExecutionCmd(), fake)
	root.SetArgs(append([]string{"execution", "results", "01EXEC", "--no-wait"}, args...))
	return buf, root.Execute()
}

const shapedResult = `{"executionId":"01EXEC","state":"COMPLETED",
	"resultMetadata":{"columns":[{"name":"amount","type":"bigint"},{"name":"token","type":"varchar"}],"totalRowCount":10},
	"data":{"rows":[{"token":"ETH","amount":9},{"token":"USDC","amount":4}]}}`

func TestShapedResultsArguments(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"getExecutionResults": shapedResult}}
	buf, err := runShaped(fake, "--columns", "amount,token", "--filter", `"amount" > 1`,
		"--sort-by", `"amount" DESC`, "--sort-by", "token")
	require.NoError(t, err)

	// --limit 0 (all rows) becomes the most one call returns.
	assert.JSONEq(t, `{"executionId":"01EXEC","timeout":0,"outputColumns":["amount","token"],
		"whereClause":"\"amount\" > 1","orderByClauses":["\"amount\" DESC","token"],"limit":32000}`, fake.Last().Args)
	assert.Regexp(t, `amount\s+token\n9\s+ETH\n4\s+USDC\n`, buf.String())
	assert.Contains(t, buf.String(), "2 of 10 rows")
}

func TestSampleSendsNoPagination(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"getExecutionResults": shapedResult}}
	_, err := runShaped(fake, "--sample", "3")
	require.NoError(t, err)
	assert.JSONEq(t, `{"executionId":"01EXEC","timeout":0,"sampleSize":3}`, fake.Last().Args)

	_, err = runShaped(fake, "--sample", "3", "--offset", "5")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the others can be")
}

func TestShapedResultsFailedExecution(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"getExecutionResults": `{"executionId":"01EXEC",
		"state":"FAILED","errorMessage":"Column 'x' cannot be resolved"}`}}
	_, err := runShaped(fake, "--filter", `"x" > 1`)
	assert.EqualError(t, err, "execution 01EXEC failed: Column 'x' cannot be resolved")
}
