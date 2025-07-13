package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type BitbucketCollector struct {
	client *BitbucketClient
	// Core metrics
	repoCount    *prometheus.Desc
	prCount      *prometheus.Desc
	userCount    *prometheus.Desc
	projectCount *prometheus.Desc
	// Project/Repo metrics
	perProjectRepos   *prometheus.Desc
	perRepoCommits    *prometheus.Desc
	perRepoPRs        *prometheus.Desc
	perRepoSize       *prometheus.Desc
	perRepoLastCommit *prometheus.Desc
	// PR metrics
	perRepoOpenPRs   *prometheus.Desc
	prMergedTotal    *prometheus.Desc
	prAgeSeconds     *prometheus.Desc
	prReviewersTotal *prometheus.Desc
	// Commit/Author metrics
	perUserCommits   *prometheus.Desc
	commitAgeSeconds *prometheus.Desc
	// User/Access metrics
	perUserAccessRepos *prometheus.Desc
	teamMembersTotal   *prometheus.Desc
	// Pipeline/Build metrics
	pipelineRunsTotal       *prometheus.Desc
	pipelineDurationSeconds *prometheus.Desc
	pipelineActiveTotal     *prometheus.Desc
	// Branch/Policy metrics
	branchRestrictionsTotal     *prometheus.Desc
	branchDefaultPolicyEnforced *prometheus.Desc
	// Webhook metrics
	webhooksTotal                  *prometheus.Desc
	webhookFailuresTotal           *prometheus.Desc
	webhookDeliveryDurationSeconds *prometheus.Desc
	// API/Exporter health
	apiRateLimitRemaining    *prometheus.Desc
	apiRateLimitResetSeconds *prometheus.Desc
	exporterUp               *prometheus.Desc
	exporterErrorsTotal      *prometheus.Desc
	// Tags/Releases/Issues
	tagsTotal     *prometheus.Desc
	issuesTotal   *prometheus.Desc
	releasesTotal *prometheus.Desc
	logLevel      string
}

