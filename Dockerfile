ARG GO_VERSION=latest

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
ARG VERSION

# these will get injected in by Docker
ARG TARGETOS TARGETARC

RUN apt-get install git make

WORKDIR /exporter
COPY Makefile go.mod go.sum /exporter/
COPY .git/ ./.git/
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build RELEASE_MODE=1 VERSION=${VERSION}

FROM scratch
ARG VERSION

LABEL maintainer="Michael DOUBEZ <michael@doubez.fr>"

# OpenContainers Annotations Spec
LABEL org.opencontainers.image.description="Prometheus exporter gathering metrics about file size, modification time and other stats." \
      org.opencontainers.image.licenses=MIT \
      org.opencontainers.image.source=https://github.com/michael-doubez/filestat_exporter \
      org.opencontainers.image.title=filestat_exporter \
      org.opencontainers.image.version=${VERSION}

WORKDIR /usr/bin/
COPY --from=build /exporter/filestat_exporter /usr/bin/filestat_exporter
COPY --from=build /etc/passwd /etc/passwd

USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]
