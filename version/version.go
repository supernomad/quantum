// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

// Package version exposes the current version of quantum in symantic format.
package version

import (
	"fmt"
)

const (
	major = 0
	minor = 14
	patch = 0
)

// Version returns the string form of the current quantum version.
func Version() string {
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
