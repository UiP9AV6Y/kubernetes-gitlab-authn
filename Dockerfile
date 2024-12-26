FROM docker.io/library/golang:1.23.4 AS build

WORKDIR /go/src/github.com/UiP9AV6Y/kubernetes-gitlab-authn
COPY go.mod go.sum ./
RUN set -xe ; \
    go mod download && \
    go mod verify

COPY .bingo/ .bingo/
RUN set -xe ; \
    go install github.com/bwplotka/bingo@latest && \
    bingo get -v -l

ENV CGO_ENABLED=0
COPY cmd/ cmd/
COPY pkg/ pkg/
# required for buildinfo
COPY .git/ .git/
RUN set -xe ; \
    go generate ./... && \
    go build -v \
      -ldflags="-s -w" \
      -o /go/bin/kubernetes-gitlab-authn \
      ./cmd/kubernetes-gitlab-authn

FROM gcr.io/distroless/static-debian12

LABEL org.opencontainers.image.title="kubernetes-gitlab-authn" \
      org.opencontainers.image.description="Kubernetes authentication service for Gitlab Personal Access Tokens" \
      org.opencontainers.image.authors="Gordon Bleux <33967640+UiP9AV6Y@users.noreply.github.com>" \
      org.opencontainers.image.url="https://github.com/UiP9AV6Y/kubernetes-gitlab-authn" \
      org.opencontainers.image.source="https://github.com/UiP9AV6Y/kubernetes-gitlab-authn.git" \
      org.opencontainers.image.licenses="AGPL-3.0-or-later"

COPY ./web /usr/share/kubernetes-gitlab-authn/public/
COPY ./config.prod.yaml /etc/kubernetes/gitlab-authn.yaml
COPY --from=build /go/bin/kubernetes-gitlab-authn /usr/local/bin/
CMD ["/usr/local/bin/kubernetes-gitlab-authn"]

