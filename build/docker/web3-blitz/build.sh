#!/bin/bash
set -e
IMAGE_NAME=${1:-web3-blitz}
TAG=${2:-latest}
docker build -t "$IMAGE_NAME:$TAG" -f "$(dirname "$0")/Dockerfile" .
