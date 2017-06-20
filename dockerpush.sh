#!/usr/bin/env bash

CGO_ENABLED=0 go build -ldflags '-extldflags "-static" -w -s'

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
docker build -t miracl/casper:$TRAVIS_TAG .
docker push miracl/casper:$TRAVIS_TAG
docker push miracl/casper:latest
