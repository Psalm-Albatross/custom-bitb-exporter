package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type BitbucketClient struct {
	BaseURL  string
	Username string
	Password string
}

func NewBitbucketClient(cfg *Config) *BitbucketClient {
	return &BitbucketClient{
		BaseURL:  cfg.BitbucketURL,
		Username: cfg.Username,
		Password: cfg.Password,
	}
}

func (c *BitbucketClient) GetRepositoryCount() (int, error) {
	url := fmt.Sprintf("%s/rest/api/1.0/repos?limit=1", c.BaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.Username, c.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var data struct {
		Size  int `json:"size"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}
	return data.Total, nil
}
