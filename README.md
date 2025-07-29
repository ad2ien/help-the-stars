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
query {
  viewer {
    starredRepositories(first: 50, after: $cursor) {
      nodes {
        nameWithOwner
        description
        issues(states: OPEN, labels: ["help-wanted"] first: 5){
          nodes{
            title
            url
            body
          }
        }
      }
    }
  }
}
```

## Run

```bash
go run *.go
```
