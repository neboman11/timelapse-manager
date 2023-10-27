package main

import (
	"flag"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/neboman11/timelapse-manager/api"
	_ "github.com/neboman11/timelapse-manager/docs"
	"github.com/neboman11/timelapse-manager/models"
)

func main() {
	log.Info().Msgf("API version: %s", CurrentVersionNumber)
	log.Trace().Msg("Successfully loaded config")

	portPtr := flag.Int("p", 3001, "Port to listen for requests on.")
	flag.Parse()

	// Verify the given port is a valid port number
	if *portPtr < 1 || *portPtr > 65535 {
		log.Fatal().Msg("Invalid port specified.")
	}

	startDaemon(*portPtr)
}

// Setup IPFS and start listening for requests
func startDaemon(port int) {
	open_database()

	go encodeInProgress()

	log.Info().Msg("Ready for requests")
	api.HandleRequests(port, db)
}

func encodeInProgress() {
	// ticker := time.NewTicker(24 * time.Hour)
	ticker := time.NewTicker(10 * time.Minute)

	for {
		<-ticker.C
		var inProgressTimelapses []models.InProgress
		db.Where(models.InProgress{Status: "InProgress"}).Find(&inProgressTimelapses)

		for _, timelapse := range inProgressTimelapses {
			var ffmpegArgs []string

			ffmpegArgs = append(ffmpegArgs, "-f")
			ffmpegArgs = append(ffmpegArgs, "image2")
			ffmpegArgs = append(ffmpegArgs, "-r")
			ffmpegArgs = append(ffmpegArgs, "1")
			ffmpegArgs = append(ffmpegArgs, "-i")
			ffmpegArgs = append(ffmpegArgs, path.Join(timelapse.Folder, "%05d.jpg"))
			ffmpegArgs = append(ffmpegArgs, path.Join(timelapse.Folder, "complete.mp4"))

			cmd := exec.Command("ffmpeg", ffmpegArgs...)

			stdout, err := cmd.CombinedOutput()
			if err != nil {
				log.Error().Err(err)
			}

			log.Info().Msgf("%s", stdout)

			timelapse.Status = "Complete"
			db.Save(timelapse)

			var video models.Videos
			video.Date = time.Now()
			video.Location = path.Join(timelapse.Folder, "complete.mp4")
			db.Create(&video)

			files, err := os.ReadDir(timelapse.Folder)
			if err != nil {
				log.Error().Err(err)
			}

			for _, file := range files {
				if strings.HasSuffix(file.Name(), "jpg") {
					os.Remove(path.Join(timelapse.Folder, file.Name()))
				}
			}
		}
	}
}
