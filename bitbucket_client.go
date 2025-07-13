package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BitbucketClient struct {
	BaseURL   string
	Username  string
	Password  string
	Cloud     bool
	Workspace string // for Bitbucket Cloud
}

func NewBitbucketClient(cfg *Config, cloud bool) *BitbucketClient {
	workspace := cfg.Username // default fallback
	if cloud && cfg.Workspace != "" {
		workspace = cfg.Workspace
	}
	return &BitbucketClient{
		BaseURL:   cfg.BitbucketURL,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Cloud:     cloud,
		Workspace: workspace,
	}
}

// List all projects, then all repos per project for Bitbucket Cloud
func (c *BitbucketClient) GetAllCloudRepos() ([]struct{ ProjectKey, RepoSlug, RepoName string }, error) {
	var allRepos []struct{ ProjectKey, RepoSlug, RepoName string }
	if c.Cloud {
		// 1. List all projects
		projURL := "https://api.bitbucket.org/2.0/workspaces/" + c.Workspace + "/projects?pagelen=100"
		for {
			req, _ := http.NewRequest("GET", projURL, nil)
			req.SetBasicAuth(c.Username, c.Password)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != 200 {
				return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
			}
			var projData struct {
				Values []struct {
					Key string `json:"key"`
				} `json:"values"`
				Next string `json:"next"`
			}
			if err := json.Unmarshal(body, &projData); err != nil {
				return nil, err
			}
			for _, project := range projData.Values {
				if project.Key == "" {
					continue
				}
				// 2. For each project, list all repos
				repoURL := "https://api.bitbucket.org/2.0/repositories/" + c.Workspace + "?q=project.key=\"" + project.Key + "\"&pagelen=100"
				for {
					reqR, _ := http.NewRequest("GET", repoURL, nil)
					reqR.SetBasicAuth(c.Username, c.Password)
					respR, err := http.DefaultClient.Do(reqR)
					if err != nil {
						return nil, err
					}
					bodyR, _ := io.ReadAll(respR.Body)
					respR.Body.Close()
					if respR.StatusCode != 200 {
						return nil, fmt.Errorf("unexpected status: %d", respR.StatusCode)
					}
					var repoData struct {
						Values []struct {
							Slug string `json:"slug"`
							Name string `json:"name"`
						} `json:"values"`
						Next string `json:"next"`
					}
					if err := json.Unmarshal(bodyR, &repoData); err != nil {
						return nil, err
					}
					for _, repo := range repoData.Values {
						allRepos = append(allRepos, struct{ ProjectKey, RepoSlug, RepoName string }{project.Key, repo.Slug, repo.Name})
					}
					if repoData.Next == "" {
						break
					}
					repoURL = repoData.Next
				}
			}
			if projData.Next == "" {
				break
			}
			projURL = projData.Next
		}
	}
	return allRepos, nil
}

func (c *BitbucketClient) GetRepositoryCount() (int, error) {
	if c.Cloud {
		// For Bitbucket Cloud, count all repos under workspace/projects
		allRepos, err := c.GetAllCloudRepos()
		if err != nil {
			return 0, err
		}
		return len(allRepos), nil
	}
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
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Size  int `json:"size"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}
	return data.Total, nil
}

func (c *BitbucketClient) GetOpenPullRequestCount() (int, error) {
	if c.Cloud {
		allRepos, err := c.GetAllCloudRepos()
		if err != nil {
			return 0, err
		}
		totalPRs := 0
		for _, repo := range allRepos {
			prURL := "https://api.bitbucket.org/2.0/repositories/" + c.Workspace + "/" + repo.RepoSlug + "/pullrequests?state=OPEN&pagelen=1"
			reqPR, _ := http.NewRequest("GET", prURL, nil)
			reqPR.SetBasicAuth(c.Username, c.Password)
			respPR, err := http.DefaultClient.Do(reqPR)
			if err != nil {
				return 0, err
			}
			defer respPR.Body.Close()
			if respPR.StatusCode != 200 {
				return 0, fmt.Errorf("unexpected status: %d", respPR.StatusCode)
			}
			bodyPR, _ := io.ReadAll(respPR.Body)
			var prData struct {
				Size int `json:"size"`
			}
			if err := json.Unmarshal(bodyPR, &prData); err != nil {
				return 0, err
			}
			totalPRs += prData.Size
		}
		return totalPRs, nil
	}
	url := c.BaseURL + "/rest/api/1.0/pull-requests?state=OPEN&limit=1"
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
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Size  int `json:"size"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}
	return data.Total, nil
}

