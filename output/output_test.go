package output

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	err := PrintJSON(&buf, map[string]int{"query_id": 42})
	require.NoError(t, err)
	assert.JSONEq(t, `{"query_id": 42}`, buf.String())
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
