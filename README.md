# Help the stars

## Needs

- Given a github token and smtp credentials
- Once a day, check issues of the repo stared by user
- Filter label `help-wanted` and open
- If new store them
- If some close, clear them
- send a mail / message to a matrix user with link and summary
- an interface allow to see all the issues and sort them

## How

- SQLITE
- watch for rate limiting, display percentage of requests if needed

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

## Run

```bash
go run *.go
```

## dev

```bash
alias sqlc="docker run --rm -v $(pwd):/src -w /src sqlc/sqlc"
```

### migrations

```bash
docker run -v $(pwd)/migrations:/migrations --network host migrate/migrate -path=/migrations -database "sqlite://help-stars.db" create -ext sql -dir /migrations -seq init_schema
```
