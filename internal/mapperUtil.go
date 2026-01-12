package internal

import (
	"database/sql"
	"help-the-stars/internal/persistence"

	"github.com/charmbracelet/log"
)

// map GhQuery to HelpWantedIssue only if there's an issue
func mapGhQueryToHelpWantedIssue(query GhQuery) []Repo {
	var repos []Repo

	for _, repo := range query.Data.Viewer.StarredRepositories.Nodes {
		if len(repo.Issues.Nodes) == 0 {
			continue
		}
		log.Debug("Processing", "repo", repo.NameWithOwner)
		r := Repo{
			RepoOwner:       repo.NameWithOwner,
			RepoDescription: repo.Description,
			StargazersCount: repo.StargazerCount,
		}
		for _, issue := range repo.Issues.Nodes {
			helpWantedIssue := HelpWantedIssue{
				Title:            string(issue.Title),
				IssueDescription: string(issue.Body),
				Url:              string(issue.Url),
				CreationDate:     issue.CreatedAt,
			}
			r.Issues = append(r.Issues, helpWantedIssue)
		}
		repos = append(repos, r)
	}

	return repos
}

func mapModelToDbParameter(issue HelpWantedIssue, repo Repo) persistence.CreateIssueParams {
	return persistence.CreateIssueParams{
		Url:             issue.Url,
		Title:           issue.Title,
		Description:     issue.IssueDescription,
		RepoWithOwner:   repo.RepoOwner,
		RepoDescription: repo.RepoDescription,
		CreationDate:    issue.CreationDate,
		StargazersCount: sql.NullInt64{Int64: int64(repo.StargazersCount), Valid: true},
	}
}

func mapDbResultToViewModel(issues []persistence.Issue, taskData persistence.TaskDatum) ThankStarsData {
	return ThankStarsData{
		Repos:             mapDbIssuesToViewRepos(issues),
		LastUpdate:        taskData.LastRun.Time,
		CurrentlyUpdating: taskData.InProgress.Valid && taskData.InProgress.Bool,
	}
}

func mapDbIssuesToViewRepos(issues []persistence.Issue) []Repo {
	repoMap := make(map[string]Repo)
	for _, issue := range issues {
		if _, ok := repoMap[issue.RepoWithOwner]; !ok {
			repoMap[issue.RepoWithOwner] = Repo{
				RepoOwner:       issue.RepoWithOwner,
				RepoDescription: issue.RepoDescription,
				StargazersCount: int(issue.StargazersCount.Int64),
				Issues: []HelpWantedIssue{
					mapDbIssueToViewIssue(issue),
				},
			}
			continue
		}
		repo := repoMap[issue.RepoWithOwner]
		repo.Issues = append(repo.Issues, mapDbIssueToViewIssue(issue))
		repoMap[issue.RepoWithOwner] = repo
	}
	result := make([]Repo, 0, len(repoMap))
	for _, repo := range repoMap {
		result = append(result, repo)
	}
	return result
}

func mapDbIssueToViewIssue(issue persistence.Issue) HelpWantedIssue {
	return HelpWantedIssue{
		Title:            issue.Title,
		IssueDescription: issue.Description,
		Url:              issue.Url,
		CreationDate:     issue.CreationDate,
	}
}

func mapDbIssuesToViewIssues(issues []persistence.Issue) []HelpWantedIssue {
	mappedIssues := make([]HelpWantedIssue, len(issues))
	for i, issue := range issues {
		mappedIssues[i] = mapDbIssueToViewIssue(issue)
	}
	return mappedIssues
}

func flattenIssues(repos []Repo) []HelpWantedIssue {
	var issues []HelpWantedIssue
	for _, repo := range repos {
		issues = append(issues, repo.Issues...)
	}
	return issues
}
