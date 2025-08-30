# Help the stars

![GoLand](https://img.shields.io/badge/GoLand-0f0f0f?&logo=goland&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?logo=docker&logoColor=white)
[![Gitmoji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg)](https://gitmoji.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build status](https://img.shields.io/github/actions/workflow/status/ad2ien/help-the-stars/build.yml?label=CI&logo=github)](https://github.com/ad2ien/help-the-stars/actions)

## What

Given a github token and matrix credentials

- Once a day, check `help-wanted` issues of the repo stared by the user
- Update info in an sqlite database
- Send message to a matrix room with link and title
- Serves an interface listing open help wanting issues

## How

Github graphql request

```graphql
{
  viewer {
    starredRepositories(first: 50, after: "$cursor") {
      nodes {
        nameWithOwner
        description
        stargazerCount
        issues(states: OPEN, labels: ["help-wanted"], first: 5) {
          nodes {
            title
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
```

## dev

Run

```bash
go run *.go
```

Or with docker

```bash
docker compose up
```

Generate code after queries modifications

```bash
alias sqlc="docker run --rm -v $(pwd):/src -w /src sqlc/sqlc"
sqlc generate
```

Create migration

```bash
docker run -v $(pwd)/migrations:/migrations --network host migrate/migrate -path=/migrations -database "sqlite://db/help-the-stars.db" create -ext sql -dir /migrations -seq MIGRATION_NAME
```

## TODO

- [ ] lint go/docker
- [ ] some interactivity with htmx
