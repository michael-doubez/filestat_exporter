FROM golang:1.20-alpine3.17 AS build
RUN apk add git
RUN apk add build-base

WORKDIR /exporter
COPY .git Makefile *.go go.mod go.sum /exporter/
RUN make build RELEASE_MODE=1

FROM alpine:3.13
LABEL maintainer="Michael DOUBEZ <michael@doubez.fr>"

COPY --from=build /exporter/filestat_exporter /usr/bin/

USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]
