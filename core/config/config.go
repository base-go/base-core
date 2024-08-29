package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBPath        string `envconfig:"DB_PATH" default:"./sqlite.db"`
	ServerAddress string `envconfig:"SERVER_ADDRESS" default:":8080"`
	JWTSecret     string `envconfig:"JWT_SECRET" required:"true"`
}

var AppConfig Config

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	err = envconfig.Process("", &AppConfig)
	if err != nil {
		log.Fatal("Cannot process config:", err)
	}
}
