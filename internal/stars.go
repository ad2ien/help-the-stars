package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

const GRAPHQL_URL = "https://api.github.com/graphql"

const GRAPHQL_TEMPLATE = `
{
  viewer {
    starredRepositories(first: 50, after: "{{.RepoCursor}}") {
      nodes {
        nameWithOwner
        description
        stargazerCount
        languages(first:10){
        	edges{
            size
            node{
              name
            }
          }
        }
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
  rateLimit {
     limit
     remaining
     used
     resetAt
   }
}
`

var ErrUnexpectedStatusCode = errors.New("unexpected status code")

type GhStarsService struct {
	settingsService *SettingsService
	tmpl            *template.Template
}

type GhIssue struct {
	Title     string    `json:"title"`
	Url       string    `json:"url"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type GhRepository struct {
	NameWithOwner  string `json:"nameWithOwner"`
	Description    string `json:"description"`
	StargazerCount int    `json:"stargazerCount"`
	Issues         struct {
		Nodes []GhIssue `json:"nodes"`
	} `json:"issues"`
	Languages struct {
		Edges []GhLanguageEdge `json:"edges"`
	} `json:"languages"`
}

type GhLanguageEdge struct {
	Size int            `json:"size"`
	Node GhLanguageNode `json:"node"`
}

type GhLanguageNode struct {
	Name string `json:"name"`
}

type RateLimit struct {
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	Used      int    `json:"used"`
	ResetAt   string `json:"resetAt"`
}

type GhQuery struct {
	Data struct {
		Viewer struct {
			StarredRepositories struct {
				Nodes    []GhRepository `json:"nodes"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"starredRepositories"`
		} `json:"viewer"`
		RateLimit RateLimit `json:"rateLimit"`
	} `json:"data"`
}

func NewGithubStarService(settingsService *SettingsService) *GhStarsService {
	return &GhStarsService{
		settingsService: settingsService,
		tmpl:            template.Must(template.New("graphql").Parse(GRAPHQL_TEMPLATE)),
	}
}
func (ghs *GhStarsService) GetStaredRepos(ctx context.Context) ([]Repo, error) {
	result := make([]Repo, 0)

	cursor := ""
	for {
		log.Debug("Api call", "cursor", cursor)

		response, err := ghs.fetchQueryResults(ctx, cursor)
		if err != nil {
			return nil, err
		}

		log.Debug("Remaining rate limit", "remaining", response.Data.RateLimit.Remaining)

		result = append(result, mapGhQueryToHelpWantedIssue(response)...)

		if response.Data.Viewer.StarredRepositories.PageInfo.HasNextPage {
			cursor = response.Data.Viewer.StarredRepositories.PageInfo.EndCursor
		} else {
			break
		}
	}

	return result, nil
}

func (ghs *GhStarsService) buildQueryFromTemplate(repoCursor string) (string, error) {
	data := struct {
		RepoCursor string
		Labels     string
		MaxIssues  int
	}{
		RepoCursor: repoCursor,
		Labels:     ghs.settingsService.GetSettings().ConfiguredLabels,
		MaxIssues:  MAX_ISSUES_PER_REPO,
	}

	var query bytes.Buffer

	if err := ghs.tmpl.Execute(&query, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return query.String(), nil
}

func (ghs *GhStarsService) fetchQueryResults(ctx context.Context, cursor string) (GhQuery, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghs.settingsService.GetSettings().GhToken},
	)
	httpClient := oauth2.NewClient(ctx, src)

	// Create the request body
	query, err := ghs.buildQueryFromTemplate(cursor)
	if err != nil {
		log.Error("Error building query with template", "error", err)

		return GhQuery{}, err
	}

	requestBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		log.Error("Error marshaling query: %v", err)

		return GhQuery{}, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GRAPHQL_URL, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Error creating request: %v", err)

		return GhQuery{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	ghResponse, err := doWithRetry(httpClient, req)
	if err != nil {
		log.Error("Error sending request: %v", err)

		return GhQuery{}, err
	}

	return responseToResult(ghResponse)
}

func doWithRetry(httpclient *http.Client, req *http.Request) (*http.Response, error) {
	i := 1

	for {
		resp, err := httpclient.Do(req) //nolint:gosec // URL is built internally
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			log.Warn("Rate limit exceeded, wait until reset...")

			queryRes, err := responseToResult(resp)
			if err != nil {
				return nil, err
			}

			resetTime, err := ParseGhDate(queryRes.Data.RateLimit.ResetAt)
			if err != nil {
				return nil, err
			}
			// Wait until the rate limit is reset
			time.Sleep(time.Until(resetTime))
		}

		if err == nil {
			log.Warn("Github server error", "status", resp.StatusCode)
		}

		i++
		if i >= MAX_RETRY {
			return nil, fmt.Errorf("failed to execute request after %d retries: %w", MAX_RETRY, err)
		}

		time.Sleep(BACKOFF_DELAY * time.Duration(i))
	}
}

func responseToResult(resp *http.Response) (GhQuery, error) {
	queryResult := GhQuery{}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading rate limited body : %v", err)

		return queryResult, fmt.Errorf("failed to read response: %w", err)
	}
	defer closeBody(resp.Body)

	if err = json.Unmarshal(body, &queryResult); err != nil {
		log.Error("Error unmarshaling response: %v", err)

		return queryResult, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return queryResult, nil
}

func closeBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		log.Warn("Error closing response body: %v", err)
	}
}

func ParseGhDate(date string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %w", err)
	}

	return t, nil
}
