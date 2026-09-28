// SPDX-License-Identifier: Unlicense

package tdata

import (
	"reflect"
	"testing"
)

func TestZip(t *testing.T) {
	data, err := CreateZipBytes(expected)
	if err != nil {
		t.Error(err)
	}
	result, err := ReadZipBytes(data)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v to be equal to %+v", result, expected)
	}
}
