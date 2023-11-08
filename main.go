package main

import (
	"flag"
	"fmt"
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
	ticker := time.NewTicker(24 * time.Hour)

	for {
		log.Debug().Msg("Beginning encoding process")
		var inProgressTimelapses []models.InProgress
		db.Where(models.InProgress{Status: "InProgress"}).Find(&inProgressTimelapses)

		for _, timelapse := range inProgressTimelapses {
			// Check to not encode one that is in progress within 24 hours
			if timelapse.Date.After(time.Now().Add(-24 * time.Hour)) {
				continue
			}

			log.Debug().Msgf("Starting encoding for %d", timelapse.Id)
			timelapse.Status = "Complete"
			db.Save(timelapse)

			var ffmpegArgs []string

			ffmpegArgs = append(ffmpegArgs, "-y")
			ffmpegArgs = append(ffmpegArgs, "-f")
			ffmpegArgs = append(ffmpegArgs, "image2")
			ffmpegArgs = append(ffmpegArgs, "-r")
			ffmpegArgs = append(ffmpegArgs, "3")
			ffmpegArgs = append(ffmpegArgs, "-i")
			ffmpegArgs = append(ffmpegArgs, path.Join(timelapse.Folder, "%05d.jpg"))
			ffmpegArgs = append(ffmpegArgs, path.Join(timelapse.Folder, "complete.mp4"))

			cmd := exec.Command("ffmpeg", ffmpegArgs...)

			stdout, err := cmd.CombinedOutput()
			if err != nil {
				log.Error().Str("encode_step", "exec").Err(err).Msgf("%s", stdout)
				continue
			}

			log.Debug().Str("working_directory", timelapse.Folder).Str("command", fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))).Msgf("%s", stdout)

			var video models.Video
			video.Date = time.Now()
			video.Location = path.Join(timelapse.Folder, "complete.mp4")
			db.Create(&video)

			files, err := os.ReadDir(timelapse.Folder)
			if err != nil {
				log.Error().Err(err).Msg("error reading folder")
				continue
			}

			for _, file := range files {
				if strings.HasSuffix(file.Name(), "jpg") {
					os.Remove(path.Join(timelapse.Folder, file.Name()))
				}
			}
		}

		log.Debug().Msg("Finished encoding process")

		// Wait for ticker at end so it runs on startup
		<-ticker.C
	}
}
