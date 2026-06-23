# Help the stars

![Golang](https://img.shields.io/badge/language-Golang-00ADD8?&logo=go&logoColor=white)
[![Gitmoji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg?logo=git)](https://gitmoji.dev)
![Docker](https://img.shields.io/badge/build-docker-%230db7ed.svg?logo=docker&logoColor=white)
[![License](https://img.shields.io/badge/📝%20License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
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

There's a [justfile](https://just.systems/man/en/)

all the commands available in the [./justfile](./justfile)

```bash
# List commands
just -l
# run
just run
# ...
```

Or with docker

```bash
docker compose build
docker compose up
```

Generate code after queries modifications

```bash
just sqlcgen
```

Create migration

```bash
just new-migration MIGRATION_NAME
```
