package appconfig

import (
	"log"

	"github.com/arifin2018/splitbill-arifin.git/config"
	"github.com/joho/godotenv"
)

func initEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config.DB = config.PostgresSQL()
}
