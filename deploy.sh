#!/bin/bash

usage() {

cat <<EOS
Webhook Deployment Script
usage ./deploy.sh repoName refId commitId

Examples :
./deploy.sh refs/heads/master a09c01b8cefff3d7cb831c13c3551d9bc358a7f1

EOS
}

if [ $# -ne 3 ]; then
  usage
  exit
fi

echo "Deployment script inited"
echo "Name   : $1"
echo "Ref    : $2"
echo "Commit : $3"
