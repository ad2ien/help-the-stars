package internal

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const LIMIT_CALLS = 5

type GhIssue struct {
	Title githubv4.String
	URL   githubv4.String
	Body  githubv4.String
}

type GhRepository struct {
	NameWithOwner  githubv4.String
	Description    githubv4.String
	StargazerCount githubv4.Int
	Issues         struct {
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

func GetStaredRepos(first int) ([]HelpWantedIssue, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GetSetting("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	variables := map[string]interface{}{
		"cursor": (*githubv4.String)(nil),
	}

	result := make([]HelpWantedIssue, 0)
	i := 0
	for {
		var query GhQuery
		fmt.Print("-")

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
