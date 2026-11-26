package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/duneanalytics/cli/cmdutil"
	"github.com/duneanalytics/cli/output"
	"github.com/spf13/cobra"
)

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image <dashboard_id>",
		Short: "Save a JPEG image of a whole dashboard",
		Long: `Render a whole dashboard as dune.com shows it and save it as a JPEG file, to
check its layout after 'dune dashboard update' or 'dune dashboard update-widget'.

The image uses each query's latest completed results; it does not run any query.
Rendering takes several seconds.

Dune withholds the image when a chart on the dashboard shows enterprise datasets
that are not available to AI tools, or a query that has never completed; the
command then fails and lists those visualizations.

For a private dashboard, rendering needs a share link: Dune reuses the
dashboard's active share link or creates one, as the Share button on dune.com
does. The command never prints that link.

Examples:
  dune dashboard image 12345
  dune dashboard image 12345 --file layout.jpg`,
		Args: cobra.ExactArgs(1),
		RunE: runImage,
	}

	cmd.Flags().String("file", "", "where to save the JPEG (default dashboard-<id>.jpg)")
	output.AddFormatFlag(cmd, "text")

	return cmd
}

type dashboardImage struct {
	DashboardID              int    `json:"dashboardId"`
	DashboardURL             string `json:"dashboardUrl"`
	DashboardImageURL        string `json:"dashboardImageUrl"`
	WithheldVisualizationIDs []int  `json:"withheldVisualizationIds"`
}

func runImage(cmd *cobra.Command, args []string) error {
	dashboardID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid dashboard ID %q: must be an integer", args[0])
	}
	path, _ := cmd.Flags().GetString("file")
	if path == "" {
		path = fmt.Sprintf("dashboard-%d.jpg", dashboardID)
	}

	var res dashboardImage
	raw, images, err := cmdutil.CallToolWithImages(cmd, "getDashboardImage", map[string]any{"dashboardId": dashboardID}, &res)
	if err != nil {
		return err
	}
	if len(res.WithheldVisualizationIDs) > 0 {
		ids := make([]string, len(res.WithheldVisualizationIDs))
		for i, id := range res.WithheldVisualizationIDs {
			ids[i] = strconv.Itoa(id)
		}
		return fmt.Errorf("the image is withheld because visualizations %s show enterprise datasets "+
			"that are not available to AI tools, or a query that has never completed", strings.Join(ids, ", "))
	}

	var jpeg []byte
	switch {
	case len(images) > 0:
		jpeg = images[0]
	case res.DashboardImageURL != "":
		// The tool leaves out images larger than it returns inline.
		if jpeg, err = download(cmd, res.DashboardImageURL); err != nil {
			return err
		}
	default:
		return fmt.Errorf("getDashboardImage returned no image for dashboard %d", dashboardID)
	}
	if err := os.WriteFile(path, jpeg, 0o644); err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if output.FormatFromCmd(cmd) == output.FormatJSON {
		// The image URL of a private dashboard carries its share key, so it is left out.
		var out map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			return err
		}
		delete(out, "dashboardImageUrl")
		delete(out, "imageIncluded")
		out["file"] = path
		return output.PrintJSON(w, out)
	}
	fmt.Fprintf(w, "Saved an image of dashboard %d to %s\n", dashboardID, path)
	fmt.Fprintf(w, "Dashboard: %s\n", res.DashboardURL)
	return nil
}

func download(cmd *cobra.Command, imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, errors.New("getDashboardImage returned an invalid image URL")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Drop the URL from the error: it may carry the dashboard's share key.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("downloading the dashboard image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading the dashboard image failed: %s", resp.Status)
	}
	// A 200 with an HTML error or login page must not be saved as the image.
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "image/") {
		return nil, fmt.Errorf("downloading the dashboard image failed: got %q instead of an image", ct)
	}
	return io.ReadAll(resp.Body)
}
