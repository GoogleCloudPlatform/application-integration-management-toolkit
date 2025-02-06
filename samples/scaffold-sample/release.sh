#!/bin/sh
RELEASE_NAME=test-release
gcloud deploy releases create $RELEASE_NAME--$(date +'%Y%m%d%H%M%S') --project=${PROJECT_ID} --region=${LOCATION} --to-target=dev --delivery-pipeline=appint-sample-pipeline
