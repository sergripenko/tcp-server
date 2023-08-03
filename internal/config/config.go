// Package config - contains config struct and methods to load config from file.
package config

import (
	"strconv"

	"github.com/joho/godotenv"
)

// Config - config struct. Contains all config fields.
type Config struct {
	ServerHost            string
	ServerPort            int64
	HashcashZeros         int
	HashcashDuration      int64
	HashcashMaxIterations int
}

// LoadConfig - loads config from file on path, after that gets env-variables (overrides values from file).
func LoadConfig(confPath string) (*Config, error) {
	var myEnvs map[string]string

	myEnvs, err := godotenv.Read(confPath)
	if err != nil {
		return nil, err
	}

	serverPort, err := strconv.ParseInt(myEnvs["SERVER_PORT"], 10, 64)
	if err != nil {
		return nil, err
	}

	hashcashZeros, err := strconv.Atoi(myEnvs["HASHCASH_ZEROS"])
	if err != nil {
		return nil, err
	}

	hashcashDuration, err := strconv.ParseInt(myEnvs["HASHCASH_DURATION"], 10, 64)
	if err != nil {
		return nil, err
	}

	hashcashMaxIterations, err := strconv.Atoi(myEnvs["HASHCASH_MAX_ITERATIONS"])
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerHost:            myEnvs["SERVER_HOST"],
		ServerPort:            serverPort,
		HashcashZeros:         hashcashZeros,
		HashcashDuration:      hashcashDuration,
		HashcashMaxIterations: hashcashMaxIterations,
	}, nil
}
