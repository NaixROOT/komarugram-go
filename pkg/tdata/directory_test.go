// SPDX-License-Identifier: Unlicense
package tdata

import (
	_ "embed"
	"os"
	"reflect"
	"testing"
)

func TestDirectory(t *testing.T) {

	dir, err := os.MkdirTemp("", "tdata_test_*")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)

	if err := WriteDir(dir, expected); err != nil {
		t.Error(err)
	}

	result, err := ReadDir(dir)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v to be equal to %+v", result, expected)
	}
}