func (c *BitbucketClient) GetUserCount() (int, error) {
	if c.Cloud {
		// Bitbucket Cloud: https://api.bitbucket.org/2.0/workspaces/{workspace}/members
		url := "https://api.bitbucket.org/2.0/workspaces/" + c.Workspace + "/members?pagelen=1"
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
		body, _ := io.ReadAll(resp.Body)
		var data struct {
			Size   int           `json:"size"`
			Values []interface{} `json:"values"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			return 0, err
		}
		return data.Size, nil
	}
	url := c.BaseURL + "/rest/api/1.0/users?limit=1"
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
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Size  int `json:"size"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}
	return data.Total, nil
}

func (c *BitbucketClient) GetProjectCount() (int, error) {
	if c.Cloud {
		projURL := "https://api.bitbucket.org/2.0/workspaces/" + c.Workspace + "/projects?pagelen=100"
		projectKeys := make(map[string]bool)
		for {
			req, _ := http.NewRequest("GET", projURL, nil)
			req.SetBasicAuth(c.Username, c.Password)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return 0, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
			}
			body, _ := io.ReadAll(resp.Body)
			var projData struct {
				Values []struct {
					Key string `json:"key"`
				} `json:"values"`
				Next string `json:"next"`
			}
			if err := json.Unmarshal(body, &projData); err != nil {
				return 0, err
			}
			for _, proj := range projData.Values {
				projectKeys[proj.Key] = true
			}
			if projData.Next == "" {
				break
			}
			projURL = projData.Next
		}
		return len(projectKeys), nil
	}
	url := c.BaseURL + "/rest/api/1.0/projects?limit=1"
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
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Size  int `json:"size"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}
	return data.Total, nil
}

// New: Get commit count and top committer for each repo in Bitbucket Cloud
func (c *BitbucketClient) GetRepoCommitStats() (map[string]int, map[string]map[string]int, error) {
	commitCounts := make(map[string]int)
	committers := make(map[string]map[string]int)
	if c.Cloud {
		allRepos, err := c.GetAllCloudRepos()
		if err != nil {
			return nil, nil, err
		}
		for _, repo := range allRepos {
			commitsURL := "https://api.bitbucket.org/2.0/repositories/" + c.Workspace + "/" + repo.RepoSlug + "/commits?pagelen=100"
			totalCommits := 0
			committerMap := make(map[string]int)
			for {
				reqC, _ := http.NewRequest("GET", commitsURL, nil)
				reqC.SetBasicAuth(c.Username, c.Password)
				respC, err := http.DefaultClient.Do(reqC)
				if err != nil {
					return nil, nil, err
				}
				defer respC.Body.Close()
				if respC.StatusCode != 200 {
					return nil, nil, fmt.Errorf("unexpected status: %d", respC.StatusCode)
				}
				bodyC, _ := io.ReadAll(respC.Body)
				var commitData struct {
					Values []struct {
						Author struct {
							Raw string `json:"raw"`
						} `json:"author"`
					} `json:"values"`
					Next string `json:"next"`
				}
				if err := json.Unmarshal(bodyC, &commitData); err != nil {
					return nil, nil, err
				}
				totalCommits += len(commitData.Values)
				for _, commit := range commitData.Values {
					committerMap[commit.Author.Raw]++
				}
				if commitData.Next == "" {
					break
				}
				commitsURL = commitData.Next
			}
			commitCounts[repo.ProjectKey+"/"+repo.RepoName] = totalCommits
			committers[repo.ProjectKey+"/"+repo.RepoName] = committerMap
		}
	}
	return commitCounts, committers, nil
}
