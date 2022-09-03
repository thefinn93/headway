package tasks

import (
	"encoding/csv"
	"path/filepath"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/headwaymaps/headway/cmd/headway-build/tasks/task"
)

const (
	// gtfsListURL comes from https://database.mobilitydata.org
	gtfsListURL                    = "https://storage.googleapis.com/storage/v1/b/mdb-csv/o/sources.csv?alt=media"
	gtfsListColumnProviderName     = 6
	gtfsListColumnURL              = 14
	gtfsListColumnMinimumLatitude  = 16
	gtfsListColumnMaximumLatitude  = 17
	gtfsListColumnMinimumLongitude = 18
	gtfsListColumnMaximumLongitude = 19
	gtfsListMinExpectedColumns     = gtfsListColumnMaximumLongitude // highest column number
)

type GTFSDownloadTask struct {
	processed int
	Feeds     []GTFSFeed
	bounds    [4]float64
}

type GTFSFeed struct {
	Provider string
	Filename string
	URL string
}

func GTFSDownload(bounds [4]float64) []GTFSFeed {
	t := GTFSDownloadTask{
		bounds: bounds,
	}
	task.Execute(&t)
	return t.Feeds
}

func (g GTFSDownloadTask) View() string {
	return fmt.Sprintf("looking for gtfs feeds within selected area (%d/%d)", len(g.Feeds), g.processed)
}

func (g *GTFSDownloadTask) Run() (task.Result, error) {
	res, err := http.Get(gtfsListURL)
	if err != nil {
		return task.Result{}, fmt.Errorf("error fetching GTFS feed list: %v", err)
	}
	defer res.Body.Close()

	csvReader := csv.NewReader(res.Body)

	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return task.Result{}, fmt.Errorf("error parsing line %d of GTFS feed list: %v", g.processed, err)
		}

		if len(row) < gtfsListMinExpectedColumns {
			fmt.Println(" ⚠️ invalid line in gtfs feed list: ", g.processed)
			continue
		}

		minLong, _ := strconv.ParseFloat(row[gtfsListColumnMinimumLongitude], 64)
		maxLong, _ := strconv.ParseFloat(row[gtfsListColumnMaximumLongitude], 64)
		minLat, _ := strconv.ParseFloat(row[gtfsListColumnMinimumLatitude], 64)
		maxLat, _ := strconv.ParseFloat(row[gtfsListColumnMaximumLatitude], 64)

		g.processed = g.processed + 1

		if maxLat-minLat > 18 || maxLong-minLong > 16 {
			// This almost certainly just means the transit provider operates "everywhere".
			continue
		}

		if row[gtfsListColumnURL] == "" {
			continue
		}

		if maxLong > g.bounds[0] && minLong > g.bounds[0] || minLong < g.bounds[2] && maxLong < g.bounds[2] {
			continue
		}

		if maxLat > g.bounds[1] && minLat > g.bounds[1] || minLat < g.bounds[3] && maxLat < g.bounds[3] {
			continue
		}

		u, err := url.Parse(row[gtfsListColumnURL])
		if err != nil {
			return task.Result{}, fmt.Errorf("error parsing %s feed URL %s: %v", row[gtfsListColumnProviderName], row[gtfsListColumnURL], err)
		}

		g.Feeds = append(g.Feeds, GTFSFeed{
			Provider: row[gtfsListColumnProviderName],
			Filename: filepath.Base(u.Path),
			URL: u.String(),
		})
	}

	resultIcon := task.ResultIconSuccess
	if len(g.Feeds) == 0 {
		resultIcon = task.ResultIconUnchanged
	}

	return task.Result{
		Icon: resultIcon,
		Message: fmt.Sprintf("found %d transit feeds", len(g.Feeds)),
	}, nil
}
