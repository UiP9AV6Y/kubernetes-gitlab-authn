FROM docker.io/library/golang:1.23.4 AS build

WORKDIR /go/src/github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock
COPY go.mod go.sum ./
RUN set -xe ; \
    go mod download && \
    go mod verify

ENV CGO_ENABLED=0
COPY . .
RUN set -xe ; \
    go build -v \
      -ldflags="-s -w" \
      -o /go/bin/gitlab-mock \
      .

FROM scratch

COPY --from=build /go/bin/gitlab-mock /
ENTRYPOINT ["/gitlab-mock"]

