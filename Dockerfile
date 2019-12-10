FROM golang:1.13-alpine AS build
RUN ls -R
RUN mkdir /exporter/
WORKDIR /exporter
COPY .git Makefile *.go go.mod go.sum /exporter/
RUN apk add git
RUN apk add build-base
RUN make build RELEASE_MODE=1

FROM alpine
COPY --from=build /exporter/filestat_exporter /usr/bin/
USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]
