FROM golang:1.13-alpine AS build
RUN mkdir /build/
RUN apk add git
RUN apk add make
RUN ls -R .
RUN make build BUILD_DEST=/exporter/ RELEASE_MODE=1

FROM alpine
COPY --from=build /exporter/filestat_exporter /usr/bin/
USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]
