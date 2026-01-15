package internal

import (
	"database/sql"
	"help-the-stars/internal/persistence"
	"time"

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

		lang := ""
		if len(repo.Languages.Nodes) > 0 {
			lang = repo.Languages.Nodes[0].Name
		}

		r := Repo{
			RepoOwner:       repo.NameWithOwner,
			RepoDescription: repo.Description,
			StargazersCount: repo.StargazerCount,
			Language:        lang,
		}
		lastCreationTime := time.Time{}
		for _, issue := range repo.Issues.Nodes {
			helpWantedIssue := HelpWantedIssue{
				Title:            string(issue.Title),
				IssueDescription: string(issue.Body),
				Url:              string(issue.Url),
				CreationDate:     issue.CreatedAt,
				RepoWithOwner:        repo.NameWithOwner,
			}
			r.Issues = append(r.Issues, helpWantedIssue)

			if helpWantedIssue.CreationDate.After(lastCreationTime) {
				lastCreationTime = helpWantedIssue.CreationDate
			}
		}
		r.LastIssueCreationTime = lastCreationTime
		repos = append(repos, r)
	}

	return repos
}

func mapModelToIssueDbParameter(issue HelpWantedIssue) persistence.CreateIssueParams {
	return persistence.CreateIssueParams{
		Url:           issue.Url,
		Title:         issue.Title,
		Description:   issue.IssueDescription,
		RepoWithOwner: issue.RepoWithOwner,
		CreationDate:  issue.CreationDate,
	}
}

func mapModelToRepoDbParameter(repo Repo) persistence.CreateRepoParams {
	return persistence.CreateRepoParams{
		RepoWithOwner:   repo.RepoOwner,
		Description:     sql.NullString{String: repo.RepoDescription, Valid: true},
		StargazersCount: sql.NullInt64{Int64: int64(repo.StargazersCount), Valid: true},
		Language:        sql.NullString{String: repo.Language, Valid: true},
	}
}

func mapDbResultToViewModel(
	issues []persistence.Issue,
	repos []persistence.Repo,
	taskData persistence.TaskDatum) ThankStarsData {
	return ThankStarsData{
		Repos:             mapDbIssuesToViewRepos(issues, repos),
		LastUpdate:        taskData.LastRun.Time,
		CurrentlyUpdating: taskData.InProgress.Valid && taskData.InProgress.Bool,
	}
}

func mapDbIssuesToViewRepos(issues []persistence.Issue, repos []persistence.Repo) []Repo {

	result := make([]Repo, len(repos))
	for i, repo := range repos {

		filteredIssues, lastIssueDate := findIssuesAndLastIssueDateByRepoOwner(repo.RepoWithOwner, issues)
		result[i] = Repo{
			RepoOwner:       repo.RepoWithOwner,
			RepoDescription: repo.Description.String,
			StargazersCount: int(repo.StargazersCount.Int64),
			Language:        repo.Language.String,
			Issues:          filteredIssues,
			LastIssueCreationTime:   lastIssueDate,
		}
	}

	return result
}

func findIssuesAndLastIssueDateByRepoOwner(repoOwner string, issues []persistence.Issue) (filteredIssues []HelpWantedIssue,
	lastIssueDate time.Time) {
	for _, issue := range issues {
		if issue.RepoWithOwner == repoOwner {
			filteredIssues = append(filteredIssues, mapDbIssueToViewIssue(issue))
			if issue.CreationDate.After(lastIssueDate) {
				lastIssueDate = issue.CreationDate
			}
		}
	}
	return filteredIssues, lastIssueDate
}

func mapDbIssueToViewIssue(issue persistence.Issue) HelpWantedIssue {
	return HelpWantedIssue{
		Title:            issue.Title,
		IssueDescription: issue.Description,
		Url:              issue.Url,
		CreationDate:     issue.CreationDate,
		RepoWithOwner:        issue.RepoWithOwner,
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

func mapDbRepoToViewRepo(repo persistence.Repo) Repo {
	return Repo{
		RepoOwner:       repo.RepoWithOwner,
		RepoDescription: repo.Description.String,
		StargazersCount: int(repo.StargazersCount.Int64),
		Language:        repo.Language.String,
		Issues:          nil,
		LastIssueCreationTime: time.Time{},
	}
}

func mapDbReposToViewRepos(repos []persistence.Repo) []Repo {
	mappedRepos := make([]Repo, len(repos))
	for i, repo := range repos {
		mappedRepos[i] = mapDbRepoToViewRepo(repo)
	}
	return mappedRepos
}
