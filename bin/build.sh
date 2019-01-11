#!/usr/bin/env bash

export GOOS=linux
go build -o navexplorer-api-linux-amd64
export GOOS=darwin