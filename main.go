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
	const timelapseDuration = 24 * time.Hour
	const outputVideoName = "complete.mp4"
	ticker := time.NewTicker(timelapseDuration)

	for {
		log.Debug().Msg("Beginning encoding process")
		var inProgressTimelapses []models.Timelapse
		db.Where("status IN ?", []string{"Encoding", "InProgress"}).Find(&inProgressTimelapses)

		for _, timelapse := range inProgressTimelapses {
			// Check to not encode one that is in progress within 24 hours plus 5 minutes for a little bit of wiggle room
			if timelapse.StartDate.After(time.Now().Add(-timelapseDuration).Add(5 * time.Minute)) {
				continue
			}

			log.Debug().Msgf("Starting encoding for %d", timelapse.Id)
			timelapse.Status = "Encoding"
			db.Save(timelapse)

			var ffmpegArgs []string

			ffmpegArgs = append(ffmpegArgs, "-y")
			ffmpegArgs = append(ffmpegArgs, "-f")
			ffmpegArgs = append(ffmpegArgs, "image2")
			ffmpegArgs = append(ffmpegArgs, "-r")
			ffmpegArgs = append(ffmpegArgs, "3")
			ffmpegArgs = append(ffmpegArgs, "-i")
			ffmpegArgs = append(ffmpegArgs, path.Join(timelapse.Folder, "%05d.jpg"))
			ffmpegArgs = append(ffmpegArgs, path.Join(timelapse.Folder, outputVideoName))

			cmd := exec.Command("ffmpeg", ffmpegArgs...)

			stdout, err := cmd.CombinedOutput()
			if err != nil {
				log.Error().Str("encode_step", "exec").Err(err).Msgf("%s", stdout)
				continue
			}

			log.Debug().Str("working_directory", timelapse.Folder).Str("command", fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))).Msgf("%s", stdout)

			timelapse.EndDate = time.Now()
			timelapse.VideoFile = outputVideoName
			timelapse.Status = "Complete"
			db.Save(timelapse)

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
