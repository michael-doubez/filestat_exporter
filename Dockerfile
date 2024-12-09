FROM golang:1.23.4 AS build
RUN apt-get install git make

WORKDIR /exporter
COPY Makefile go.mod go.sum /exporter/
COPY .git/ ./.git/
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN ls -al
RUN make build RELEASE_MODE=1

FROM scratch
LABEL maintainer="Michael DOUBEZ <michael@doubez.fr>"

COPY --from=build /exporter/filestat_exporter /usr/bin/

USER nobody
EXPOSE 9943
ENTRYPOINT ["/usr/bin/filestat_exporter"]
