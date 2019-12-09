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
	"hash/crc32"
	"io"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "file"

var (
	fileSizeBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "stat", "size_bytes"),
		"Size of file in bytes",
		[]string{"path"}, nil,
	)
	fileModificationTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "stat", "modification_time_seconds"),
		"Last modification time of file in epoch time",
		[]string{"path"}, nil,
	)
	fileCRC32HashDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "hash", "crc32_ieee"),
		"CRC32 hash of file content using the IEEE polynomial",
		[]string{"path"}, nil,
	)
)

// Collector compute metrics for each file matching the patterns
type fileStatusCollector struct {
	filesPatterns     []string
	enableCRC32Metric bool
}

// Describe implements the prometheus.Collector interface.
func (c *fileStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fileSizeBytesDesc
	ch <- fileModificationTimeSecondsDesc
	if c.enableCRC32Metric {
		ch <- fileCRC32HashDesc
	}
}

// Collect implements the prometheus.Collector interface.
func (c *fileStatusCollector) Collect(ch chan<- prometheus.Metric) {
	set := make(map[string]struct{})
	for _, pattern := range c.filesPatterns {
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, filePath := range matches {
				log.Debugln("Collecting file ", filePath)
				if _, ok := set[filePath]; !ok {
					set[filePath] = struct{}{}
					collectFileMetrics(ch, filePath)
					if c.enableCRC32Metric {
						collectCRC32Metric(ch, filePath)
					}
				}
			}
		} else {
			log.Debugln("Error getting matches for glob", pattern, "-", err)
		}
	}
}

// Collect metrics for a file and feed
func collectFileMetrics(ch chan<- prometheus.Metric, filePath string) {
	// Metrics based on Fileinfo
	if fileinfo, err := os.Stat(filePath); err == nil {
		if fileinfo.IsDir() {
			return
		}
		ch <- prometheus.MustNewConstMetric(fileSizeBytesDesc, prometheus.GaugeValue,
			float64(fileinfo.Size()),
			filePath)
		modTime := fileinfo.ModTime()
		ch <- prometheus.MustNewConstMetric(fileModificationTimeSecondsDesc, prometheus.GaugeValue,
			float64(modTime.Unix())+float64(modTime.Nanosecond())/1000000000.0,
			filePath)
	} else {
		log.Debugln("Error getting file info for", filePath, "-", err)
		return
	}
}

// Collect metrics for a file and feed
func collectCRC32Metric(ch chan<- prometheus.Metric, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Debugln("Error getting CRC32 file hash when opening", filePath, "-", err)
		return
	}
	defer file.Close()

	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, file); err != nil {
		log.Debugln("Error generating CRC32 hash of file", filePath, "-", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(fileCRC32HashDesc, prometheus.GaugeValue,
		float64(hash.Sum32()),
		filePath)
}
