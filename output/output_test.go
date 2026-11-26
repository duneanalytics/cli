package output

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jsonCmd returns a command with the format flags, writing to buf.
func jsonCmd(buf *bytes.Buffer, args ...string) *cobra.Command {
	cmd := &cobra.Command{}
	AddFormatFlag(cmd, "json")
	cmd.SetOut(buf)
	_ = cmd.ParseFlags(args)
	return cmd
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	err := PrintJSON(jsonCmd(&buf), map[string]int{"query_id": 42})
	require.NoError(t, err)
	assert.Equal(t, "{\n  \"query_id\": 42\n}\n", buf.String())

	// URLs and SQL stay readable instead of turning & < > into \u0026 \u003c \u003e.
	buf.Reset()
	v := map[string]string{"url": "https://x/?a=1&b=2", "sql": "SELECT 1 WHERE 1 < 2"}
	require.NoError(t, PrintJSON(jsonCmd(&buf), v))
	assert.Contains(t, buf.String(), `"https://x/?a=1&b=2"`)
	assert.Contains(t, buf.String(), `"SELECT 1 WHERE 1 < 2"`)
}

func TestCompactJSONIsOneLine(t *testing.T) {
	var buf bytes.Buffer
	v := map[string]any{"rows": []map[string]any{{"a": 1}, {"a": 2}}, "url": "https://x/?a=1&b=2"}
	require.NoError(t, PrintJSON(jsonCmd(&buf, "--compact"), v))
	assert.Equal(t, `{"rows":[{"a":1},{"a":2}],"url":"https://x/?a=1&b=2"}`+"\n", buf.String())

	buf.Reset()
	raw := []byte("{\n  \"total\": 1,\n  \"results\": [ {\"x\": \"a b\"} ]\n}")
	require.NoError(t, PrintRawJSON(jsonCmd(&buf, "--compact"), raw))
	assert.Equal(t, `{"total":1,"results":[{"x":"a b"}]}`+"\n", buf.String())

	buf.Reset()
	require.NoError(t, PrintRawJSON(jsonCmd(&buf), []byte(`{"total":1}`)))
	assert.Equal(t, "{\n  \"total\": 1\n}\n", buf.String())
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	PrintTable(&buf, []string{"ID", "NAME"}, [][]string{
		{"1", "alpha"},
		{"2", "beta"},
	})
	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "alpha")
	assert.Contains(t, out, "beta")
}

func TestAddFormatFlag_DefaultValue(t *testing.T) {
	cmd := &cobra.Command{}
	AddFormatFlag(cmd, "json")
	assert.Equal(t, "json", FormatFromCmd(cmd))
}

func TestFormatFromCmd(t *testing.T) {
	cmd := &cobra.Command{}
	AddFormatFlag(cmd, "text")
	_ = cmd.Flags().Set("output", "json")
	assert.Equal(t, "json", FormatFromCmd(cmd))
}
