#!/bin/sh
gcloud deploy releases create test-release-$((RANDOM % 900 + 100)) --project=${PROJECT_ID} --region=${LOCATION} --delivery-pipeline=appint-sample-pipeline
