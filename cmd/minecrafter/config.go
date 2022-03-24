package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dateiexplorer/go-dotenv"
)

type config struct {
	CLIExecutable string
	GuildID       string
	Token         string
	ServerIP      string
	PaperHomePath string
	MaxAttempts   int
	WatchInterval time.Duration
}

func load() (*config, error) {
	config := &config{}

	err := dotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load dotenv: %w", err)
	}

	if val, ok := os.LookupEnv("CLI_EXECUTABLE"); ok {
		config.CLIExecutable = val
	} else {
		return nil, fmt.Errorf("cannot find CLI_EXECUTABLE env var")
	}

	if val, ok := os.LookupEnv("PAPER_HOME"); ok {
		config.PaperHomePath = val
	} else {
		return nil, fmt.Errorf("cannot find PAPER_HOME env var")
	}

	if val, ok := os.LookupEnv("GUILD_ID"); ok {
		config.GuildID = val
	} else {
		return nil, fmt.Errorf("cannot find GUILD_ID env var")
	}

	if val, ok := os.LookupEnv("DISCORD_TOKEN"); ok {
		config.Token = val
	} else {
		return nil, fmt.Errorf("cannot find DISCORD_TOKEN env var")
	}

	if val, ok := os.LookupEnv("SERVER_IP"); ok {
		config.ServerIP = val
	} else {
		return nil, fmt.Errorf("cannot find SERVER_IP env var")
	}

	if val, ok := os.LookupEnv("MAX_ATTEMPTS"); ok {
		config.MaxAttempts, err = strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("cannot parse value of MAX_ATTEMPTS '%v' to int", val)
		}
	} else {
		config.MaxAttempts = 3
	}

	if val, ok := os.LookupEnv("WATCH_INTERVAL"); ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("cannot parse value of WATCH_INTERVAL '%v' to time.Duration", val)
		}
		config.WatchInterval = time.Duration(i) * time.Second
	} else {
		config.WatchInterval = 60 * time.Second
	}

	return config, nil
}