func NewBitbucketCollector(client *BitbucketClient, logLevel string) *BitbucketCollector {
	return &BitbucketCollector{
		client:                         client,
		repoCount:                      prometheus.NewDesc("bitbucket_repository_count", "Total number of repositories", nil, nil),
		prCount:                        prometheus.NewDesc("bitbucket_open_pull_requests", "Total number of open pull requests", nil, nil),
		userCount:                      prometheus.NewDesc("bitbucket_user_count", "Total number of users", nil, nil),
		projectCount:                   prometheus.NewDesc("bitbucket_project_count", "Total number of projects", nil, nil),
		perProjectRepos:                prometheus.NewDesc("bitbucket_project_repos", "Number of repositories per project", []string{"project_key", "project_name", "project_uuid", "project_type", "project_is_private", "project_created_on", "project_updated_on", "project_has_publicly_visible_repos"}, nil),
		perRepoCommits:                 prometheus.NewDesc("bitbucket_repo_commits", "Number of commits per repo", []string{"project_key", "project_name", "repo_slug", "repo_name"}, nil),
		perRepoPRs:                     prometheus.NewDesc("bitbucket_repo_open_prs", "Number of open PRs per repo", []string{"project_key", "project_name", "repo_slug", "repo_name"}, nil),
		perRepoSize:                    prometheus.NewDesc("bitbucket_repo_size_bytes", "Size of each repository in bytes", []string{"project_key", "project_name", "repo_slug", "repo_name"}, nil),
		perRepoLastCommit:              prometheus.NewDesc("bitbucket_repo_last_commit_timestamp", "Unix timestamp of last commit in repo", []string{"project_key", "project_name", "repo_slug", "repo_name"}, nil),
		perRepoOpenPRs:                 prometheus.NewDesc("bitbucket_repo_open_prs", "Number of open PRs per repository", []string{"project_key", "project_name", "repo_slug", "repo_name"}, nil),
		prMergedTotal:                  prometheus.NewDesc("bitbucket_pull_requests_merged_total", "Cumulative number of merged pull requests", nil, nil),
		prAgeSeconds:                   prometheus.NewDesc("bitbucket_pull_request_age_seconds", "Age of each PR in seconds", []string{"project_key", "repo_slug", "pr_id", "state"}, nil),
		prReviewersTotal:               prometheus.NewDesc("bitbucket_pull_request_reviewers_total", "Number of reviewers per PR", []string{"project_key", "repo_slug", "pr_id"}, nil),
		perUserCommits:                 prometheus.NewDesc("bitbucket_user_commits", "Number of commits per user per repo", []string{"project_key", "project_name", "repo_slug", "repo_name", "user"}, nil),
		commitAgeSeconds:               prometheus.NewDesc("bitbucket_commit_age_seconds", "Age of commits in seconds (latest only)", []string{"repo_slug"}, nil),
		perUserAccessRepos:             prometheus.NewDesc("bitbucket_user_access_repos_total", "Number of repositories accessed per user", []string{"user", "permission_level"}, nil),
		teamMembersTotal:               prometheus.NewDesc("bitbucket_team_members_total", "Number of members per team (Bitbucket Cloud)", []string{"team_name"}, nil),
		pipelineRunsTotal:              prometheus.NewDesc("bitbucket_pipeline_runs_total", "Number of pipeline runs by result", []string{"repo_slug", "result"}, nil),
		pipelineDurationSeconds:        prometheus.NewDesc("bitbucket_pipeline_duration_seconds", "Duration of each pipeline run", []string{"repo_slug"}, nil),
		pipelineActiveTotal:            prometheus.NewDesc("bitbucket_pipeline_active_total", "Currently running pipelines", []string{"repo_slug"}, nil),
		branchRestrictionsTotal:        prometheus.NewDesc("bitbucket_branch_restrictions_total", "Number of branch restrictions per branch", []string{"project_key", "repo_slug", "branch_name", "restriction_type"}, nil),
		branchDefaultPolicyEnforced:    prometheus.NewDesc("bitbucket_branch_default_policy_enforced", "Whether default branch has policy enforced (0/1)", []string{"repo_slug", "branch_name"}, nil),
		webhooksTotal:                  prometheus.NewDesc("bitbucket_webhooks_total", "Total number of webhooks configured", []string{"repo_slug", "status"}, nil),
		webhookFailuresTotal:           prometheus.NewDesc("bitbucket_webhook_failures_total", "Number of webhook failures", []string{"repo_slug", "event_type", "endpoint"}, nil),
		webhookDeliveryDurationSeconds: prometheus.NewDesc("bitbucket_webhook_delivery_duration_seconds", "Duration of webhook delivery", []string{"repo_slug", "event_type"}, nil),
		apiRateLimitRemaining:          prometheus.NewDesc("bitbucket_api_rate_limit_remaining", "Remaining API rate limit (Cloud)", nil, nil),
		apiRateLimitResetSeconds:       prometheus.NewDesc("bitbucket_api_rate_limit_reset_seconds", "Time in seconds until rate limit reset", nil, nil),
		exporterUp:                     prometheus.NewDesc("bitbucket_exporter_up", "Whether the Bitbucket exporter is running successfully", nil, nil),
		exporterErrorsTotal:            prometheus.NewDesc("bitbucket_exporter_errors_total", "Total number of errors in exporter", []string{"error_type", "component"}, nil),
		tagsTotal:                      prometheus.NewDesc("bitbucket_tags_total", "Number of Git tags in repository", []string{"repo_slug"}, nil),
		issuesTotal:                    prometheus.NewDesc("bitbucket_issues_total", "Number of open issues (Cloud only, if enabled)", []string{"repo_slug", "status"}, nil),
		releasesTotal:                  prometheus.NewDesc("bitbucket_releases_total", "Number of releases per repository (if supported)", []string{"repo_slug"}, nil),
		logLevel:                       logLevel,
	}
}

func (c *BitbucketCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.repoCount
	ch <- c.prCount
	ch <- c.userCount
	ch <- c.projectCount
	ch <- c.perRepoCommits
	ch <- c.perRepoPRs
	ch <- c.perProjectRepos
	ch <- c.perUserCommits
}

