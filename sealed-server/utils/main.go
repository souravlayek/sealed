package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/souravlayek/sealed-server/pkg/logger"
)

func LoadEnv() error {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Println("Unable to load env", err)
		return err
	}

	logger.Logger.Infow("ENV LOADED")
	return nil
}
