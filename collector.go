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
)

// Collector compute metrics for each file matching the patterns
type fileStatusCollector struct {
	filesPatterns []string
}

// Describe implements the prometheus.Collector interface.
func (c *fileStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fileSizeBytesDesc
}

// Collect implements the prometheus.Collector interface.
func (c *fileStatusCollector) Collect(ch chan<- prometheus.Metric) {
	set := make(map[string]struct{})
	for _, pattern := range c.filesPatterns {
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, file := range matches {
				log.Debugln("Collecting file ", file)
				if _, ok := set[file]; !ok {
					set[file] = struct{}{}
					collectFileMetrics(ch, file)
				}
			}
		} else {
			log.Debugln("Error getting matches for glob", pattern, "-", err)
		}
	}
}

// Collect metrics for a file and feed
func collectFileMetrics(ch chan<- prometheus.Metric, file string) {
	// Metrics based on Fileinfo
	if fileinfo, err := os.Stat(file); err == nil {
		if fileinfo.IsDir() {
			return
		}
		ch <- prometheus.MustNewConstMetric(fileSizeBytesDesc, prometheus.GaugeValue,
			float64(fileinfo.Size()),
			file)
		modTime := fileinfo.ModTime()
		ch <- prometheus.MustNewConstMetric(fileModificationTimeSecondsDesc, prometheus.GaugeValue,
			float64(modTime.Unix())+float64(modTime.Nanosecond())/1000000000.0,
			file)
	} else {
		log.Debugln("Error getting file info for", file, "-", err)
		return
	}
}
