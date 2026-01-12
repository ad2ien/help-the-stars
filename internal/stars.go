package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/oauth2"
)

const MAX_RETRY = 3
const BACKOFF_DELAY = 10 * time.Second
const MAX_ISSUES_PER_REPO = 5

type GhIssue struct {
	Title     string
	Url       string
	Body      string
	CreatedAt time.Time
}

type GhRepository struct {
	NameWithOwner  string
	Description    string
	StargazerCount int
	Issues         struct {
		Nodes []GhIssue
	}
}

type GhQuery struct {
	Data struct {
		Viewer struct {
			StarredRepositories struct {
				Nodes    []GhRepository
				PageInfo struct {
					EndCursor   string
					HasNextPage bool
				}
			}
		}
	}
}

const GRAPHQL_URL = "https://api.github.com/graphql"

const GRAPHQL_TEMPLATE = `
{
  viewer {
    starredRepositories(first: 50, after: "{{.RepoCursor}}") {
      nodes {
        nameWithOwner
        description
        stargazerCount
        issues(states: OPEN, labels: [{{.Labels}}], first: {{.MaxIssues}}) {
          nodes {
            title
            labels(first: 5) {
              edges{
                node {
                  name
                }
              }
            }
            url
            body
            createdAt
          }
          pageInfo {
            hasNextPage
          }
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}
`

// Parse the template once at package initialization
var tmpl = template.Must(template.New("graphql").Parse(GRAPHQL_TEMPLATE))

func buildQueryFromTemplate(repoCursor string) (string, error) {
	data := struct {
		RepoCursor string
		Labels     string
		MaxIssues  int
	}{
		RepoCursor: repoCursor,
		Labels:     GetSetting("LABELS"),
		MaxIssues:  MAX_ISSUES_PER_REPO,
	}

	var query bytes.Buffer
	err := tmpl.Execute(&query, data)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return query.String(), nil
}

func GetStaredRepos() ([]Repo, error) {
	result := make([]Repo, 0)
	cursor := ""
	for {
		log.Debug("Api call", "cursor", cursor)

		response, err := fetchQueryResults(cursor)
		if err != nil {
			return nil, err
		}

		result = append(result, mapGhQueryToHelpWantedIssue(response)...)

		if response.Data.Viewer.StarredRepositories.PageInfo.HasNextPage {
			cursor = response.Data.Viewer.StarredRepositories.PageInfo.EndCursor
		} else {
			break
		}
	}

	return result, nil

}

func fetchQueryResults(cursor string) (GhQuery, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GetSetting("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	// Create the request body
	query, err := buildQueryFromTemplate(cursor)
	if err != nil {
		log.Fatal("Error building query with template", "error", err)
		return GhQuery{}, err
	}

	requestBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		log.Fatalf("Error marshaling query: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", GRAPHQL_URL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := doWithRetry(httpClient, req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer closeBody(resp.Body)

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var queryResult GhQuery
	err = json.Unmarshal(body, &queryResult)
	if err != nil {
		log.Fatalf("Error unmarshaling response: %v", err)
	}

	return queryResult, nil
}

func doWithRetry(http *http.Client, req *http.Request) (*http.Response, error) {
	i := 1
	for {
		resp, err := http.Do(req)
		if err == nil {
			if resp.StatusCode >= 500 {
				log.Fatalf("Github server error: %d", resp.StatusCode)
			}
			return resp, nil
		}

		i++
		if i >= MAX_RETRY {
			return nil, err
		}
		time.Sleep(BACKOFF_DELAY * time.Duration(i))

	}
}

func closeBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.Warn("Error closing response body: %v", err)
	}
}
