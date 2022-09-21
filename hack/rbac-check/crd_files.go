// Copyright 2022 TriggerMesh Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"path"
)

// Prefixes for CRD manifest files.
const (
	filePrefixLen = 4

	crdPrefixEventing = "300-"
)

// crdFilenameToResource extracts the singular resource name from the given
// file name.
// The given file name is expected to be valid (for example using isCRDFile).
func crdFilenameToResource(name string) string {
	return name[filePrefixLen : len(name)-len(path.Ext(name))]
}

// isCRDFile asserts that the given file name corresponds to
// a manifest file for a CRD.
func isCRDFile(name string) bool {
	const expectPattern = "3xx-*.yaml"
	if len(name) < len(expectPattern) {
		return false
	}

	// starts with /3[0-9]{2}/
	if name[0] != '3' || !isDigit(name[1]) || !isDigit(name[2]) {
		return false
	}

	// ends with '.yaml'
	return path.Ext(name) == ".yaml"
}

// isDigit reports whether the given char is a digit in the
// latin unicode character range.
func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
