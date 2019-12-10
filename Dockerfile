FROM golang:1.13-alpine AS build
RUN ls -R
RUN mkdir /build/
WORKDIR /build
COPY Makefile *.go go.mod go.sum /build/
RUN ls -R
RUN apk add git
RUN apk add make
RUN make build RELEASE_MODE=1

FROM alpine
COPY --from=build /exporter/filestat_exporter /usr/bin/
USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]