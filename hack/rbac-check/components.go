// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import "fmt"

// TriggerMesh's API groups.
const (
	apiGroupEventing = "eventing.triggermesh.io"
)

// componentsDictionary is a dictionary of component names indexed by API group.
type componentsDictionary map[ /*API group*/ string][]string

// AddSource adds a resource to the dictionary's "eventing" apiGroup.
func (d componentsDictionary) AddSource(resource string) {
	d[apiGroupEventing] = append(d[apiGroupEventing], resource)
}

// readComponents returns a componentsDictionary populated with the resources
// discovered in the given config directory.
func readComponents(dir string) (componentsDictionary, error) {
	dirEntries, err := filesystem.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	components := make(componentsDictionary)

	for _, e := range dirEntries {
		fileName := e.Name()

		// ignore non-regular files (e.g. symlinks, directories)
		// ignore non-CRD manifests (3xx-*.yaml)
		if !e.Type().IsRegular() || !isCRDFile(fileName) {
			continue
		}

		resource := crdFilenameToResource(fileName)

		switch prefix := fileName[:filePrefixLen]; prefix {
		case crdPrefixSources:
			components.AddSource(resource)
		case crdPrefixTargets:
			components.AddTarget(resource)
		case crdPrefixRouting:
			components.AddRouter(resource)
		case crdPrefixExtensions:
			components.AddExtension(resource)
		case crdPrefixFlow:
			components.AddFlow(resource)
		default:
			// This shouldn't happen.
			// Fail loudly if we find a file with an unknown prefix.
			panic(fmt.Errorf("undefined file prefix %q in file name %s", prefix, fileName))
		}
	}

	return components, nil
}
