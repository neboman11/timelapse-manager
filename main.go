package main

import (
	"flag"
	"os"
	"os/exec"
	"path"

	"github.com/rs/zerolog/log"

	"github.com/neboman11/timelapse-manager/api"
	_ "github.com/neboman11/timelapse-manager/docs"
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

	go encodeInProgress()

	startDaemon(*portPtr)
}

// Setup IPFS and start listening for requests
func startDaemon(port int) {
	open_database()
	log.Info().Msg("Ready for requests")
	api.HandleRequests(port, db)
}

func encodeInProgress() {
	baseInProgressFolder := os.Getenv("BASE_INPROGRESS_FOLDER")
	workingEncodingFolder := path.Join(baseInProgressFolder, "60e849a9-04c9-4edb-9204-d830f64d5cd0")

	var ffmpegArgs []string

	ffmpegArgs = append(ffmpegArgs, "-f")
	ffmpegArgs = append(ffmpegArgs, "image2")
	ffmpegArgs = append(ffmpegArgs, "-r")
	ffmpegArgs = append(ffmpegArgs, "1")
	ffmpegArgs = append(ffmpegArgs, "-i")
	ffmpegArgs = append(ffmpegArgs, path.Join(workingEncodingFolder, "%05d.jpg"))
	ffmpegArgs = append(ffmpegArgs, path.Join(workingEncodingFolder, "test.avi"))

	cmd := exec.Command("ffmpeg", ffmpegArgs...)

	stdout, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal().Err(err)
	}

	log.Info().Msgf("%s", stdout)
}
