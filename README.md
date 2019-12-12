# File statistics exporter

[![CircleCI](https://circleci.com/gh/michael-doubez/filestat_exporter/tree/master.svg?style=shield)][circleci]
[![Docker Pulls](https://img.shields.io/docker/pulls/mdoubez/filestat_exporter.svg?maxAge=604800)][dockerhub]
[![Go Report Card](https://goreportcard.com/badge/github.com/michael-doubez/filestat_exporter)][goreportcard]

Prometheus exporter gathering metrics about file size, modification and other statistics.

## Usage

Configure target files on command line, passing glob patterns in parameters

    ./filestat_exporter './*.*'

Optional flags:
* __`-log.level`:__ Logging level \[debug, info, warn, error, fatal\]. (default: `info`)
* __`-version`:__ Print the version of the exporter and exit.
* __`-web.listen-address`:__ Address to listen on for web interface and telemetry. (default: `9943`)
* __`-web.telemetry-path <URL ptath>`:__ Path under which to expose metrics. (default: `/metrics`)
* __`-metric.crc32`:__ Generate CRC32 hash metric of files.


## Exported Metrics

| Metric                       | Description                                  | Labels   |
| ---------------------------- | -------------------------------------------- | -------- |
| file_glob_match_number       | Number of files matching pattern             | pattern  |
| file_stat_size_bytes         | Size of file in bytes                        | path     |
| file_stat_modif_time_seconds | Last modification time of file in epoch time | path     |
| file_content_hash_crc32      | CRC32 hash of file content                   | path     |

## Building and running

Prerequisites:

* [Go compiler](https://golang.org/dl/) - currently version v1.13
* Linux with make installed

### Building

    go get github.com/michael.doubez/filestat_exporter
    cd ${GOPATH-$HOME/go}/src/github.com/michael.doubez/filestat_exporter
    make
    ./filestat_exporter <flags> <file glob patterns ...>

To see all available configuration flags:

    ./filestat_exporter -h

### Running checks and tests

    make check

### Cross compiled distribution

To build all distribustion packages

    make dist

To build a specific os/architecture package

    make dist-<os>-<archi>
    make dist-linux-amd64
    ...

## Using Docker
The `filestat_exporter` is designed to monitor files on the host system.

Try it out in minutes on [Katakoda docker playground][dockerplay]:
```bash
# create locale files
docker container run --rm -d -v ~/my_files:/my_files --name my_files bash -c 'echo "Hello world" > /my_files/sample.txt'
# launch exporter watching the files
docker run -d -p 9943:9943 --name=filestats -v ~/my_files:/data mdoubez/filestat_exporter -path.cwd /data '*'
# see file metrics
curl -s docker:9943/metrics | grep file_
```

[circleci]: https://circleci.com/gh/michael-doubez/filestat_exporter
[dockerhub]: https://hub.docker.com/r/mdoubez/filestat_exporter/
[goreportcard]: https://goreportcard.com/report/github.com/michael-doubez/filestat_exporter
[dockerplay]: https://www.katacoda.com/courses/docker/playground

