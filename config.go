package main

import (
	"os"
)

type Config struct {
	BitbucketURL string
	Username     string
	Password     string
}

func LoadConfig() (*Config, error) {
	return &Config{
		BitbucketURL: os.Getenv("BITBUCKET_URL"),
		Username:     os.Getenv("BITBUCKET_USERNAME"),
		Password:     os.Getenv("BITBUCKET_PASSWORD"),
	}, nil
}
