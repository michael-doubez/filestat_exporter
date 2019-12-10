FROM golang:1.13-alpine AS build
RUN ls -R
RUN mkdir /exporter/
WORKDIR /exporter
COPY Makefile *.go go.mod go.sum /exporter/
RUN ls -R
RUN apk add git
RUN apk add make
RUN apk add gcc
RUN make build RELEASE_MODE=1

FROM alpine
COPY --from=build /exporter/filestat_exporter /usr/bin/
USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]
