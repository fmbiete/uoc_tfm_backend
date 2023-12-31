package main

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"tfm_backend/api"
	"tfm_backend/models"
	"tfm_backend/orm"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var cfg models.Config
	raw, err := os.ReadFile("config.json")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		return
	}
	json.Unmarshal(raw, &cfg)

	database := orm.NewDatabase(&cfg)
	server := api.NewServer(cfg.Server, database)

	err = database.Setup()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize database")
		return
	}

	err = server.Listen()
	if err != nil {
		log.Error().Err(err).Msg("Faile to listen REST API")
		return
	}
}
