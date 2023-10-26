package main

import (
	"flag"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/neboman11/timelapse-manager/api"
	_ "github.com/neboman11/timelapse-manager/docs"
)

// Folder to search for config file
const configFileLocation = "."

// Name of config file (without extension)
const configFileName = "configuration"

// Type of config file
const configFileExtension = "toml"

func main() {
	log.Info().Msgf("API version: %s", CurrentVersionNumber)
	readConfigFile()
	log.Trace().Msg("Successfully loaded config")

	portPtr := flag.Int("p", 3001, "Port to listen for requests on.")
	flag.Parse()

	// Verify the given port is a valid port number
	if *portPtr < 1 || *portPtr > 65535 {
		log.Fatal().Msg("Invalid port specified.")
	}

	startDaemon(*portPtr)
}

// Read in config values to viper and check that necessary values are set
func readConfigFile() {
	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileExtension)
	viper.AddConfigPath(configFileLocation)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found ignore error and use defaults
			log.Info().Msg("Configuration file not found, using defaults.")
		} else {
			log.Fatal().Msg(err.Error())
		}
	}
}

// Setup IPFS and start listening for requests
func startDaemon(port int) {
	open_database()
	log.Info().Msg("Ready for requests")
	api.HandleRequests(port, db)
}
