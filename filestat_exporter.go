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
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
)

const (
	defaultConfigFile    = "filestat.yaml"
	defaultLogLevel      = "info"
	defaultWorkingDir    = "."
	defaultListenAddress = ":9943"
	defaultMetricsPath   = "/metrics"
)

func main() {
	commandLine := flag.NewFlagSet("filestat_exporter", flag.ExitOnError)
	var (
		cfgFile       = commandLine.String("config.file", defaultConfigFile, "The path to the configuration file (use \"none\" to disable).")
		logLevel      = commandLine.String("log.level", defaultLogLevel, "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error].")
		crc32Metric   = commandLine.Bool("metric.crc32", false, "Generate CRC32 hash metric of files.")
		lineNbMetric  = commandLine.Bool("metric.nb_lines", false, "Generate line number metric of files.")
		workingDir    = commandLine.String("path.cwd", defaultWorkingDir, "Working directory of path pattern collection")
		printVersion  = commandLine.Bool("version", false, "Print the version of the exporter and exit.")
		listenAddress = commandLine.String("web.listen-address", defaultListenAddress, "The address to listen on for HTTP requests.")
		metricsPath   = commandLine.String("web.telemetry-path", defaultMetricsPath, "The path under which to expose metrics.")
	)
	commandLine.Parse(os.Args[1:])

	if *printVersion {
		fmt.Fprintf(os.Stderr, "%s\n", version.Print("filestat_exporter"))
		os.Exit(0)
	}

	// configuration
	defaultCollector := &collectorConfig{
		collectorMetricConfig: collectorMetricConfig{
			EnableCRC32Metric:  crc32Metric,
			EnableNbLineMetric: lineNbMetric,
		},
		GlobPatternPath: commandLine.Args(),
	}

	promlogConfig := &promlog.Config{}
	promlogConfig.Level = &promlog.AllowedLevel{}
	promlogConfig.Format = &promlog.AllowedFormat{}

	if err := promlogConfig.Level.Set(*logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Wrong loglevel parameter - %s\n", err)
		os.Exit(1)
	}
	promlogConfig.Format.Set("logfmt")

	logger := promlog.New(promlogConfig)

	config, err := readConfig(*cfgFile, defaultCollector, logger)
	if config == nil {
		level.Error(logger).Log("msg", "Error reading config", "file", *cfgFile, "reason", err)
		os.Exit(1)
	}

	if len(config.Exporter.Files) == 0 {
		level.Error(logger).Log("msg", "filestat_exporter requires a config file with patterns or at least one argument file to match")
		os.Exit(1)
	}

	// adjust working directory globally
	if *workingDir != defaultWorkingDir {
		if len(config.Exporter.WorkingDirectory) != 0 {
			level.Info(logger).Log("msg", "Config override", "from", "parameter", "working_directory", *workingDir)
		}
		config.Exporter.WorkingDirectory = *workingDir
	}
	if len(config.Exporter.WorkingDirectory) != 0 {
		if err := os.Chdir(config.Exporter.WorkingDirectory); err != nil {
			level.Error(logger).Log("msg", "Could not change to directory", "path", config.Exporter.WorkingDirectory, "reason", err)
			os.Exit(1)
		}
	}

	// create collector
	collector := config.generateCollector(logger)
	if err := prometheus.Register(collector); err != nil {
		level.Error(logger).Log("msg", "Could not register collector", "reason", err)
	} else {
		level.Info(logger).Log("msg", "Collector ready to collect files")
	}

	// setting up exporter
	level.Info(logger).Log("msg", "Starting file_status_exporter", "version", version.Info(), "build", version.BuildContext())

	// metrics path
	hasMetricsPathConfig := len(config.Exporter.MetricsPath) != 0
	if *metricsPath != defaultMetricsPath || !hasMetricsPathConfig {
		if hasMetricsPathConfig {
			level.Info(logger).Log("msg", "Config override", "from", "parameter", "metrics_path", *metricsPath)
		}
		config.Exporter.MetricsPath = *metricsPath
	}
	actualMetricsPath := path.Clean("/" + config.Exporter.MetricsPath)
	level.Info(logger).Log("msg", "Path to metrics", "path", actualMetricsPath)

	// listenAddress
	hasListenAddrConfig := len(config.Exporter.ListenAddress) != 0
	if *listenAddress != defaultListenAddress || !hasListenAddrConfig {
		if hasListenAddrConfig {
			level.Info(logger).Log("msg", "Config override", "from", "parameter", "listen_address", *listenAddress)
		}
		config.Exporter.ListenAddress = *listenAddress
	}

	http.HandleFunc("/", IndexHandler(actualMetricsPath))
	http.Handle(actualMetricsPath, promhttp.Handler())

	// run exporter
	level.Info(logger).Log("msg", "Listening", "port", config.Exporter.ListenAddress)
	server := &http.Server{Addr: config.Exporter.ListenAddress}
	if err := server.ListenAndServe(); err != nil {
		level.Error(logger).Log("msg", "Listening error", "reason", err)
	}
}

// IndexHandler returns a handler for root display
func IndexHandler(metricsPath string) http.HandlerFunc {
	indexHTML := `<html>
  <head><title>File Status Exporter</title></head>
  <body>
    <h1>File Status Exporter</h1>
    <p><a href="%s">Metrics</a></p>
  </body>
</html>
`
	index := []byte(fmt.Sprintf(indexHTML, metricsPath))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Write(index)
	}
}
