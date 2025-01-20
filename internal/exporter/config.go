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
	"log/slog"
	"os"

	yaml "gopkg.in/yaml.v3"
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

func (cfg *configContent) readFile(cfgFile string, logger slog.Logger) error {
	info, err := os.Stat(cfgFile)
	if err == nil && !info.IsDir() {
		logger.Info("Reading config", "file", cfgFile)
		r, err := os.Open(cfgFile)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(r)
		decoder.KnownFields(true)
		err = decoder.Decode(cfg)
		if err != nil {
			return err
		}
	} else {
		logger.Info("Could not read config", "file", cfgFile)
	}
	return nil
}

func readConfig(cfgFile string, defaultCollector *treeConfig, logger slog.Logger) (*configContent, error) {
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
			logger.Info("Config", "from", "parameter", "tree_name", *defaultCollector.TreeName)
		}
	} else {
		logger.Info("Config", "from", "general", "tree_name", *cfg.Exporter.TreeName)
	}
	if defaultCollector.TreeRoot != nil {
		logger.Info("Config", "from", "parameter", "tree_root", *defaultCollector.TreeRoot)
		cfg.Exporter.TreeRoot = defaultCollector.TreeRoot
	} else if len(*cfg.Exporter.TreeRoot) > 0 {
		logger.Info("Config", "from", "general", "tree_root", *cfg.Exporter.TreeRoot)
	}
	if cfg.Exporter.EnableCRC32Metric == nil {
		logger.Info("Config", "from", "parameter", "enable_harsh_crc32_metric", *defaultCollector.EnableCRC32Metric)
	} else {
		logger.Info("Config", "from", "general", "enable_harsh_crc32_metric", *cfg.Exporter.EnableCRC32Metric)
	}
	if cfg.Exporter.EnableNbLineMetric == nil {
		logger.Info("Config", "from", "parameter", "enable_nb_line_metric", *defaultCollector.EnableNbLineMetric)
	} else {
		logger.Info("Config", "from", "general", "enable_nb_line_metric", *cfg.Exporter.EnableNbLineMetric)
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
		logger.Info("Config", "from", "default", "tree_name", "<empty>")
		cfg.Exporter.TreeName = &emptyTreeName
		for _, tree := range cfg.Exporter.Trees {
			if tree.TreeName == nil {
				tree.TreeName = &emptyTreeName
			}
		}
	}

	// patterns from command line
	if len(defaultCollector.GlobPatternPath) != 0 {
		logger.Info("Adding collection of patterns", "from", "command line")
		cfg.Exporter.Files = append(cfg.Exporter.Files, &defaultCollector.collectorConfig)
	}

	logger.Debug("Success config", "content", cfg.toString())

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
func (cfg *configContent) generateCollector(logger slog.Logger) *filesCollector {
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
		logger.Debug("Collector creation", "has_at_least_a_crc32_metric", hasAtleastOneCRC32Metric)
		c.useFileCRC32Metric()
	}
	if hasAtleastOneLineNbMetric {
		logger.Debug("Collector creation", "has_at_least_a_line_nb_metric", hasAtleastOneLineNbMetric)
		c.useLineNbMetric()
	}

	return c
}
