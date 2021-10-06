## v0.3.4 / 20211006

## Changes

* [BUGFIX] pattern match with base path and directories matching pattern are ignored
* [BUGFIX] reconstruct the fully qualified pathname of pattern


## v0.3.3 / 20210721

KNOMN BUG: pattern with base path don't work.

## Changes

* [ENHANCEMENT] glob format now supports double star and thus recursive walk of directory


## v0.3.2 / 20210720

## Changes

* [UPDATE] update to latest golang (1.16) and dependencies
* [CHANGE] log format changed because of switch to go-kit/log


## v0.3.1 / 20200115

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

