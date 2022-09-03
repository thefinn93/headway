package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/headwaymaps/headway/cmd/headway-build/tasks"
	"github.com/spf13/cobra"
)

var (
	dataDir string
	area    string
	country string
	rootCmd = &cobra.Command{
		Use:   "headway-build",
		Short: "builds headway components",
		Run: func(_ *cobra.Command, _ []string) {
			tasks.Download(fmt.Sprintf("https://download.bbbike.org/osm/bbbike/%s/%s.osm.pbf", area, area), filepath.Join(dataDir, "data.osm.pbf"))

			if tasks.Download("https://f000.backblazeb2.com/file/headway/sources.tar", filepath.Join(dataDir, "sources.tar")) {
				tasks.Untar(filepath.Join(dataDir, "sources.tar"), filepath.Join(dataDir, "sources"))
			}

			// tasks.RunContainer("ghcr.io/onthegomap/planetiler", tasks.RunContainerOptions{
			// 	Command: []string{"--force", "--osm_path=/data/data.osm.pbf"},
			// 	Volumes: []tasks.ContainerVolume{
			// 		tasks.ContainerVolume{Destination: "/data", Source: dataDir},
			// 	},
			// })

			gtfsFeeds := tasks.GTFSDownload(cities[area])
			for _, feed := range gtfsFeeds {
				tasks.Download(feed.URL, filepath.Join(dataDir, "gtfs", feed.Filename))
			}
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
