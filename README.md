# File statistics exporter

[![CircleCI](https://circleci.com/gh/michael-doubez/filestat_exporter/tree/master.svg?style=shield)][circleci]
[![Docker Pulls](https://img.shields.io/docker/pulls/mdoubez/filestat_exporter.svg?maxAge=604800)][dockerhub]
[![GitHub All Releases](https://img.shields.io/github/downloads/michael-doubez/filestat_exporter/total)][releases]
[![Go Report Card](https://goreportcard.com/badge/github.com/michael-doubez/filestat_exporter)][goreportcard]

Prometheus exporter gathering metrics about file size, modification time and other statistics.

## Quickstart

Pre-built binaries are available on the [GitHub release page][releases].

### Usage

Configure target files on command line, passing glob patterns in parameters

    ./filestat_exporter '*'

Optional flags:
* __`-config.file <yaml>`:__ The path to the configuration file (use "none" to disable).
* __`-debug`:__ Activate debug mode which forces log level to debug and enables pprof.
* __`-log.level <level>`:__ Logging level \[debug, info, warn, error\]. (default: `info`)
* __`-version`:__ Print the version of the exporter and exit.
* __`-web.listen-address <port>`:__ Address to listen on for web interface and telemetry. (default: `9943`)
* __`-web.telemetry-path <URL path>`:__ Path under which to expose metrics. (default: `/metrics`)
* __`-web.config <web file>`:__ Path to config yaml file that can enable TLS or authentication.
* __`-web.systemd-socket`:__ Flag to use systemd socket activation listeners instead of port listeners (Linux only).
* __`-path.cwd <path>`:__ Change working directory of path pattern collection.
* __`-tree.name <name>`:__ Change default tree name used as `tree` label in metrics
* __`-tree.root <path>`:__ Chnage default root path of files
* __`-metric.crc32`:__ Generate CRC32 hash metric of files.
* __`-metric.nb_lines`:__ Generate line number metric of files.

The exporter can read a config file in yaml format (`filestat.yaml` by default).

```yaml
exporter:
  # Optional working directory - overridden by parameter '-path.cwd'
  working_directory: "/path/to/my/project"
  # Optional network parameters
  listen_address: ':9943'
  #metrics_path: /metrics
  
  # Optional working directory - overridden by parameter '-path.cwd'
  working_directory: "/path/to/my/project"
  # Optional default tree name and root path - overridden by parameter '-root.name' and '-root.path'
  #tree_name: ""
  #tree_root: ""

  # Default enable/disable of metrics - overridden if not set by parameter '-metric.*'
  enable_crc32_metric: true
  # enable_nb_line_metric: false
  # list of patterns to apply - metrics can be enable/disabled for each group
  files:
    - patterns: ["*.html","assets/*.css","scripts/*.js"]
    - patterns: ["data/*.csv"]
      enable_nb_line_metric: true
    - patterns: ["archives/*.tar.gz"]
      enable_crc32_metric: false
      enable_nb_line_metric: false

  # other trees
  trees: []
```

Notes:

  - if a file is matched by a pattern more than once, only the first match's config is used
  - if no tree name is defined, the label is not used

### Pattern format

Pattern uses the glob implementation of [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar#patterns) project:
* Doublestar (`**`) can be used to match directories recursively. It should appear surrounded by path separators such as `/**/`.
* Usual [Glob syntax](https://en.wikipedia.org/wiki/Glob_(programming)#Syntax) is still supported.

Pattern can also use golang template format with the following functions:

| Function | Description                                  | Example                       |
|----------|----------------------------------------------|-------------------------------|
| now      | Current [time](https://pkg.go.dev/time#Time) | `{{ now.Locale.Year }}/*.tgz` |
| add      | Add two int together                         | `{{ add 1 1 }}`               |
| sub      | Subtracts two int together                   | `{{ sub 1 1 }}`               |
| addMonth | Add int to time.Month                        | `{{ addMonth now.Month 1 }}`  |
| subMonth | Subtract int from time.Month                 | `{{ subMonth now.Month 1 }}`  |

### Trees

Trees extract files' statistics relatively to a tree root.
If tree root is empty, it is not applied. The name of the
tree is used for scoping the metrics.

If defined, the tree root must exists. It can be templated but usual patterns are not used.

```yaml
- tree_name: name of tree   # optional
  tree_root: path/to/tree/  # optional
  #enable_*_metric: true|false # default for tree
  files: [] # as usual
```


### Exported Metrics

| Metric                         | Description                                  | Labels             |
| ------------------------------ | -------------------------------------------- | ------------------ |
| `file_glob_match_number`       | Number of files matching pattern             | `tree`, `pattern`  |
| `file_stat_size_bytes`         | Size of file in bytes                        | `tree`, `path`     |
| `file_stat_modif_time_seconds` | Last modification time of file in epoch time | `tree`, `path`     |
| `file_content_hash_crc32`  (*) | CRC32 hash of file content                   | `tree`, `path`     |
| `file_content_line_number` (*) | Number of lines in file                      | `tree`, `path`     |

Note: metrics with `(*)` are only provided if configured


## Building and running

Prerequisites:

* [Go compiler](https://golang.org/dl/) - currently version v1.22
* Linux with make installed
* Essential build environment for dependencies

### Building

    go get github.com/michael-doubez/filestat_exporter
    cd ${GOPATH-$HOME/go}/src/github.com/michael-doubez/filestat_exporter
    make
    ./filestat_exporter <flags> <file glob patterns ...>

To see all available configuration flags:

    ./filestat_exporter -h

The Makefile provides several targets:
* `make check`: Running checks and tests
* `make run`: Run exporter from go
* `make version`: Print current version
* `make build`: Build exporter without checks

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

# create local file
docker container run --rm -d -v ~/my_files:/my_files --name my_files bash -c 'echo "Hello world" > /my_files/sample.txt'
# launch exporter watching the files
docker run -d -p 9943:9943 --name=filestats -v ~/my_files:/data mdoubez/filestat_exporter -path.cwd /data '*'
# see file metrics
curl -s docker:9943/metrics | grep file_
```

## TLS and basic authentication

Filestat Exporter supports TLS and basic authentication. This enables better
control of the various HTTP endpoints.

To use TLS and/or basic authentication, you need to pass a configuration file
using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).


## License

Apache License 2.0, see [LICENSE](https://github.com/michael-doubez/filestat_exporter/blob/master/LICENSE).


[circleci]: https://circleci.com/gh/michael-doubez/filestat_exporter
[dockerhub]: https://hub.docker.com/r/mdoubez/filestat_exporter/
[goreportcard]: https://goreportcard.com/report/github.com/michael-doubez/filestat_exporter
[dockerplay]: https://www.katacoda.com/courses/docker/playground
[releases]: https://github.com/michael-doubez/filestat_exporter/releases
