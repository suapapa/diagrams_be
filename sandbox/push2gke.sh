#!/bin/bash
#git tag -a $1 -m "add tag for $1"
#git push --tags

IMAGE_TAG=gcr.io/homin-dev/diagrams_sandbox
IMAGE_NAME=$IMAGE_TAG:$1 
IMAGE_NAME_LATEST=$IMAGE_TAG:latest

docker buildx build --platform linux/amd64 -f Dockerfile.gvisor -t $IMAGE_NAME .
docker push $IMAGE_NAME

docker tag $IMAGE_NAME $IMAGE_NAME_LATEST
docker push $IMAGE_NAME_LATEST
