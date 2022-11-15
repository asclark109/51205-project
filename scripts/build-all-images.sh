#!/usr/bin/env bash
# Filename: build-all-images.sh

# stop script upon encountering the first error
# without closing terminal window
set -e


#####
# this script builds all images for the project from
# dockerfiles
#####

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

# need: location of docker file for the image you want to build
# need: a name for the image you are going to create

POSTGRES_IMG_DOCKERFILE="$SCRIPT_DIR/../postgres/"
POSTGRES_IMG_NAME="postgres-server"

RABBITMQ_IMG_DOCKERFILE="$SCRIPT_DIR/../rabbitmq/"
RABBITMQ_IMG_NAME="rabbitmq-server"

AUCTIONS_IMG_DOCKERFILE="$SCRIPT_DIR/../auctions-service/"
AUCTIONS_IMG_NAME="auctions-service"

echo RUNNING SCRIPT TO BUILD PROJECT IMAGES...
echo SCRIPT_DIR=$SCRIPT_DIR
echo

# we create each image from a docker file, using the docker command:
# 'docker build -t IMAGETAGNAME:VERSION PATHTODOCKERFILE'
# RABBIT MQ
echo "building RabbitMQ image ($RABBITMQ_IMG_NAME) from dockerfile..."
docker build -t "$RABBITMQ_IMG_NAME:latest" $RABBITMQ_IMG_DOCKERFILE

# Postgres SQL
echo "building Postgres image ($POSTGRES_IMG_NAME) from dockerfile..."
docker build -t "$POSTGRES_IMG_NAME:latest" $POSTGRES_IMG_DOCKERFILE

# RABBIT MQ
echo "building AuctionsService ($AUCTIONS_IMG_NAME) image from dockerfile..."
docker build -t "$AUCTIONS_IMG_NAME:latest" $AUCTIONS_IMG_DOCKERFILE

echo "done"