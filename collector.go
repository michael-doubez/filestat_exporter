// Copyright 2019 Michael DOUBEZ
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "file"

var (
	fileMatchingGlobNbDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "glob", "match_number"),
		"Number of files matching pattern",
		[]string{"pattern"}, nil,
	)
	fileSizeBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "stat", "size_bytes"),
		"Size of file in bytes",
		[]string{"path"}, nil,
	)
	fileModifTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "stat", "modif_time_seconds"),
		"Last modification time of file in epoch time",
		[]string{"path"}, nil,
	)
	fileCRC32HashDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "content", "hash_crc32"),
		"CRC32 hash of file content using the IEEE polynomial",
		[]string{"path"}, nil,
	)
	lineNbMetricDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "content", "line_number"),
		"Number of lines in file",
		[]string{"path"}, nil,
	)
)

// Collector compute metrics for each file matching the patterns
type fileStatusCollector struct {
	filesPatterns      []string
	enableCRC32Metric  bool
	enableLineNbMetric bool
}

// Describe implements the prometheus.Collector interface.
func (c *fileStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fileMatchingGlobNbDesc
	ch <- fileSizeBytesDesc
	ch <- fileModifTimeSecondsDesc
	if c.enableCRC32Metric {
		ch <- fileCRC32HashDesc
	}
	if c.enableLineNbMetric {
		ch <- lineNbMetricDesc
	}
}

// Collect implements the prometheus.Collector interface.
func (c *fileStatusCollector) Collect(ch chan<- prometheus.Metric) {
	set := make(map[string]struct{})
	for _, pattern := range c.filesPatterns {
		matchingFileNb := 0
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, filePath := range matches {
				log.Debugln("Collecting file ", filePath)
				if _, ok := set[filePath]; !ok {
					set[filePath] = struct{}{}
					collectFileMetrics(ch, filePath, &matchingFileNb)
					if c.enableCRC32Metric || c.enableLineNbMetric {
						collectContentMetrics(ch, filePath,
							c.enableCRC32Metric,
							c.enableLineNbMetric)
					}
				}
			}
		} else {
			log.Debugln("Error getting matches for glob", pattern, "-", err)
		}
		ch <- prometheus.MustNewConstMetric(fileMatchingGlobNbDesc, prometheus.GaugeValue,
			float64(matchingFileNb),
			pattern)
	}
}

// Collect metrics for a file and feed
func collectFileMetrics(ch chan<- prometheus.Metric, filePath string, nbFile *int) {
	// Metrics based on Fileinfo
	if fileinfo, err := os.Stat(filePath); err == nil {
		if fileinfo.IsDir() {
			return
		}
		*nbFile++
		ch <- prometheus.MustNewConstMetric(fileSizeBytesDesc, prometheus.GaugeValue,
			float64(fileinfo.Size()),
			filePath)
		modTime := fileinfo.ModTime()
		ch <- prometheus.MustNewConstMetric(fileModifTimeSecondsDesc, prometheus.GaugeValue,
			float64(modTime.Unix())+float64(modTime.Nanosecond())/1000000000.0,
			filePath)
	} else {
		log.Debugln("Error getting file info for", filePath, "-", err)
		return
	}
}

// Collect metrics for a file content
func collectContentMetrics(ch chan<- prometheus.Metric, filePath string,
	enableCRC32 bool, enableLineNb bool) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Debugln("Error getting content file hash when opening", filePath, "-", err)
		return
	}
	defer file.Close()

	hash := crc32.NewIEEE()
	lineNb := 0

	// read chunks of 32k
	buf := make([]byte, 32*1024)
	lineSep := []byte{'\n'}

ReadFile:
	for {
		c, err := file.Read(buf)
		slice := buf[:c]
		if enableLineNb {
			lineNb += bytes.Count(slice, lineSep)
		}
		if enableCRC32 {
			if _, errHash := hash.Write(slice); errHash != nil {
				log.Debugln("Error generating CRC32 hash of file", filePath, "-", errHash)
				enableCRC32 = false
			}
		}

		switch {
		case err == io.EOF:
			break ReadFile

		case err != nil:
			log.Debugln("Error reading content of file", filePath, "-", err)
			return
		}
	}

	if enableCRC32 {
		ch <- prometheus.MustNewConstMetric(fileCRC32HashDesc, prometheus.GaugeValue,
			float64(hash.Sum32()),
			filePath)
	}
	if enableLineNb {
		ch <- prometheus.MustNewConstMetric(lineNbMetricDesc, prometheus.GaugeValue,
			float64(lineNb),
			filePath)
	}
}
