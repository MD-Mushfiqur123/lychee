#!/bin/sh
# Quick Docker test — builds image, checks it starts
docker build -t lychee:test .
docker run -d --name lychee-test -p 11435:11434 lychee:test
sleep 3
curl http://localhost:11435/api/version
docker stop lychee-test && docker rm lychee-test
