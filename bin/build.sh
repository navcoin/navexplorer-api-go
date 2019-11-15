#!/usr/bin/env bash
set -e

dingo -src=./internal/config/di -dest=./generated

docker build . -t navexplorer/api:dev
docker push navexplorer/api:dev
