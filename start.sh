#!/bin/bash
set -e

source .env
go run *.go --debug \
	--gh-token $GH_TOKEN \
	--labels $LABELS \
	--db-file $DB_FILE \
	--matrix-server $MATRIX_SERVER \
	--matrix-username $MATRIX_USERNAME \
	--matrix-password $MATRIX_PASSWORD \
	--matrix-room $MATRIX_ROOM
