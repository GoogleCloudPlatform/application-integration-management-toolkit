#!/bin/sh
gcloud deploy apply --file=clouddeploy.yaml --region=${LOCATION} --project=${PROJECT_ID}
