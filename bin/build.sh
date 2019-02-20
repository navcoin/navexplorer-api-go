#!/usr/bin/env bash

export GOOS=linux
export GOARG=amd64

go build -o navexplorerApi

docker build -t dantudor/navexplorer:navexplorerapi-0.0.1 .
docker push dantudor/navexplorer:navexplorerapi-0.0.1