func (c *BitbucketCollector) Collect(ch chan<- prometheus.Metric) {
	// This collector supports both Bitbucket Cloud (-cloud=true) and Data Center/Server (-cloud=false, default).
	// Cloud mode uses Bitbucket Cloud 2.0 API and emits all advanced metrics.
	// Data Center/Server mode uses the 1.0 API and only emits metrics supported by Server.
	logf := func(format string, v ...interface{}) {
		if c.logLevel == "debug" {
			log.Printf(format, v...)
		}
	}

	var exporterUpValue float64 = 1.0
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] exporter recovered: %v", r)
			exporterUpValue = 0
		}
		ch <- prometheus.MustNewConstMetric(c.exporterUp, prometheus.GaugeValue, exporterUpValue)
	}()

	repoCount := 0
	prCount, err := c.client.GetOpenPullRequestCount()
	if err != nil {
		log.Printf("error collecting open PR count: %v", err)
	} else {
		logf("open PR count: %d", prCount)
		ch <- prometheus.MustNewConstMetric(c.prCount, prometheus.GaugeValue, float64(prCount))
	}

	userCount, err := c.client.GetUserCount()
	if err != nil {
		log.Printf("error collecting user count: %v", err)
	} else {
		logf("user count: %d", userCount)
		ch <- prometheus.MustNewConstMetric(c.userCount, prometheus.GaugeValue, float64(userCount))
	}

	if c.client.Cloud {
		log.Println("Fetching all projects from Bitbucket Cloud API...")
		projectLabels := make(map[string]struct {
			Name                    string
			UUID                    string
			Type                    string
			IsPrivate               string
			CreatedOn               string
			UpdatedOn               string
			HasPubliclyVisibleRepos string
		})
		projectCount := 0
		projURL := "https://api.bitbucket.org/2.0/workspaces/" + c.client.Workspace + "/projects?pagelen=100"
		for {
			req, _ := http.NewRequest("GET", projURL, nil)
			req.SetBasicAuth(c.client.Username, c.client.Password)
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != 200 {
				log.Printf("Failed to fetch projects: %v, status: %v", err, resp.StatusCode)
				break
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var projData struct {
				Values []struct {
					Key                     string `json:"key"`
					Name                    string `json:"name"`
					UUID                    string `json:"uuid"`
					Type                    string `json:"type"`
					IsPrivate               bool   `json:"is_private"`
					CreatedOn               string `json:"created_on"`
					UpdatedOn               string `json:"updated_on"`
					HasPubliclyVisibleRepos bool   `json:"has_publicly_visible_repos"`
				} `json:"values"`
				Next string `json:"next"`
			}
			if err := json.Unmarshal(body, &projData); err != nil {
				log.Printf("Failed to unmarshal project data: %v", err)
				break
			}
			for _, p := range projData.Values {
				if p.Key == "" {
					continue
				}
				projectLabels[p.Key] = struct {
					Name                    string
					UUID                    string
					Type                    string
					IsPrivate               string
					CreatedOn               string
					UpdatedOn               string
					HasPubliclyVisibleRepos string
				}{
					Name:                    p.Name,
					UUID:                    p.UUID,
					Type:                    p.Type,
					IsPrivate:               boolToString(p.IsPrivate),
					CreatedOn:               p.CreatedOn,
					UpdatedOn:               p.UpdatedOn,
					HasPubliclyVisibleRepos: boolToString(p.HasPubliclyVisibleRepos),
				}
				logf("Found project: key=%s, name=%s, uuid=%s, type=%s, is_private=%v, created_on=%s, updated_on=%s, has_publicly_visible_repos=%v", p.Key, p.Name, p.UUID, p.Type, p.IsPrivate, p.CreatedOn, p.UpdatedOn, p.HasPubliclyVisibleRepos)
			}
			projectCount += len(projData.Values)
			if projData.Next == "" {
				break
			}
			projURL = projData.Next
		}
		log.Printf("Total projects found: %d", projectCount)
		ch <- prometheus.MustNewConstMetric(c.projectCount, prometheus.GaugeValue, float64(projectCount))

		log.Println("Fetching all repositories from Bitbucket Cloud API...")
		repoURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "?pagelen=100"
		projectRepoCount := make(map[string]int)
		var allRepos []struct {
			ProjectKey  string
			ProjectName string
			RepoSlug    string
			RepoName    string
		}
		for {
			req, _ := http.NewRequest("GET", repoURL, nil)
			req.SetBasicAuth(c.client.Username, c.client.Password)
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != 200 {
				log.Printf("Failed to fetch repos: %v, status: %v", err, resp.StatusCode)
				break
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var repoData struct {
				Values []struct {
					Slug    string `json:"slug"`
					Name    string `json:"name"`
					Project struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					} `json:"project"`
				} `json:"values"`
				Next string `json:"next"`
			}
			if err := json.Unmarshal(body, &repoData); err != nil {
				log.Printf("Failed to unmarshal repo data: %v", err)
				break
			}
			for _, repo := range repoData.Values {
				if repo.Project.Key == "" {
					continue
				}
				projectRepoCount[repo.Project.Key]++
				allRepos = append(allRepos, struct {
					ProjectKey  string
					ProjectName string
					RepoSlug    string
					RepoName    string
				}{repo.Project.Key, repo.Project.Name, repo.Slug, repo.Name})
				logf("Found repo: project_key=%s, project_name=%s, repo_slug=%s, repo_name=%s", repo.Project.Key, repo.Project.Name, repo.Slug, repo.Name)
			}
			if repoData.Next == "" {
				break
			}
			repoURL = repoData.Next
		}
		repoCount = len(allRepos)
		log.Printf("Total repositories found: %d", repoCount)
		ch <- prometheus.MustNewConstMetric(c.repoCount, prometheus.GaugeValue, float64(repoCount))
		for projectKey, count := range projectRepoCount {
			p := projectLabels[projectKey]
			ch <- prometheus.MustNewConstMetric(
				c.perProjectRepos, prometheus.GaugeValue, float64(count), projectKey, p.Name, p.UUID, p.Type, p.IsPrivate, p.CreatedOn, p.UpdatedOn, p.HasPubliclyVisibleRepos)
		}
		// Per-repo PRs and commits, and per-user commits (aggregate first)
		for _, repo := range allRepos {
			// Open PRs per repo
			prURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/pullrequests?state=OPEN&pagelen=1"
			reqPR, _ := http.NewRequest("GET", prURL, nil)
			reqPR.SetBasicAuth(c.client.Username, c.client.Password)
			respPR, err := http.DefaultClient.Do(reqPR)
			prCount := 0
			if err == nil && respPR.StatusCode == 200 {
				bodyPR, _ := io.ReadAll(respPR.Body)
				var prData struct {
					Size int `json:"size"`
				}
				if err := json.Unmarshal(bodyPR, &prData); err == nil {
					prCount = prData.Size
				}
				respPR.Body.Close()
			}
			ch <- prometheus.MustNewConstMetric(
				c.perRepoPRs, prometheus.GaugeValue, float64(prCount), repo.ProjectKey, repo.ProjectName, repo.RepoSlug, repo.RepoName)

			// Commits per repo and user (aggregate before emitting)
			commitsURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/commits?pagelen=100"
			totalCommits := 0
			committerMap := make(map[string]int)
			for {
				reqC, _ := http.NewRequest("GET", commitsURL, nil)
				reqC.SetBasicAuth(c.client.Username, c.client.Password)
				respC, err := http.DefaultClient.Do(reqC)
				if err != nil || respC.StatusCode != 200 {
					break
				}
				bodyC, _ := io.ReadAll(respC.Body)
				respC.Body.Close()
				var commitData struct {
					Values []struct {
						Author struct {
							Raw string `json:"raw"`
						} `json:"author"`
					} `json:"values"`
					Next string `json:"next"`
				}
				if err := json.Unmarshal(bodyC, &commitData); err != nil {
					break
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
			ch <- prometheus.MustNewConstMetric(
				c.perRepoCommits, prometheus.GaugeValue, float64(totalCommits), repo.ProjectKey, repo.ProjectName, repo.RepoSlug, repo.RepoName)
			for user, count := range committerMap {
				ch <- prometheus.MustNewConstMetric(
					c.perUserCommits, prometheus.GaugeValue, float64(count), repo.ProjectKey, repo.ProjectName, repo.RepoSlug, repo.RepoName, user)
			}
		}

		// Per-repo size and last commit, issues, releases, branch count, webhooks, branch restrictions
		for _, repo := range allRepos {
			// Per-repo size and last commit
			repoInfoURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug
			reqInfo, _ := http.NewRequest("GET", repoInfoURL, nil)
			reqInfo.SetBasicAuth(c.client.Username, c.client.Password)
			respInfo, err := http.DefaultClient.Do(reqInfo)
			if respInfo != nil {
				defer respInfo.Body.Close()
			}
			if err != nil || respInfo == nil || respInfo.StatusCode != 200 {
				log.Printf("Failed to fetch repo info for %s: %v, status: %v", repo.RepoSlug, err, statusCodeSafe(respInfo))
				exporterUpValue = 0
				continue
			}
			bodyInfo, _ := io.ReadAll(respInfo.Body)
			var repoInfo struct {
				Size int64 `json:"size"`
			}
			if err := json.Unmarshal(bodyInfo, &repoInfo); err != nil {
				log.Printf("Failed to unmarshal repo info for %s: %v", repo.RepoSlug, err)
				exporterUpValue = 0
				continue
			}
			ch <- prometheus.MustNewConstMetric(
				c.perRepoSize, prometheus.GaugeValue, float64(repoInfo.Size), repo.ProjectKey, repo.ProjectName, repo.RepoSlug, repo.RepoName)

			// Last commit timestamp
			commitsURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/commits?pagelen=1"
			reqLast, _ := http.NewRequest("GET", commitsURL, nil)
			reqLast.SetBasicAuth(c.client.Username, c.client.Password)
			respLast, err := http.DefaultClient.Do(reqLast)
			if respLast != nil {
				defer respLast.Body.Close()
			}
			if err != nil || respLast == nil || respLast.StatusCode != 200 {
				log.Printf("Failed to fetch last commit for %s: %v, status: %v", repo.RepoSlug, err, statusCodeSafe(respLast))
				exporterUpValue = 0
				continue
			}
			bodyLast, _ := io.ReadAll(respLast.Body)
			var commitData struct {
				Values []struct {
					Date string `json:"date"`
				} `json:"values"`
			}
			if err := json.Unmarshal(bodyLast, &commitData); err != nil {
				log.Printf("Failed to unmarshal last commit for %s: %v", repo.RepoSlug, err)
				exporterUpValue = 0
				continue
			}
			if len(commitData.Values) > 0 {
				ts, err := parseRFC3339ToUnix(commitData.Values[0].Date)
				if err == nil {
					ch <- prometheus.MustNewConstMetric(
						c.perRepoLastCommit, prometheus.GaugeValue, float64(ts), repo.ProjectKey, repo.ProjectName, repo.RepoSlug, repo.RepoName)
				} else {
					log.Printf("Failed to parse commit date for %s: %v", repo.RepoSlug, err)
					exporterUpValue = 0
				}
			}

			// Issues (open)
			issuesURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/issues?state=open"
			reqIssues, _ := http.NewRequest("GET", issuesURL, nil)
			reqIssues.SetBasicAuth(c.client.Username, c.client.Password)
			respIssues, err := http.DefaultClient.Do(reqIssues)
			if err == nil && respIssues.StatusCode == 200 {
				bodyIssues, _ := io.ReadAll(respIssues.Body)
				respIssues.Body.Close()
				var issuesData struct {
					Size int `json:"size"`
				}
				if err := json.Unmarshal(bodyIssues, &issuesData); err == nil {
					ch <- prometheus.MustNewConstMetric(
						c.issuesTotal, prometheus.GaugeValue, float64(issuesData.Size), repo.RepoSlug, "open")
				}
			}

			// Releases (tags)
			tagsURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/refs/tags?pagelen=100"
			reqTags, _ := http.NewRequest("GET", tagsURL, nil)
			reqTags.SetBasicAuth(c.client.Username, c.client.Password)
			respTags, err := http.DefaultClient.Do(reqTags)
			if err == nil && respTags.StatusCode == 200 {
				bodyTags, _ := io.ReadAll(respTags.Body)
				respTags.Body.Close()
				var tagsData struct {
					Size int `json:"size"`
				}
				if err := json.Unmarshal(bodyTags, &tagsData); err == nil {
					ch <- prometheus.MustNewConstMetric(
						c.releasesTotal, prometheus.GaugeValue, float64(tagsData.Size), repo.RepoSlug)
				}
			}

			// Branch count
			branchesURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/refs/branches?pagelen=100"
			reqBranches, _ := http.NewRequest("GET", branchesURL, nil)
			reqBranches.SetBasicAuth(c.client.Username, c.client.Password)
			respBranches, err := http.DefaultClient.Do(reqBranches)
			if err == nil && respBranches.StatusCode == 200 {
				bodyBranches, _ := io.ReadAll(respBranches.Body)
				respBranches.Body.Close()
				var branchesData struct {
					Size int `json:"size"`
				}
				if err := json.Unmarshal(bodyBranches, &branchesData); err == nil {
					ch <- prometheus.MustNewConstMetric(
						prometheus.NewDesc("bitbucket_repo_branches_total", "Total number of branches in repo", []string{"repo_slug"}, nil), prometheus.GaugeValue, float64(branchesData.Size), repo.RepoSlug)
				}
			}

			// Webhooks
			webhooksURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/hooks?pagelen=100"
			reqHooks, _ := http.NewRequest("GET", webhooksURL, nil)
			reqHooks.SetBasicAuth(c.client.Username, c.client.Password)
			respHooks, err := http.DefaultClient.Do(reqHooks)
			if err == nil && respHooks.StatusCode == 200 {
				bodyHooks, _ := io.ReadAll(respHooks.Body)
				respHooks.Body.Close()
				var hooksData struct {
					Size int `json:"size"`
				}
				if err := json.Unmarshal(bodyHooks, &hooksData); err == nil {
					ch <- prometheus.MustNewConstMetric(
						c.webhooksTotal, prometheus.GaugeValue, float64(hooksData.Size), repo.RepoSlug, "active")
				}
			}

			// Branch restrictions
			restrictionsURL := "https://api.bitbucket.org/2.0/repositories/" + c.client.Workspace + "/" + repo.RepoSlug + "/branch-restrictions?pagelen=100"
			reqRestrict, _ := http.NewRequest("GET", restrictionsURL, nil)
			reqRestrict.SetBasicAuth(c.client.Username, c.client.Password)
			respRestrict, err := http.DefaultClient.Do(reqRestrict)
			if err == nil && respRestrict.StatusCode == 200 {
				bodyRestrict, _ := io.ReadAll(respRestrict.Body)
				respRestrict.Body.Close()
				var restrictData struct {
					Values []struct {
						Branch string `json:"branch"`
						Type   string `json:"type"`
					} `json:"values"`
				}
				if err := json.Unmarshal(bodyRestrict, &restrictData); err == nil {
					for _, r := range restrictData.Values {
						ch <- prometheus.MustNewConstMetric(
							c.branchRestrictionsTotal, prometheus.GaugeValue, 1, repo.ProjectKey, repo.RepoSlug, r.Branch, r.Type)
					}
				}
			}

			// Branch deleted (not available in API, log warning)
			logf("Branch deleted metric not available in Bitbucket Cloud API; skipping.")
		}
	} else {
		// Data Center logic (unchanged)
		repoCount, err := c.client.GetRepositoryCount()
		if err != nil {
			log.Printf("error collecting repo count: %v", err)
		} else {
			logf("repo count: %d", repoCount)
			ch <- prometheus.MustNewConstMetric(c.repoCount, prometheus.GaugeValue, float64(repoCount))
		}
		projectCount, err := c.client.GetProjectCount()
		if err != nil {
			log.Printf("error collecting project count: %v", err)
		} else {
			logf("project count: %d", projectCount)
			ch <- prometheus.MustNewConstMetric(c.projectCount, prometheus.GaugeValue, float64(projectCount))
		}
		// Add more Data Center/Server-safe metrics here as needed.
	}
	// Remove this line to avoid duplicate metric emission:
	// ch <- prometheus.MustNewConstMetric(c.exporterUp, prometheus.GaugeValue, exporterUpValue)

	// API Rate Limit (Bitbucket Cloud only)
	if c.client.Cloud {
		limitURL := "https://api.bitbucket.org/2.0/workspaces/" + c.client.Workspace + "/rate-limits/"
		req, _ := http.NewRequest("GET", limitURL, nil)
		req.SetBasicAuth(c.client.Username, c.client.Password)
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var limitData struct {
				Limits map[string]struct {
					Remaining int `json:"remaining"`
					Reset     int `json:"reset"`
				} `json:"limits"`
			}
			if err := json.Unmarshal(body, &limitData); err == nil {
				for name, l := range limitData.Limits {
					ch <- prometheus.MustNewConstMetric(c.apiRateLimitRemaining, prometheus.GaugeValue, float64(l.Remaining))
					ch <- prometheus.MustNewConstMetric(c.apiRateLimitResetSeconds, prometheus.GaugeValue, float64(l.Reset))
					if l.Remaining < 10 {
						log.Printf("[WARN] Bitbucket API rate limit for %s is low: %d remaining", name, l.Remaining)
					}
				}
			}
		}
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func parseRFC3339ToUnix(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func statusCodeSafe(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}
