package utils

import (
	"github.com/joho/godotenv"
	"github.com/souravlayek/sealed-server/pkg/logger"
)

func LoadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}

	logger.Logger.Infow("ENV LOADED")
	return nil
}
