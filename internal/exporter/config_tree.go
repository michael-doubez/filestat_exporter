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
	"slices"
)

type collectorMetricConfig struct {
	EnableCRC32Metric  *bool `yaml:"enable_crc32_metric,omitempty"`
	EnableNbLineMetric *bool `yaml:"enable_nb_line_metric,omitempty"`
}

type collectorConfig struct {
	collectorMetricConfig `yaml:",inline"`

	GlobPatternPath []string `yaml:"patterns"`
}

type treeConfig struct {
	collectorConfig `yaml:",inline"`

	TreeName *string            `yaml:"tree_name,omitempty"`
	TreeRoot *string            `yaml:"tree_root,omitempty"`
	Files    []*collectorConfig `yaml:"files"`
}

func mergeTreeConfig(collectorTree *treeConfig, defaultTree *treeConfig) {
	mergeCollectorMetrics(&collectorTree.collectorMetricConfig, &defaultTree.collectorMetricConfig)
	if collectorTree.TreeName == nil && defaultTree.TreeName != nil {
		collectorTree.TreeName = defaultTree.TreeName
	}
	if collectorTree.TreeRoot == nil && defaultTree.TreeRoot != nil {
		collectorTree.TreeRoot = defaultTree.TreeRoot
	}

	for _, collector := range collectorTree.Files {
		mergeCollectorMetrics(&collector.collectorMetricConfig, &collectorTree.collectorMetricConfig)
	}
}

func mergeCollectorMetrics(collector *collectorMetricConfig, defaultCollector *collectorMetricConfig) {
	if collector.EnableCRC32Metric == nil {
		collector.EnableCRC32Metric = defaultCollector.EnableCRC32Metric
	}
	if collector.EnableNbLineMetric == nil {
		collector.EnableNbLineMetric = defaultCollector.EnableNbLineMetric
	}
}

func (tree *treeConfig) createFileStatCollector(colCfg *collectorConfig) fileStatCollector {
	col := fileStatCollector{}

	if tree.TreeName != nil {
		col.labels = []string{*tree.TreeName}
	}
	if tree.TreeRoot != nil {
		col.treeRoot = *tree.TreeRoot
	}
	col.filesPatterns = slices.Concat(colCfg.GlobPatternPath, tree.GlobPatternPath)

	col.enableCRC32Metric = colCfg.EnableCRC32Metric != nil && *colCfg.EnableCRC32Metric
	col.enableLineNbMetric = colCfg.EnableNbLineMetric != nil && *colCfg.EnableNbLineMetric

	return col
}
