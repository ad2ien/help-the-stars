package internal

import (
	"database/sql"
	"help-the-stars/internal/persistence"
)

func mapGhQueryToHelpWantedIssue(query GhQuery) []HelpWantedIssue {
	var helpLookingIssues []HelpWantedIssue

	for _, repo := range query.Viewer.StarredRepositories.Nodes {
		if repo.Issues.Nodes == nil ||
			len(repo.Issues.Nodes) == 0 {
			continue
		}
		for _, issue := range repo.Issues.Nodes {
			helpWantedIssue := HelpWantedIssue{
				Title:            string(issue.Title),
				IssueDescription: string(issue.Body),
				Url:              string(issue.URL),
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
