package internal

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const LIMIT_CALLS = 2

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
			Nodes    []GhRepository
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage githubv4.Boolean
			}
		} `graphql:"starredRepositories(first: 10, after: $cursor)"`
	}
}

func GetStaredRepos(first int) ([]HelpLookingRepo, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GetSetting("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
	}

	result := make([]HelpLookingRepo, 0)
	i := 0
	for {
		var query GhQuery
		fmt.Print(".")

		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			return nil, err
		}

		result = append(result, mapGhQueryToHelpWantedIssue(query)...)

		if query.Viewer.StarredRepositories.PageInfo.HasNextPage {
			variables["cursor"] = *githubv4.NewString(query.Viewer.StarredRepositories.PageInfo.EndCursor)
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

func mapGhQueryToHelpWantedIssue(query GhQuery) []HelpLookingRepo {
	var helpLookingRepos []HelpLookingRepo

	for _, repo := range query.Viewer.StarredRepositories.Nodes {
		if repo.Issues.Nodes == nil ||
			len(repo.Issues.Nodes) == 0 {
			continue
		}
		helpLookingRepo := HelpLookingRepo{
			RepoOwner:       string(repo.NameWithOwner),
			RepoDescription: string(repo.Description),
		}
		for _, issue := range repo.Issues.Nodes {
			helpWantedIssue := HelpWantedIssue{
				Title:            string(issue.Title),
				IssueDescription: string(issue.Body),
				URL:              string(issue.URL),
			}
			helpLookingRepo.Issues = append(helpLookingRepo.Issues, helpWantedIssue)
		}
		helpLookingRepos = append(helpLookingRepos, helpLookingRepo)
	}

	return helpLookingRepos
}
