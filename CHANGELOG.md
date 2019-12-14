## 0.2.0 / 2019-12-15

### Changes

* [ENHANCEMENT] add metric `file_content_line_number` for line number in files
* [ENHANCEMENT] add parameter to change working directory of glob pattern
* [BREAKING] rename metric `file_hash_content_crc32` to `file_content_hash_crc32`
* [DOCKER] add docker file to generate image


## 0.1.0 / 2019-12-10

## master / unreleased

### Changes

* [ENHANCEMENT] add parameter to change working directory of glob pattern
* [BREAKING] rename metric `file_hash_content_crc32` to `file_content_hash_crc32`
* [DOCKER] add docker file to generate image

### Changes

* [ENHANCEMENT] add metric counting the number of files matching glob pattern
* [ENHANCEMENT] add metric describing crc32 hash of file
* [BREAKING] change package content to reflect os/archi


## 0.0.1 / 2019-12-08

### Changes

* [INITIAL] publish basic metrics from `stat()` of file
* [INITIAL] build and release from CircleCI

