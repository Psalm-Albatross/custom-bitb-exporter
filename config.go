package main

import (
	"os"
)

type Config struct {
	BitbucketURL string
	Username     string
	Password     string
	Workspace    string // for Bitbucket Cloud
}

func LoadConfig() (*Config, error) {
	return &Config{
		BitbucketURL: os.Getenv("BITBUCKET_URL"),
		Username:     os.Getenv("BITBUCKET_USERNAME"),
		Password:     os.Getenv("BITBUCKET_PASSWORD"),
		Workspace:    os.Getenv("BITBUCKET_WORKSPACE"),
	}, nil
}
