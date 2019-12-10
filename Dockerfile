FROM golang:1.13-alpine AS build
RUN mkdir /build/
RUN apk add git
RUN make build BUILD_DEST=/build/

FROM alpine
COPY --from=build /build/filestat_exporter /usr/bin/
ENTRYPOINT ["/usr/bin/filestat_exporter"]
