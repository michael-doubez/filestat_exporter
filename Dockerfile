FROM golang:1.16-alpine AS build
RUN mkdir /exporter/
WORKDIR /exporter
COPY .git Makefile *.go go.mod go.sum /exporter/
RUN apk add git
RUN apk add build-base
RUN make build RELEASE_MODE=1

FROM alpine
LABEL maintainer="Michael DOUBEZ <michael@doubez.fr>"

COPY --from=build /exporter/filestat_exporter /usr/bin/

USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]
