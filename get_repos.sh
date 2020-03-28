#!/bin/sh

ORG=9Spokes
PROJECT=9Spokes
PAT="a5wiz5oymcg7dqjn7o246sbsn7t5lqh3gz5qu3qj4btrka26jrxa"

URL=https://dev.azure.com/${ORG}/${PROJECT}/_apis/git/repositories?api-version=4.1

echo URL is ${URL}

curl -sq ${URL} -u :${PAT} | jq '.value[].sshUrl' | perl -ne 's/"//g;~/\/([^\/]+)$/; print "- name: $1  dir: $1  url: $_\n"' > sentinel.yaml