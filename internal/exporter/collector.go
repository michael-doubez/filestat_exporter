// Copyright 2019-2025 Michael DOUBEZ
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

package exporter

import (
	"bytes"
	"hash/crc32"
	"io"
	"log/slog"
	"os"
	"path"
	"slices"
	"text/template"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "file"

var (
	fileMatchingGlobNbOpts = prometheus.Opts{
		Namespace: namespace,
		Subsystem: "glob",
		Name:      "match_number",
		Help:      "Number of files matching pattern",
	}
	fileSizeBytesOpts = prometheus.Opts{
		Namespace: namespace,
		Subsystem: "stat",
		Name:      "size_bytes",
		Help:      "Size of file in bytes",
	}
	fileModifTimeSecondsOpts = prometheus.Opts{
		Namespace: namespace,
		Subsystem: "stat",
		Name:      "modif_time_seconds",
		Help:      "Last modification time of file in epoch time",
	}
	fileCRC32HashOpts = prometheus.Opts{
		Namespace: namespace,
		Subsystem: "content",
		Name:      "hash_crc32",
		Help:      "CRC32 hash of file content using the IEEE polynomial",
	}
	lineNbMetricOpts = prometheus.Opts{
		Namespace: namespace,
		Subsystem: "content",
		Name:      "line_number",
		Help:      "Number of lines in file",
	}
)

// Transform opts into desc
func optsToDesc(opts *prometheus.Opts, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		labels,
		opts.ConstLabels)
}

// Collector compute metrics for each file matching the patterns in tree
type fileStatCollector struct {
	enableCRC32Metric  bool
	enableLineNbMetric bool
	labels             []string

	treeRoot string

	filesPatterns []string
}

// Collector compute metrics for each tree
type treeCollector struct {
	collectors []fileStatCollector
}

// Files collector
type filesCollector struct {
	trees  map[string]*treeCollector
	common []string

	fileMatchingGlobNbDesc   *prometheus.Desc
	fileSizeBytesDesc        *prometheus.Desc
	fileModifTimeSecondsDesc *prometheus.Desc
	fileCRC32HashDesc        *prometheus.Desc
	lineNbMetricDesc         *prometheus.Desc

	logger slog.Logger
}

func createFilesCollector(logger slog.Logger, hasTree bool) *filesCollector {
	c := filesCollector{}
	c.trees = make(map[string]*treeCollector)
	c.common = []string{}
	c.logger = logger

	if hasTree {
		c.common = append(c.common, "tree")
	}

	patternLabels := slices.Concat([]string{"pattern"}, c.common)
	c.fileMatchingGlobNbDesc = optsToDesc(&fileMatchingGlobNbOpts, patternLabels)

	pathLabels := slices.Concat([]string{"path"}, c.common)
	c.fileSizeBytesDesc = optsToDesc(&fileSizeBytesOpts, pathLabels)
	c.fileModifTimeSecondsDesc = optsToDesc(&fileModifTimeSecondsOpts, pathLabels)

	return &c
}

// add file start collector in given tree
func (c *filesCollector) addFileStatCollector(treeName *string, col fileStatCollector) {
	name := ""
	if treeName != nil {
		name = *treeName
	}
	tree, found := c.trees[name]
	if !found {
		tree = &treeCollector{}
		c.trees[name] = tree
	}
	tree.collectors = append(tree.collectors, col)
}

// initialize usage of crc32 hash metric
func (c *filesCollector) useFileCRC32Metric() {
	if c.fileCRC32HashDesc != nil {
		return
	}
	pathLabels := slices.Concat([]string{"path"}, c.common)
	c.fileCRC32HashDesc = optsToDesc(&fileCRC32HashOpts, pathLabels)

}

// initialize usage of line number metrics
func (c *filesCollector) useLineNbMetric() {
	if c.lineNbMetricDesc != nil {
		return
	}
	pathLabels := slices.Concat([]string{"path"}, c.common)
	c.lineNbMetricDesc = optsToDesc(&lineNbMetricOpts, pathLabels)
}

// Describe implements the prometheus.Collector interface.
func (c *filesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.fileMatchingGlobNbDesc
	ch <- c.fileSizeBytesDesc
	ch <- c.fileModifTimeSecondsDesc
	if c.fileCRC32HashDesc != nil {
		ch <- c.fileCRC32HashDesc
	}
	if c.lineNbMetricDesc != nil {
		ch <- c.lineNbMetricDesc
	}
}

// Collect implements the prometheus.Collector interface.
func (c *filesCollector) Collect(ch chan<- prometheus.Metric) {
	templater := template.New("pattern").Funcs(
		template.FuncMap{
			"now":      time.Now,
			"sub":      func(a, b int) int { return a - b },
			"add":      func(a, b int) int { return a + b },
			"subMonth": func(a time.Month, b int) int { return int(a) - b },
			"addMonth": func(a time.Month, b int) int { return int(a) + b },
		})
	for _, tree := range c.trees {
		c.CollectTree(ch, templater, tree)
	}
}

// apply template on path
func apply(templater *template.Template, pattern string) (string, error) {
	p, err := templater.Parse(pattern)
	if err != nil {
		return pattern, err
	}

	var pout bytes.Buffer
	if err = p.Execute(&pout, nil); err != nil {
		return pattern, err
	}

	return pout.String(), nil
}

