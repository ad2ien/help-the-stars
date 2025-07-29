package main

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type HelpWantedIssue struct {
	Title            string
	IssueDescription string
	URL              string
	RepoOwner        string
	RepoDescription  string
}

type GhIssue struct {
	Title githubv4.String
	URL   githubv4.String
	Body  githubv4.String
}

type GhRepository struct {
	NameWithOwner githubv4.String
	Description   githubv4.String
	Issues        struct {
		Nodes []GhIssue
	} `graphql:"issues(states: OPEN, labels: [\"help-wanted\"], first: 5)"`
}

type GhQuery struct {
	Viewer struct {
		StarredRepositories struct {
			Nodes []GhRepository
		} `graphql:"starredRepositories(first: 50)"`
	}
}

func GetStaredRepos(first int) ([]HelpWantedIssue, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GetSetting("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	var query GhQuery
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		return nil, err
	}

	return mapGhQueryToHelpWantedIssue(query), nil
}

func mapGhQueryToHelpWantedIssue(query GhQuery) []HelpWantedIssue {
	var helpWantedIssues []HelpWantedIssue

	for _, repo := range query.Viewer.StarredRepositories.Nodes {
		if repo.Issues.Nodes == nil ||
			len(repo.Issues.Nodes) == 0 {
			continue
		}
		for _, issue := range repo.Issues.Nodes {
			helpWantedIssue := HelpWantedIssue{
				Title:            string(issue.Title),
				IssueDescription: string(issue.Body),
				URL:              string(issue.URL),
				RepoOwner:        string(repo.NameWithOwner),
				RepoDescription:  string(repo.Description),
			}
			helpWantedIssues = append(helpWantedIssues, helpWantedIssue)
		}
	}

	return helpWantedIssues
}
