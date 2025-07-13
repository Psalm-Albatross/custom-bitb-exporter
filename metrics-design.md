# 📊 Bitbucket Exporter Metrics Design

This document defines the complete set of Prometheus metrics for the Bitbucket Exporter (supporting Bitbucket Cloud and Data Center/Server).

---

## 🔹 1. Project & Repository Metrics

```
# HELP bitbucket_project_count Total number of projects
# TYPE bitbucket_project_count gauge

# HELP bitbucket_project_repos Number of repositories per project
# TYPE bitbucket_project_repos gauge
# LABELS: project_key, project_name

# HELP bitbucket_repository_count Total number of repositories
# TYPE bitbucket_repository_count gauge

# HELP bitbucket_repo_size_bytes Size of each repository in bytes
# TYPE bitbucket_repo_size_bytes gauge
# LABELS: project_key, project_name, repo_slug, repo_name

# HELP bitbucket_repo_last_commit_timestamp Unix timestamp of last commit in repo
# TYPE bitbucket_repo_last_commit_timestamp gauge
# LABELS: project_key, project_name, repo_slug, repo_name
```

## 🔹 2. Pull Request (PR) Metrics

```
# HELP bitbucket_open_pull_requests Total number of open pull requests
# TYPE bitbucket_open_pull_requests gauge

# HELP bitbucket_repo_open_prs Number of open PRs per repository
# TYPE bitbucket_repo_open_prs gauge
# LABELS: project_key, project_name, repo_slug, repo_name

# HELP bitbucket_pull_requests_merged_total Cumulative number of merged pull requests
# TYPE bitbucket_pull_requests_merged_total counter

# HELP bitbucket_pull_request_age_seconds Age of each PR in seconds
# TYPE bitbucket_pull_request_age_seconds gauge
# LABELS: project_key, repo_slug, pr_id, state

# HELP bitbucket_pull_request_reviewers_total Number of reviewers per PR
# TYPE bitbucket_pull_request_reviewers_total gauge
# LABELS: project_key, repo_slug, pr_id
```

## 🔹 3. Commit & Author Metrics

```
# HELP bitbucket_repo_commits Number of commits per repository
# TYPE bitbucket_repo_commits gauge
# LABELS: project_key, project_name, repo_slug, repo_name

# HELP bitbucket_user_commits Number of commits per user per repo
# TYPE bitbucket_user_commits gauge
# LABELS: project_key, project_name, repo_slug, repo_name, user

# HELP bitbucket_commit_age_seconds Age of commits in seconds (latest only)
# TYPE bitbucket_commit_age_seconds gauge
# LABELS: repo_slug
```

## 🔹 4. User & Access Metrics

```
# HELP bitbucket_user_count Total number of users
# TYPE bitbucket_user_count gauge

# HELP bitbucket_user_access_repos_total Number of repositories accessed per user
# TYPE bitbucket_user_access_repos_total gauge
# LABELS: user, permission_level

# HELP bitbucket_team_members_total Number of members per team (Bitbucket Cloud)
# TYPE bitbucket_team_members_total gauge
# LABELS: team_name
```

## 🔹 5. Pipeline / Build Metrics (Cloud Only)

```
# HELP bitbucket_pipeline_runs_total Number of pipeline runs by result
# TYPE bitbucket_pipeline_runs_total counter
# LABELS: repo_slug, result

# HELP bitbucket_pipeline_duration_seconds Duration of each pipeline run
# TYPE bitbucket_pipeline_duration_seconds histogram
# LABELS: repo_slug

# HELP bitbucket_pipeline_active_total Currently running pipelines
# TYPE bitbucket_pipeline_active_total gauge
# LABELS: repo_slug
```

## 🔹 6. Branch Policy / Protection Metrics

```
# HELP bitbucket_branch_restrictions_total Number of branch restrictions per branch
# TYPE bitbucket_branch_restrictions_total gauge
# LABELS: project_key, repo_slug, branch_name, restriction_type

# HELP bitbucket_branch_default_policy_enforced Whether default branch has policy enforced (0/1)
# TYPE bitbucket_branch_default_policy_enforced gauge
# LABELS: repo_slug, branch_name
```

## 🔹 7. Webhook Metrics

```
# HELP bitbucket_webhooks_total Total number of webhooks configured
# TYPE bitbucket_webhooks_total gauge
# LABELS: repo_slug, status

# HELP bitbucket_webhook_failures_total Number of webhook failures
# TYPE bitbucket_webhook_failures_total counter
# LABELS: repo_slug, event_type, endpoint

# HELP bitbucket_webhook_delivery_duration_seconds Duration of webhook delivery
# TYPE bitbucket_webhook_delivery_duration_seconds histogram
# LABELS: repo_slug, event_type
```

## 🔹 8. API Usage & Exporter Health

```
# HELP bitbucket_api_rate_limit_remaining Remaining API rate limit (Cloud)
# TYPE bitbucket_api_rate_limit_remaining gauge

# HELP bitbucket_api_rate_limit_reset_seconds Time in seconds until rate limit reset
# TYPE bitbucket_api_rate_limit_reset_seconds gauge

# HELP bitbucket_exporter_up Whether the Bitbucket exporter is running successfully
# TYPE bitbucket_exporter_up gauge

# HELP bitbucket_exporter_errors_total Total number of errors in exporter
# TYPE bitbucket_exporter_errors_total counter
# LABELS: error_type, component
```

## 🔹 9. Tags / Releases / Issues

```
# HELP bitbucket_tags_total Number of Git tags in repository
# TYPE bitbucket_tags_total gauge
# LABELS: repo_slug

# HELP bitbucket_issues_total Number of open issues (Cloud only, if enabled)
# TYPE bitbucket_issues_total gauge
# LABELS: repo_slug, status

# HELP bitbucket_releases_total Number of releases per repository (if supported)
# TYPE bitbucket_releases_total gauge
# LABELS: repo_slug
```

---

All metrics follow Prometheus best practices:

* `snake_case` naming
* Use of `gauge`, `counter`, and `histogram` types appropriately
* Low cardinality in labels
* Grouped under the `bitbucket_` prefix

This design enables scalable, granular observability into Bitbucket health, activity, and usage.
