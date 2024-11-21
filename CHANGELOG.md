## v0.4.0 / 2024-11-21

## Changes

* [FEATURE] add scoping of files by tree
* [UPDATE] update dependencies - golang v1.22.9

## v0.3.8 / 2024-06-24

## Changes

* [FEATURE] add templating of patterns
* [UPDATE] update dependencies


## v0.3.7 / 2024-02-16

## Changes

* [UPDATE] update to latest golang (1.22) and dependencies


## v0.3.6 / 2023-02-28

## Changes

* [BUGFIX] increment all pattern matching file
* [UPDATE] update to latest golang (1.20) and dependencies
* [UPDATE] adapt to latest version of TLS support for exporters


## v0.3.5 / 2021-11-17

## Changes

* [ENHANCEMENT] add support of TLS


## v0.3.4 / 2021-10-06

## Changes

* [BUGFIX] pattern match with base path and directories matching pattern are ignored
* [BUGFIX] reconstruct the fully qualified pathname of pattern


## v0.3.3 / 2021-07-21

KNOMN BUG: pattern with base path don't work.

## Changes

* [ENHANCEMENT] glob format now supports double star and thus recursive walk of directory


## v0.3.2 / 2021-07-20

## Changes

* [UPDATE] update to latest golang (1.16) and dependencies
* [CHANGE] log format changed because of switch to go-kit/log


## v0.3.1 / 2020-01-15

### Changes

* [BUGFIX] crc32 and nb line global configuration not propagated
* [ENHANCEMENT] add listen address and metrics path config from file


## v0.3.0 / 2019-12-18

### Changes

* [ENHANCEMENT] add config file to replace/complement parameters


## v0.2.0 / 2019-12-15

### Changes

* [ENHANCEMENT] add metric `file_content_line_number` for line number in files
* [ENHANCEMENT] add parameter to change working directory of glob pattern
* [BREAKING] rename metric `file_hash_content_crc32` to `file_content_hash_crc32`
* [DOCKER] add docker file to generate image


## v0.1.0 / 2019-12-10

### Changes

* [ENHANCEMENT] add parameter to change working directory of glob pattern
* [BREAKING] rename metric `file_hash_content_crc32` to `file_content_hash_crc32`
* [DOCKER] add docker file to generate image

### Changes

* [ENHANCEMENT] add metric counting the number of files matching glob pattern
* [ENHANCEMENT] add metric describing crc32 hash of file
* [BREAKING] change package content to reflect os/archi


## v0.0.1 / 2019-12-08

### Changes

* [INITIAL] publish basic metrics from `stat()` of file
* [INITIAL] build and release from CircleCI

