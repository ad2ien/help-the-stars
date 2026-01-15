# Help the stars

![GoLand](https://img.shields.io/badge/GoLand-0f0f0f?&logo=goland&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?logo=docker&logoColor=white)
[![Gitmoji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg)](https://gitmoji.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Build status](https://img.shields.io/github/actions/workflow/status/ad2ien/help-the-stars/build.yml?label=CI&logo=github)](https://github.com/ad2ien/help-the-stars/actions)

## What

Given a github token and matrix credentials

- Once a day, check configured labels like `help-wanted` issues of the repo stared by the user
- Update info in an sqlite database
- Send message to a matrix room with link and title
- Serves an interface listing help wanting repos

## Configuration

```bash
go run main.go --help
```

Matrix config is optional

## How

Github graphql request : <https://github.com/ad2ien/help-the-stars/blob/main/internal/stars.go#L53>

## Run

Create an env file

```.env
GITHUB_TOKEN=
DB_FILE="db/help-the-stars-dev.db"
LABELS='"help-wanted", "help wanted","junior friendly","good first issue"'

# optionally
MATRIX_HOMESERVER=
MATRIX_USERNAME=
MATRIX_PASSWORD=
MATRIX_ROOM_ID=
```

Start

```bash
source .env
go run *.go --debug \
--gh-token $GITHUB_TOKEN \
--db-file $DB_FILE \
--labels $LABELS \
--matrix-server $MATRIX_HOMESERVER \
--matrix-username $MATRIX_USERNAME \
--matrix-password $MATRIX_PASSWORD \
--matrix-room $MATRIX_ROOM_ID

```

Or with docker

```bash
docker compose build
docker compose up
```

Generate code after queries modifications

```bash
alias sqlc="docker run --rm -v $PWD:/src -w /src sqlc/sqlc"
sqlc generate
```

Create migration

```bash
docker run -v $(pwd)/migrations:/migrations --network host migrate/migrate -path=/migrations -database "sqlite://db/help-the-stars.db" create -ext sql -dir /migrations -seq MIGRATION_NAME
```

## Todo

- [ ] not the same number of issue in debug or script
- [ ] delete task db table row crashes
