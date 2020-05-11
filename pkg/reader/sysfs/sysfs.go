/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package sysfs

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

type CoreIdList []int

type PCIResourceList []string

type Node struct {
	Id           int
	Cores        CoreIdList
	PCIResources PCIResourceList
}

type Topology struct {
	NUMANodeCount int
	NUMANodes     []Node
}

func (tp Topology) CoreIdToNUMANodeIdMap() map[int]int {
	// CPU IDs -> NUMA Node ID
	CPUToNUMANode := make(map[int]int)
	for _, node := range tp.NUMANodes {
		for _, coreId := range node.Cores {
			CPUToNUMANode[coreId] = node.Id
		}
	}
	return CPUToNUMANode
}

type SysFS string

func parseSysfsNodeOnline(data string) (int, error) {
	/*
	    The file content is expected to be:
	   "0\n" in one-node case
	   "0-K\n" in N-node case where K=N-1
	*/
	info := strings.TrimSpace(data)
	pair := strings.SplitN(info, "-", 2)
	if len(pair) != 2 {
		return 1, nil
	}
	out, err := strconv.Atoi(pair[1])
	if err != nil {
		return 0, err
	}
	return out + 1, nil
}

func (sf SysFS) String() string {
	return string(sf)
}

func (sf SysFS) NUMANodeCount() (int, error) {
	data, err := ioutil.ReadFile(path.Join(string(sf), "devices", "system", "node", "online"))
	if err != nil {
		return 0, err
	}
	nodeNum, err := parseSysfsNodeOnline(string(data))
	if err != nil {
		return 0, err
	}
	return nodeNum, nil
}

func (sf SysFS) CoresPerNUMANode(nodeNum int) (CoreIdList, bool, error) {
	data, err := ioutil.ReadFile(path.Join(string(sf), "devices", "system", "node", fmt.Sprintf("node%d", nodeNum), "cpulist"))
	if err != nil {
		return nil, false, err
	}
	cpus, err := cpuset.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, true, err
	}
	return CoreIdList(cpus.ToSlice()), true, nil
}

func (sf SysFS) PCIDeviceNUMANode(pciDevice string) (int, error) {
	data, err := ioutil.ReadFile(path.Join(string(sf), "bus", "pci", "devices", pciDevice, "numa_node"))
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func numaNodePCIDeviceMap(sf SysFS, pciDevices []string) (map[int][]string, error) {
	numaToDevs := make(map[int][]string)
	for _, pciDev := range pciDevices {
		nodeNum, err := sf.PCIDeviceNUMANode(pciDev)
		if err != nil {
			return numaToDevs, err
		}

		numaToDevs[nodeNum] = append(numaToDevs[nodeNum], pciDev)
	}
	return numaToDevs, nil
}

func NewTopologyFromSysFs(sysFsPath string, pciResources map[string]string) (Topology, error) {
	var err error
	tp := Topology{}
	sf := SysFS(sysFsPath)

	tp.NUMANodeCount, err = sf.NUMANodeCount()
	if err != nil {
		return tp, err
	}

	var pciDevices []string
	for pciAddr, _ := range pciResources {
		pciDevices = append(pciDevices, pciAddr)
	}

	numaToDevs, err := numaNodePCIDeviceMap(sf, pciDevices)
	if err != nil {
		return tp, err
	}

	for nodeId := 0; nodeId < tp.NUMANodeCount; nodeId++ {
		cores, found, err := sf.CoresPerNUMANode(nodeId)
		if found && err != nil {
			return tp, err
		}
		// else node offline (TODO: validate: can Linux put a node offline?)

		var pciResourceList []string
		for _, pciDev := range numaToDevs[nodeId] {
			pciResourceList = append(pciResourceList, pciResources[pciDev])
		}
		tp.NUMANodes = append(tp.NUMANodes, Node{
			Id:           nodeId,
			Cores:        cores,
			PCIResources: pciResourceList,
		})
	}

	return tp, nil
}
