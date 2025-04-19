#!/bin/bash

# abbreviated git tag
export TAG
export DATE
export SHORT_SHA
export FULL_TAG

TAG="$(git describe --tags --abbrev=0)"
DATE=$(date +%Y-%m-%d_%H-%M-%S)
FULL_TAG="${TAG}-${DATE}"

SHORT_SHA="$(git rev-parse --short HEAD)"
# containerize that shit
DOCKER_BUILDKIT=1 docker build -t esacteksab/tpt:"${FULL_TAG}" .

echo "${FULL_TAG}" > .current-tag
