FROM golang:1.19 as builder

ADD ./apiclient /go/src/integrationcli/apiclient
ADD ./client /go/src/integrationcli/client
ADD ./cmd /go/src/integrationcli/cmd
ADD ./cloudkms /go/src/integrationcli/cloudkms

COPY main.go /go/src/integrationcli/main.go
COPY go.mod go.sum /go/src/integrationcli/
WORKDIR /go/src/integrationcli

ENV GO111MODULE=on
RUN go mod tidy
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -a -ldflags='-s -w -extldflags "-static"' -o /go/bin/integrationcli /go/src/integrationcli/main.go

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/integrationcli /
CMD ["/integrationcli"]
