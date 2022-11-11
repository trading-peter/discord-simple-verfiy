#!/bin/bash
mkdir -p _build
cd _build
mkdir -p docker_out
rm -rf sources
git clone https://github.com/trading-peter/discord-simple-verfiy.git sources
cd sources
git fetch --tags
ver=$(git describe --tags `git rev-list --tags --max-count=1`)
git checkout $ver

CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bot .

# A second run is needed to build the final image.
cd ..
docker build -f sources/Dockerfile --no-cache -t simple-verify-bot:${ver} .
docker save simple_verify_bot:${ver} > docker_out/simple-verify-bot_${ver}.tar
rm -rf sources bot