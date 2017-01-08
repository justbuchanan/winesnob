#!/usr/bin/env bash

set -e

function remote_run() {
    ssh justbuchanan.com "cd cellar; $1"
}

echo "=> Prepare remote server for push"
remote_run "git checkout master"

branch="oauth2"

echo "=> Push latest code to server"
git push prod $branch

echo "=> Rebuilding and re-running docker image"
remote_run "git checkout $branch; docker-compose down && docker-compose up --build -d"