// CollectTree implements the prometheus.Collector interface per tree.
func (c *filesCollector) CollectTree(ch chan<- prometheus.Metric, templater *template.Template, tree *treeCollector) {
	patternSet := make(map[string]struct{})
	fileSet := make(map[string]bool)
	for _, collector := range tree.collectors {
		treeRoot, err := apply(templater, collector.treeRoot)
		if err != nil {
			c.logger.Warn("Error applying template on tree root", "tree_root", treeRoot, "reason", err)
			continue
		}
		if len(treeRoot) != 0 {
			if _, err := os.Stat(treeRoot); os.IsNotExist(err) {
				c.logger.Debug("Skip collecting file stats because tree root not found", "tree_root", treeRoot)
				continue
			}
		}
		for _, pattern := range collector.filesPatterns {
			// expanded pattern
			realPattern, err := apply(templater, pattern)
			if err != nil {
				c.logger.Warn("Error applying template on file pattern", "pattern", pattern, "reason", err)
				continue
			}

			// only collect pattern once
			fullPattern := path.Join(treeRoot, realPattern)
			if _, ok := patternSet[fullPattern]; ok {
				continue
			}
			patternSet[fullPattern] = struct{}{}

			// get files matching pattern
			matchingFileNb := 0
			basepath, patternPart := doublestar.SplitPattern(realPattern)

			// apply treeRoot
			patternRoot := path.Join(treeRoot, basepath)
			fsys := os.DirFS(patternRoot)
			if matches, err := doublestar.Glob(fsys, patternPart); err == nil {
				for _, relFilePath := range matches {
					realFilePath := path.Join(patternRoot, relFilePath)
					filePath := path.Join(basepath, relFilePath)
					// only collect files once
					if isProcessable, ok := fileSet[realFilePath]; ok {
						if isProcessable {
							matchingFileNb++
						}
						continue
					}

					isFileProcessed := c.collectFileMetrics(ch, filePath, realFilePath, &matchingFileNb, collector.labels)
					fileSet[realFilePath] = isFileProcessed
					if isFileProcessed {
						if collector.enableCRC32Metric || collector.enableLineNbMetric {
							c.collectContentMetrics(ch, filePath, realFilePath,
								collector.enableCRC32Metric,
								collector.enableLineNbMetric,
								collector.labels)
						}
					}
				}
			} else {
				c.logger.Debug("Error getting matches for glob", "pattern", pattern, "reason", err)
			}
			ch <- prometheus.MustNewConstMetric(c.fileMatchingGlobNbDesc, prometheus.GaugeValue,
				float64(matchingFileNb),
				slices.Concat([]string{pattern}, collector.labels)...)
		}
	}
}

// Collect metrics for a file and feed
func (c *filesCollector) collectFileMetrics(ch chan<- prometheus.Metric, filePath string, realFilePath string, nbFile *int, labels []string) bool {
	// Metrics based on Fileinfo
	if fileinfo, err := os.Stat(realFilePath); err == nil {
		if fileinfo.IsDir() {
			return false
		}
		*nbFile++
		metricLabels := slices.Concat([]string{filePath}, labels)
		ch <- prometheus.MustNewConstMetric(c.fileSizeBytesDesc, prometheus.GaugeValue,
			float64(fileinfo.Size()),
			metricLabels...)
		modTime := fileinfo.ModTime()
		ch <- prometheus.MustNewConstMetric(c.fileModifTimeSecondsDesc, prometheus.GaugeValue,
			float64(modTime.Unix())+float64(modTime.Nanosecond())/1000000000.0,
			metricLabels...)
	} else {
		c.logger.Debug("Error getting file info", "path", realFilePath, "reason", err)
		return false
	}
	return true
}

// Collect metrics for a file content
func (c *filesCollector) collectContentMetrics(ch chan<- prometheus.Metric, filePath string, realFilePath string,
	enableCRC32 bool, enableLineNb bool, labels []string) {
	file, err := os.Open(realFilePath)
	if err != nil {
		c.logger.Debug("Error getting content file hash while opening", "path", realFilePath, "reason", err)
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
		b, err := file.Read(buf)
		slice := buf[:b]
		if enableLineNb {
			lineNb += bytes.Count(slice, lineSep)
		}
		if enableCRC32 {
			if _, errHash := hash.Write(slice); errHash != nil {
				c.logger.Debug("Error generating CRC32 hash of file", "path", realFilePath, "reason", errHash)
				enableCRC32 = false
			}
		}

		switch {
		case err == io.EOF:
			break ReadFile

		case err != nil:
			c.logger.Debug("Error reading content of file", "path", realFilePath, "reason", err)
			return
		}
	}

	metricLabels := slices.Concat([]string{filePath}, labels)
	if enableCRC32 {
		ch <- prometheus.MustNewConstMetric(c.fileCRC32HashDesc, prometheus.GaugeValue,
			float64(hash.Sum32()),
			metricLabels...)
	}
	if enableLineNb {
		ch <- prometheus.MustNewConstMetric(c.lineNbMetricDesc, prometheus.GaugeValue,
			float64(lineNb),
			metricLabels...)
	}
}
