#!/bin/bash
# build for pi3
env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w"
