package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/headwaymaps/headway/cmd/headway-build/tasks"
	"github.com/headwaymaps/headway/cmd/headway-build/tasks/containers"
)

var (
	dataDir string
	area    string
	country string
	rootCmd = &cobra.Command{
		Use:   "headway-build",
		Short: "builds headway components",
		Run: func(_ *cobra.Command, _ []string) {
			areaDir := filepath.Join(dataDir, area)

			tasks.Download(fmt.Sprintf("https://download.bbbike.org/osm/bbbike/%s/%s.osm.pbf", area, area), fmt.Sprintf("%s.osm.pbf", area), filepath.Join(areaDir, "data.osm.pbf"))

			if tasks.Download("https://f000.backblazeb2.com/file/headway/sources.tar", "planetiler sources", filepath.Join(dataDir, "sources.tar")) {
				tasks.Untar(filepath.Join(dataDir, "sources.tar"), filepath.Join(dataDir, "sources"))
			}

			containers.RunContainer(containers.Options{
				Image:   "ghcr.io/onthegomap/planetiler",
				Name:    containers.TaskName{Before: "generate", During: "generating", After: "generated", Suffix: "mbtiles with planetiler"},
				Command: []string{"--force", fmt.Sprintf("--osm_path=/data/%s/data.osm.pbf", area)},
				Volumes: []containers.Volume{
					containers.Volume{Destination: "/data", Source: dataDir},
				},
			})

			gtfsFeeds := tasks.GTFSDownload(cities[area])
			gtfsDir := filepath.Join(dataDir, "gtfs")
			for _, feed := range gtfsFeeds {
				tasks.Download(feed.URL, fmt.Sprintf("GTFS feed for %s", feed.Provider), filepath.Join(gtfsDir, feed.Filename))
			}

			containers.RunContainer(containers.Options{
				Image:   "docker.io/opentripplanner/opentripplanner",
				Name:    containers.TaskName{Before: "generate", During: "generating", After: "generated", Suffix: "transit graph with opentripplanner"},
				Command: []string{"--build", "--save"},
				Volumes: []containers.Volume{
					containers.Volume{Destination: "/var/opentripplanner", Source: gtfsDir},
				},
			})

			containers.RunContainer(containers.Options{
				Image:      "docker.io/gisops/valhalla",
				Name:       containers.TaskName{Before: "build", During: "building", After: "built", Suffix: "tiles with valhalla"},
				User:       "root",
				Entrypoint: []string{"/bin/bash", "-exc"},
				Command: []string{`
chown valhalla /tiles
sudo -u valhalla /bin/bash -exc '
cd $(mktemp -d)
/usr/local/bin/valhalla_build_config --mjolnir-tile-dir /tiles --mjolnir-timezone /tiles/timezones.sqlite --mjolnir-admin /tiles/admins.sqlite > valhalla.json
valhalla_build_timezones > /tiles/timezones.sqlite
valhalla_build_tiles -c valhalla.json /data.osm.pbf
'
`},
				Volumes: []containers.Volume{
					containers.Volume{Destination: "/tiles", Source: filepath.Join(dataDir, "tiles")},
					containers.Volume{Destination: "/data.osm.pbf", Source: filepath.Join(areaDir, "data.osm.pbf")},
				},
			})
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&dataDir, "data-dir", "d", "data", "directory to store downloaded artifacts in during build. will be created if needed")
	rootCmd.Flags().StringVarP(&area, "area", "a", "", "the metro area to build")
	rootCmd.Flags().StringVarP(&country, "country", "c", "", "the country that the metro area is in")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
