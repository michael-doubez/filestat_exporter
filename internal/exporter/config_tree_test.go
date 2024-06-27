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
	"testing"
)

func TestMergeTreeConfig_ShouldSetDefaultWhenNotDefined(t *testing.T) {
	defaultName, defaultRoot := "", "a/path"
	defaultTree := treeConfig{TreeName: &defaultName, TreeRoot: &defaultRoot}
	collectorTree := treeConfig{TreeName: nil, TreeRoot: nil}

	mergeTreeConfig(&collectorTree, &defaultTree)

	if collectorTree.TreeName != defaultTree.TreeName {
		t.Error("TreeName not set from default")
	}
	if collectorTree.TreeRoot != defaultTree.TreeRoot {
		t.Error("TreeRoot not set from default")
	}
}

func TestMergeTreeConfig_ShouldNotSetDefaultWhenAlreadyDefined(t *testing.T) {
	defaultName, defaultRoot := "", "a/path"
	treeName, treeRoot := "name", "b/othe"
	defaultTree := treeConfig{TreeName: &defaultName, TreeRoot: &defaultRoot}
	collectorTree := treeConfig{TreeName: &treeName, TreeRoot: &treeRoot}

	mergeTreeConfig(&collectorTree, &defaultTree)

	if collectorTree.TreeName == defaultTree.TreeName {
		t.Error("TreeName set from default")
	}
	if collectorTree.TreeRoot == defaultTree.TreeRoot {
		t.Error("TreeRoot set from default")
	}
}

func TestMergeTreeConfig_ShouldNotSetTreeWhenNeverDefined(t *testing.T) {
	defaultTree := treeConfig{TreeName: nil, TreeRoot: nil}
	collectorTree := treeConfig{TreeName: nil, TreeRoot: nil}

	mergeTreeConfig(&collectorTree, &defaultTree)

	if collectorTree.TreeName != nil {
		t.Error("TreeName set from ?")
	}
	if collectorTree.TreeRoot != nil {
		t.Error("TreeRoot set from ?")
	}
}

func TestMergeCollectorMetrics_ShouldSetFromDefaultWhenNil(t *testing.T) {
	dh, dl := true, false
	defaultCollector := collectorMetricConfig{EnableCRC32Metric: &dh, EnableNbLineMetric: &dl}
	collector := collectorMetricConfig{}

	mergeCollectorMetrics(&collector, &defaultCollector)

	if collector.EnableCRC32Metric != defaultCollector.EnableCRC32Metric {
		t.Error("EnableCRC32Metric not set from default")
	}
	if collector.EnableNbLineMetric != defaultCollector.EnableNbLineMetric {
		t.Error("EnableNbLineMetric not set from default")
	}
}

func TestMergeCollectorMetrics_ShouldNotSetWhenAlreadySet(t *testing.T) {
	dh, dl := true, false
	h, l := false, true
	defaultCollector := collectorMetricConfig{EnableCRC32Metric: &dh, EnableNbLineMetric: &dl}
	collector := collectorMetricConfig{EnableCRC32Metric: &h, EnableNbLineMetric: &l}

	mergeCollectorMetrics(&collector, &defaultCollector)

	if collector.EnableCRC32Metric == defaultCollector.EnableCRC32Metric {
		t.Error("EnableCRC32Metric set from default")
	}
	if collector.EnableNbLineMetric == defaultCollector.EnableNbLineMetric {
		t.Error("EnableNbLineMetric set from default")
	}
}

func TestMergeTreeConfig_ShouldSetCollectorMetricsFromDefaultAndTree(t *testing.T) {
	dh, dl := true, false
	defaultTree := treeConfig{
		collectorConfig: collectorConfig{
			collectorMetricConfig: collectorMetricConfig{EnableCRC32Metric: &dh, EnableNbLineMetric: nil},
		},
	}
	collector := collectorConfig{}
	collectorTree := treeConfig{
		collectorConfig: collectorConfig{
			collectorMetricConfig: collectorMetricConfig{EnableCRC32Metric: nil, EnableNbLineMetric: &dl},
		},
		Files: []*collectorConfig{&collector},
	}

	mergeTreeConfig(&collectorTree, &defaultTree)

	if collector.EnableCRC32Metric != &dh {
		t.Error("EnableCRC32Metric not sert from default")
	}
	if collector.EnableNbLineMetric != &dl {
		t.Error("EnableNbLineMetric not set from tree")
	}
}
