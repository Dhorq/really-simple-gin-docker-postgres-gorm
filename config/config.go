package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	JWT_SECRET         string
	JWT_REFRESH_SECRET string
}

func LoadEnv() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:               os.Getenv("PORT"),
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPass:             os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_NAME"),
		JWT_SECRET:         os.Getenv("JWT_SECRET"),
		JWT_REFRESH_SECRET: os.Getenv("JWT_REFRESH_SECRET"),
	}, nil
}
