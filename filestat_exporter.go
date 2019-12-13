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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

func main() {
	commandLine := flag.NewFlagSet("filestat_exporter", flag.ExitOnError)
	var (
		logLevel      = commandLine.String("log.level", "info", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal].")
		crc32Metric   = commandLine.Bool("metric.crc32", false, "Generate CRC32 hash metric of files.")
		lineNbMetric  = commandLine.Bool("metric.nb_lines", false, "Generate line number metric of files.")
		workingDir    = commandLine.String("path.cwd", ".", "Working directory of path pattern collection")
		printVersion  = commandLine.Bool("version", false, "Print the version of the exporter and exit.")
		listenAddress = commandLine.String("web.listen-address", ":9943", "The address to listen on for HTTP requests.")
		metricsPath   = commandLine.String("web.telemetry-path", "/metrics", "The path under which to expose metrics.")
	)
	commandLine.Parse(os.Args[1:])

	if *printVersion {
		fmt.Fprintf(os.Stderr, "%s\n", version.Print("filestat_exporter"))
		os.Exit(0)
	}

	if commandLine.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "filestat_exporter requires at least one argument file to match\n")
		os.Exit(1)
	}

	log.Base().SetLevel(*logLevel)

	// args are glob pattern for files to watch
	if *workingDir != "." {
		if err := os.Chdir(*workingDir); err != nil {
			log.Errorln("Could not change to directory", *workingDir, "-", err)
			os.Exit(1)
		}
	}
	collector := &fileStatusCollector{
		filesPatterns:      commandLine.Args(),
		enableCRC32Metric:  *crc32Metric,
		enableLineNbMetric: *lineNbMetric,
	}
	if err := prometheus.Register(collector); err != nil {
		log.Errorln("Could not register collector", err)
	} else {
		log.Infoln("Collector ready to collect", collector.filesPatterns)
	}

	// setting up exporter
	log.Infoln("Starting file_status_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	actualMetricsPath := path.Clean("/" + *metricsPath)

	http.HandleFunc("/", IndexHandler(actualMetricsPath))
	http.Handle(actualMetricsPath, promhttp.Handler())

	// run exporter
	log.Infoln("Listening on", *listenAddress)
	server := &http.Server{Addr: *listenAddress}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
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
