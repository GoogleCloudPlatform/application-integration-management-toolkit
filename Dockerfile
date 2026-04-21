# Copyright 2023 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.26.1 AS builder

ARG TAG
ARG COMMIT

ADD ./internal /go/src/integrationcli/internal
ADD ./cmd /go/src/integrationcli/cmd

COPY go.mod go.sum /go/src/integrationcli/
WORKDIR /go/src/integrationcli

ENV GO111MODULE=on
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=true -a -gcflags='all="-l"' -ldflags='-s -w -extldflags "-static" -X main.version='${TAG}' -X main.commit='${COMMIT}' -X main.date='$(date +%FT%H:%I:%M+%Z) -o /go/bin/integrationcli /go/src/integrationcli/cmd/integrationcli/integrationcli.go
RUN /go/bin/integrationcli prefs set --nocheck=true
RUN GOBIN=/tmp/ go install github.com/google/go-licenses@v1.6.0
RUN /tmp/go-licenses report ./... --template /go/src/apigeecli/licenses.tpl --ignore internal > /tmp/third-party-licenses.txt 2> /dev/null || echo "Ignore warnings"


FROM us-docker.pkg.dev/appintegration-toolkit/internal/jq:latest@sha256:d3a1c8a88f9223eab96bda760efab08290d274249581d2db6db010cbe20c232b AS jq

# use debug because it includes busybox
FROM gcr.io/distroless/static-debian12:debug-nonroot@sha256:afead1275cad5ec9662cdc09ce7fe5961a41467555fc30cd46a60247bf8bbdfd
LABEL org.opencontainers.image.url='https://github.com/GoogleCloudPlatform/application-integration-management-toolkit' \
    org.opencontainers.image.documentation='https://github.com/GoogleCloudPlatform/application-integration-management-toolkit' \
    org.opencontainers.image.source='https://github.com/GoogleCloudPlatform/application-integration-management-toolkit' \
    org.opencontainers.image.vendor='Google LLC' \
    org.opencontainers.image.licenses='Apache-2.0' \
    org.opencontainers.image.description='This is a tool to interact with Application Integration APIs'
COPY --from=builder /go/bin/integrationcli /usr/local/bin/integrationcli
COPY --from=builder --chown=nonroot:nonroot /root/.integrationcli/config.json /home/nonroot/.integrationcli/config.json
COPY --from=jq /jq /usr/local/bin/jq
COPY --chown=nonroot:nonroot LICENSE.txt /
COPY --from=builder --chown=nonroot:nonroot /tmp/third-party-licenses.txt /

ENTRYPOINT [ "integrationcli" ]
