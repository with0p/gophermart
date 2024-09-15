package config

import (
	"flag"
	"os"
)

// const defaultDataBaseAddress = "host=localhost port=5435 user=postgres password=1234 dbname=postgres sslmode=disable"
const defaultBaseURL = "localhost:8080"

// const defaultAccrualURL = "localhost:8081"
const defaultAccrualURL = ""
const defaultDataBaseAddress = ""

type Config struct {
	BaseURL         string
	AccrualURL      string
	DataBaseAddress string
}

var configuration *Config

func GetConfig() *Config {
	if configuration == nil {
		var conf = Config{}

		flag.StringVar(&conf.BaseURL, "a", defaultBaseURL, "RUN_ADDRESS")
		flag.StringVar(&conf.DataBaseAddress, "d", defaultDataBaseAddress, "DATABASE_URI")
		flag.StringVar(&conf.AccrualURL, "r", defaultAccrualURL, "ACCRUAL_SYSTEM_ADDRESS")
		flag.Parse()

		if envServerAddress := os.Getenv("RUN_ADDRESS"); envServerAddress != "" {
			conf.BaseURL = envServerAddress
		}

		if envDatabaseStoragePath := os.Getenv("DATABASE_URI"); envDatabaseStoragePath != "" {
			conf.DataBaseAddress = envDatabaseStoragePath
		}

		if envAccrualAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddress != "" {
			conf.AccrualURL = envAccrualAddress
		}

		configuration = &Config{
			BaseURL:         conf.BaseURL,
			DataBaseAddress: conf.DataBaseAddress,
			AccrualURL:      conf.AccrualURL,
		}
	}

	return configuration

}
