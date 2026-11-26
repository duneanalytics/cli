package dashboard_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/duneanalytics/cli/cmd/dashboard"
	"github.com/duneanalytics/cli/dunemcp/dunemcptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runDash(fake *dunemcptest.Fake, args ...string) (string, error) {
	root, buf := dunemcptest.NewRoot(dashboard.NewDashboardCmd(), fake)
	root.SetArgs(append([]string{"dashboard"}, args...))
	err := root.Execute()
	return buf.String(), err
}

const widgetResult = `{"dashboardId":5,"name":"D","dashboardUrl":"https://dune.com/alice/d",
	"visualizationWidgets":[],"textWidgets":[],"paramWidgets":[]}`

func TestUpdateWidgetSendsPositionAndText(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"updateDashboardWidget": widgetResult}}
	md := filepath.Join(t.TempDir(), "intro.md")
	require.NoError(t, os.WriteFile(md, []byte("# Intro"), 0o600))

	out, err := runDash(fake, "update-widget", "5", "--widget-type", "text", "--widget-id", "679",
		"--row", "2", "--col", "3", "--size-x", "3", "--size-y", "4", "--text", "@"+md)
	require.NoError(t, err)
	// Text and visualization widget IDs go out as numbers; sizeX/sizeY are not swapped.
	assert.JSONEq(t, `{"dashboardId":5,"widgetType":"text","widgetId":679,
		"position":{"row":2,"col":3,"sizeX":3,"sizeY":4},"text":"# Intro"}`, fake.Last().Args)
	assert.Equal(t, "Updated text widget 679 on dashboard 5\nhttps://dune.com/alice/d\n", out)
}

func TestUpdateWidgetParamIDStaysAString(t *testing.T) {
	fake := &dunemcptest.Fake{Results: map[string]string{"updateDashboardWidget": widgetResult}}
	_, err := runDash(fake, "update-widget", "5", "--widget-type", "param", "--widget-id", "123",
		"--row", "0", "--col", "0", "--size-x", "6", "--size-y", "2")
	require.NoError(t, err)
	assert.JSONEq(t, `{"dashboardId":5,"widgetType":"param","widgetId":"123",
		"position":{"row":0,"col":0,"sizeX":6,"sizeY":2}}`, fake.Last().Args)
}

func TestUpdateWidgetRejectsBadInput(t *testing.T) {
	for name, tc := range map[string]struct {
		args []string
		want string
	}{
		"nothing to change": {[]string{"--widget-type", "text", "--widget-id", "1"}, "provide a new position"},
		"partial position":  {[]string{"--widget-type", "text", "--widget-id", "1", "--row", "1"}, "size-x"},
		"non-numeric id":    {[]string{"--widget-type", "visualization", "--widget-id", "abc", "--text", "x"}, "integer IDs"},
		"unknown type":      {[]string{"--widget-type", "chart", "--widget-id", "1", "--text", "x"}, "visualization, text, or param"},
	} {
		t.Run(name, func(t *testing.T) {
			fake := &dunemcptest.Fake{}
			_, err := runDash(fake, append([]string{"update-widget", "5"}, tc.args...)...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
			assert.Empty(t, fake.Calls)
		})
	}
}

func TestImageSavesInlineJPEGAndHidesShareURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d.jpg")
	fake := &dunemcptest.Fake{
		Results: map[string]string{"getDashboardImage": `{"dashboardId":5,"dashboardUrl":"https://dune.com/alice/d",
			"dashboardImageUrl":"https://api.dune.com/screenshot?url=https://dune.com/alice/d/SHAREKEY",
			"imageIncluded":true,"withheldVisualizationIds":[]}`},
		Images: map[string][][]byte{"getDashboardImage": {{0xff, 0xd8, 0x01}}},
	}
	out, err := runDash(fake, "image", "5", "--file", path, "-o", "json")
	require.NoError(t, err)
	assert.JSONEq(t, `{"dashboardId":5}`, fake.Last().Args)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, []byte{0xff, 0xd8, 0x01}, got)
	assert.NotContains(t, out, "SHAREKEY")
	var printed map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &printed))
	assert.Equal(t, path, printed["file"])
}

func TestImageDownloadsWhenTooLargeToInline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("big-jpeg"))
	}))
	t.Cleanup(srv.Close)
	path := filepath.Join(t.TempDir(), "d.jpg")
	fake := &dunemcptest.Fake{Results: map[string]string{"getDashboardImage": `{"dashboardId":5,
		"dashboardUrl":"https://dune.com/alice/d","dashboardImageUrl":"` + srv.URL + `/KEY",
		"imageIncluded":false,"withheldVisualizationIds":[]}`}}
	out, err := runDash(fake, "image", "5", "--file", path)
	require.NoError(t, err)
	got, _ := os.ReadFile(path)
	assert.Equal(t, "big-jpeg", string(got))
	assert.NotContains(t, out, "KEY")
}

func imageFake(url string) *dunemcptest.Fake {
	return &dunemcptest.Fake{Results: map[string]string{"getDashboardImage": `{"dashboardId":5,
		"dashboardUrl":"https://dune.com/alice/d","dashboardImageUrl":"` + url + `",
		"imageIncluded":false,"withheldVisualizationIds":[]}`}}
}

func TestImageDownloadFailuresKeepTheShareKeyOut(t *testing.T) {
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	html := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>login</html>"))
	}))
	t.Cleanup(html.Close)
	forbidden := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(forbidden.Close)

	for name, base := range map[string]string{"connection refused": closed.URL, "html page": html.URL, "403": forbidden.URL} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "d.jpg")
			_, err := runDash(imageFake(base+"/SHAREKEY"), "image", "5", "--file", path)
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "SHAREKEY")
			assert.NoFileExists(t, path)
		})
	}
}

func TestImageWithheldForLicensedData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d.jpg")
	fake := &dunemcptest.Fake{Results: map[string]string{"getDashboardImage": `{"dashboardId":5,
		"dashboardUrl":"https://dune.com/alice/d","imageIncluded":false,"withheldVisualizationIds":[11,12]}`}}
	_, err := runDash(fake, "image", "5", "--file", path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "visualizations 11, 12")
	assert.NoFileExists(t, path)
}
