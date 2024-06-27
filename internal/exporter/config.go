// Copyright 2019-2024 Michael DOUBEZ
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
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	yaml "gopkg.in/yaml.v2"
)

type configExporter struct {
	treeConfig `yaml:",inline"`

	WorkingDirectory string `yaml:"working_directory,omitempty"`

	ListenAddress string `yaml:"listen_address,omitempty"`
	MetricsPath   string `yaml:"metrics_path,omitempty"`

	Trees []*treeConfig `yaml:"trees"`
}

type configContent struct {
	Exporter configExporter `yaml:"exporter"`
}

var emptyTreeName = ""

func (cfg *configContent) readFile(cfgFile string, logger log.Logger) error {
	info, err := os.Stat(cfgFile)
	if err == nil && !info.IsDir() {
		level.Info(logger).Log("msg", "Reading config", "file", cfgFile)
		r, err := os.Open(cfgFile)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(r)
		decoder.SetStrict(true)
		err = decoder.Decode(cfg)
		if err != nil {
			return err
		}
	} else {
		level.Info(logger).Log("msg", "Could not read config", "file", cfgFile)
	}
	return nil
}

func readConfig(cfgFile string, defaultCollector *treeConfig, logger log.Logger) (*configContent, error) {
	cfg := &configContent{}

	// read file if possible
	if cfgFile != "none" {
		if err := cfg.readFile(cfgFile, logger); err != nil {
			return nil, err
		}
	}

	// merge default config
	if cfg.Exporter.TreeName == nil {
		if defaultCollector.TreeName != nil {
			level.Info(logger).Log("msg", "Config", "from", "parameter", "tree_name", *defaultCollector.TreeName)
		}
	} else {
		level.Info(logger).Log("msg", "Config", "from", "general", "tree_name", *cfg.Exporter.TreeName)
	}
	if defaultCollector.TreeRoot != nil {
		level.Info(logger).Log("msg", "Config", "from", "parameter", "tree_root", *defaultCollector.TreeRoot)
		cfg.Exporter.TreeRoot = defaultCollector.TreeRoot
	} else if len(*cfg.Exporter.TreeRoot) > 0 {
		level.Info(logger).Log("msg", "Config", "from", "general", "tree_root", *cfg.Exporter.TreeRoot)
	}
	if cfg.Exporter.EnableCRC32Metric == nil {
		level.Info(logger).Log("msg", "Config", "from", "parameter", "enable_harsh_crc32_metric", *defaultCollector.EnableCRC32Metric)
	} else {
		level.Info(logger).Log("msg", "Config", "from", "general", "enable_harsh_crc32_metric", *cfg.Exporter.EnableCRC32Metric)
	}
	if cfg.Exporter.EnableNbLineMetric == nil {
		level.Info(logger).Log("msg", "Config", "from", "parameter", "enable_nb_line_metric", *defaultCollector.EnableNbLineMetric)
	} else {
		level.Info(logger).Log("msg", "Config", "from", "general", "enable_nb_line_metric", *cfg.Exporter.EnableNbLineMetric)
	}
	mergeTreeConfig(&cfg.Exporter.treeConfig, defaultCollector)

	hasAtLeastOneTreeName := (cfg.Exporter.TreeName != nil)
	for _, tree := range cfg.Exporter.Trees {
		mergeTreeConfig(tree, &cfg.Exporter.treeConfig)
		if tree.TreeName != nil {
			hasAtLeastOneTreeName = true
		}
	}
	// set default tree name is not configured
	if hasAtLeastOneTreeName && cfg.Exporter.TreeName == nil {
		level.Info(logger).Log("msg", "Config", "from", "default", "tree_name", "<empty>")
		cfg.Exporter.TreeName = &emptyTreeName
		for _, tree := range cfg.Exporter.Trees {
			if tree.TreeName == nil {
				tree.TreeName = &emptyTreeName
			}
		}
	}

	// patterns from command line
	if len(defaultCollector.GlobPatternPath) != 0 {
		level.Info(logger).Log("msg", "Adding collection of patterns", "from", "command line")
		cfg.Exporter.Files = append(cfg.Exporter.Files, &defaultCollector.collectorConfig)
	}

	level.Debug(logger).Log("msg", "Success config", "content", cfg.toString())

	// successful config
	return cfg, nil
}

func (cfg *configContent) toString() string {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return "error"
	}

	return string(b)
}

// Generate collector from config
func (cfg *configContent) generateCollector(logger log.Logger) *filesCollector {
	c := createFilesCollector(logger, (cfg.Exporter.TreeName != nil))

	hasAtleastOneCRC32Metric := false
	hasAtleastOneLineNbMetric := false
	for _, colCfg := range cfg.Exporter.Files {
		col := cfg.Exporter.treeConfig.createFileStatCollector(colCfg)
		hasAtleastOneCRC32Metric = hasAtleastOneCRC32Metric || col.enableCRC32Metric
		hasAtleastOneLineNbMetric = hasAtleastOneLineNbMetric || col.enableLineNbMetric
		c.addFileStatCollector(cfg.Exporter.TreeName, col)
	}

	for _, tree := range cfg.Exporter.Trees {
		for _, colCfg := range tree.Files {
			col := tree.createFileStatCollector(colCfg)
			hasAtleastOneCRC32Metric = hasAtleastOneCRC32Metric || col.enableCRC32Metric
			hasAtleastOneLineNbMetric = hasAtleastOneLineNbMetric || col.enableLineNbMetric
			c.addFileStatCollector(tree.TreeName, col)
		}
	}

	if hasAtleastOneCRC32Metric {
		level.Debug(logger).Log("msg", "Collector creation", "has_at_least_a_crc32_metric", hasAtleastOneCRC32Metric)
		c.useFileCRC32Metric()
	}
	if hasAtleastOneLineNbMetric {
		level.Debug(logger).Log("msg", "Collector creation", "has_at_least_a_line_nb_metric", hasAtleastOneLineNbMetric)
		c.useLineNbMetric()
	}

	return c
}
