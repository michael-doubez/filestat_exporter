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

	"github.com/prometheus/common/log"
	yaml "gopkg.in/yaml.v2"
)

type collectorMetricConfig struct {
	EnableCRC32Metric  *bool `yaml:"enable_crc32_metric,omitempty"`
	EnableNbLineMetric *bool `yaml:"enable_nb_line_metric,omitempty"`
}

type configExporter struct {
	collectorMetricConfig `yaml:",inline"`

	ListenAddress    string `yaml:"listen_address,omitempty"`
	MetricsPath      string `yaml:"metrics_path,omitempty"`
	WorkingDirectory string `yaml:"working_directory,omitempty"`

	Files []*collectorConfig `yaml:"files"`
}

type collectorConfig struct {
	collectorMetricConfig `yaml:",inline"`

	GlobPatternPath []string `yaml:"patterns"`
}

type configContent struct {
	Exporter configExporter `yaml:"exporter"`
}

func mergeCollectorMetrics(collector *collectorMetricConfig, defaultCollector *collectorMetricConfig) {
	if collector.EnableCRC32Metric == nil {
		collector.EnableCRC32Metric = defaultCollector.EnableCRC32Metric
	}
	if collector.EnableNbLineMetric == nil {
		collector.EnableNbLineMetric = defaultCollector.EnableNbLineMetric
	}
}

func readConfig(cfgFile string, defaultCollector *collectorConfig) (*configContent, error) {
	cfg := &configContent{}

	// read file if possible
	if cfgFile != "none" {
		info, err := os.Stat(cfgFile)
		if err == nil && !info.IsDir() {
			log.Infoln("Reading config file", cfgFile)
			r, err := os.Open(cfgFile)
			if err != nil {
				return nil, err
			}
			decoder := yaml.NewDecoder(r)
			decoder.SetStrict(true)
			err = decoder.Decode(cfg)
			if err != nil {
				return nil, err
			}
		} else {
			log.Infoln("Could not read config file", cfgFile)
		}
	}

	// merge default config
	if cfg.Exporter.EnableCRC32Metric == nil {
		log.Infoln("Config from parameter: enable_crc32_metric =", *defaultCollector.EnableCRC32Metric)
	} else {
		log.Infoln("General config: enable_crc32_metric =", *cfg.Exporter.EnableCRC32Metric)
	}
	if cfg.Exporter.EnableNbLineMetric == nil {
		log.Infoln("Config from parameter: enable_nb_line_metric =", *defaultCollector.EnableNbLineMetric)
	} else {
		log.Infoln("General config: enable_nb_line_metric =", *cfg.Exporter.EnableNbLineMetric)
	}
	mergeCollectorMetrics(&cfg.Exporter.collectorMetricConfig, &defaultCollector.collectorMetricConfig)

	// patterns from command line
	if len(defaultCollector.GlobPatternPath) != 0 {
		log.Infoln("Adding collection of patterns from command line")
		cfg.Exporter.Files = append(cfg.Exporter.Files, defaultCollector)
	}

	// update collectors with general config
	for _, collector := range cfg.Exporter.Files {
		mergeCollectorMetrics(&collector.collectorMetricConfig, &cfg.Exporter.collectorMetricConfig)
	}

	// successful config
	return cfg, nil
}

// Generate collector from config
func (cfg *configContent) generateCollector() *filesCollector {
	c := filesCollector{}
	for _, colCfg := range cfg.Exporter.Files {
		var col fileStatCollector
		col.filesPatterns = colCfg.GlobPatternPath
		if colCfg.EnableCRC32Metric != nil && *colCfg.EnableCRC32Metric {
			col.enableCRC32Metric = true
			c.atLeastOneCRC32Metric = true
		}
		if colCfg.EnableNbLineMetric != nil && *colCfg.EnableNbLineMetric {
			col.enableLineNbMetric = true
			c.atLeastOneLineNbMetric = true
		}
		c.collectors = append(c.collectors, col)
	}

	return &c
}
