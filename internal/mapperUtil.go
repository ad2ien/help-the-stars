package internal

import (
	"database/sql"
	"help-the-stars/internal/persistence"
)

func mapGhQueryToHelpWantedIssue(query GhQuery) []HelpWantedIssue {
	var helpLookingIssues []HelpWantedIssue

	for _, repo := range query.Viewer.StarredRepositories.Nodes {
		if len(repo.Issues.Nodes) == 0 {
			continue
		}
		for _, issue := range repo.Issues.Nodes {
			helpWantedIssue := HelpWantedIssue{
				Title:            string(issue.Title),
				IssueDescription: string(issue.Body),
				Url:              string(issue.URL),
				CreationDate:     issue.CreatedAt.Time,
				RepoOwner:        string(repo.NameWithOwner),
				RepoDescription:  string(repo.Description),
				StargazersCount:  int(repo.StargazerCount),
			}
			helpLookingIssues = append(helpLookingIssues, helpWantedIssue)
		}
	}

	return helpLookingIssues
}

func mapModelToDbParameter(issue HelpWantedIssue) persistence.CreateIssueParams {
	return persistence.CreateIssueParams{
		Url:             issue.Url,
		Title:           issue.Title,
		Description:     issue.IssueDescription,
		RepoWithOwner:   issue.RepoOwner,
		CreationDate:    issue.CreationDate,
		StargazersCount: sql.NullInt64{Int64: int64(issue.StargazersCount), Valid: true},
	}
}

func mapDbResultToViewData(issues []persistence.Issue, taskData persistence.TaskDatum) ThankStarsData {
	return ThankStarsData{
		Issues:            mapIssuesDbToModels(issues),
		LastUpdate:        taskData.LastRun.Time,
		CurrentlyUpdating: taskData.InProgress.Valid && taskData.InProgress.Bool,
	}
}

func mapIssuesDbToModels(issues []persistence.Issue) []HelpWantedIssue {
	mappedIssues := make([]HelpWantedIssue, len(issues))
	for i, issue := range issues {
		mappedIssues[i] = mapIssueDbToModel(issue)
	}
	return mappedIssues
}

func mapIssueDbToModel(issue persistence.Issue) HelpWantedIssue {
	stargazerCount := 0
	if issue.StargazersCount.Valid {
		stargazerCount = int(issue.StargazersCount.Int64)
	}
	return HelpWantedIssue{
		Url:              issue.Url,
		Title:            issue.Title,
		IssueDescription: issue.Description,
		CreationDate:     issue.CreationDate,
		RepoOwner:        issue.RepoWithOwner,
		RepoDescription:  issue.RepoDescription,
		StargazersCount:  stargazerCount,
	}

}
