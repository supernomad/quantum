// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package version

import (
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {
	ver := Version()
	if ver == "" || ver != fmt.Sprintf("%d.%d.%d", major, minor, patch) {
		t.Fatal("Version either returned an empty string, or incorrect format.")
	}
}
