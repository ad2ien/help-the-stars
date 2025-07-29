package main

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const LIMIT_CALLS = 3

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
	} `graphql:"issues(states: OPEN, labels: [\"help-wanted\"], first: 5, after: $cursor)"`
}

type GhQuery struct {
	Viewer struct {
		StarredRepositories struct {
			Nodes    []GhRepository
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage githubv4.Boolean
			}
		} `graphql:"starredRepositories(first: 50)"`
	}
}

func GetStaredRepos(first int) ([]HelpWantedIssue, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GetSetting("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	cursor := (*githubv4.String)(nil)
	variables := map[string]interface{}{
		"cursor": cursor,
	}

	result := make([]HelpWantedIssue, 0)
	i := 0
	for {
		var query GhQuery

		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, err
		}

		result = append(result, mapGhQueryToHelpWantedIssue(query)...)

		if query.Viewer.StarredRepositories.PageInfo.HasNextPage {
			cursor = githubv4.NewString(query.Viewer.StarredRepositories.PageInfo.EndCursor)
		} else {
			break
		}

		if i >= LIMIT_CALLS {
			break
		}
		i++
	}

	return result, nil
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
