#!/bin/bash
gcloud builds submit --config=cloudbuild.yaml --project=${PROJECT_ID} --region=${LOCATION}


