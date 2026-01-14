#!/bin/bash
set -e

source .env
go run *.go --debug \
--gh-token $GITHUB_TOKEN \
--db-file $DB_FILE \
--labels $LABELS \
--matrix-server $MATRIX_SERVER \
--matrix-username $MATRIX_USERNAME \
--matrix-password $MATRIX_PASSWORD \
--matrix-room $MATRIX_ROOM_ID
