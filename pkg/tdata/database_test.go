// SPDX-License-Identifier: Unlicense

package tdata

import (
	"reflect"
	"testing"
)

var inputTelethon = []TDataSession{
	{
		AuthKey: "0c3de273f089c3e485231350cd5578e657070669a53fd8b79a17b95e8005d6b279ec294403abb8f7d263009734d8a8c13af142da59cfcd3cc4e714332b550d461f0f622dadb675360269ca10591960a5fa35b6530c82cfd0c4b4204f5616c90088507ba24487391377fe752139f10e23392fe73a98c184a19f59608f2cfec87c9a6ff93f46ca1ba8d9c44982d33a1974ad0059ae18244de6e935e1c7a23dad8cfb7d855c207439fbdfb547e7d3f3c5f3ad4bca7b4ba02504c0be49f4e1bb7ccb5882da60959b33be2334be483ffeda4caa0293e85267ebc6405de8f3ec3f93e4be3e3b300d9775e05e16d137d0bf8c3b2d6bf34e6ce1b65bb950f9856c52b02d",
		DC:      2,
	},
}

var inputPyrogram = []TDataSession{
	{
		AuthKey: "0c3de273f089c3e485231350cd5578e657070669a53fd8b79a17b95e8005d6b279ec294403abb8f7d263009734d8a8c13af142da59cfcd3cc4e714332b550d461f0f622dadb675360269ca10591960a5fa35b6530c82cfd0c4b4204f5616c90088507ba24487391377fe752139f10e23392fe73a98c184a19f59608f2cfec87c9a6ff93f46ca1ba8d9c44982d33a1974ad0059ae18244de6e935e1c7a23dad8cfb7d855c207439fbdfb547e7d3f3c5f3ad4bca7b4ba02504c0be49f4e1bb7ccb5882da60959b33be2334be483ffeda4caa0293e85267ebc6405de8f3ec3f93e4be3e3b300d9775e05e16d137d0bf8c3b2d6bf34e6ce1b65bb950f9856c52b02d",
		DC:      2,
		UserID:  1337,
	},
}

func TestTelethon(t *testing.T) {
	input := inputTelethon
	data, err := CreateTelethonDatabaseBytes(input)
	if err != nil {
		t.Error(err)
	}
	result, err := ReadDatabaseBytes(data)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, input) {
		t.Errorf("expected %+v to be equal to %+v", result, input)
	}
}

func TestPyrogram(t *testing.T) {
	input := inputPyrogram
	data, err := CreatePyrogramDatabaseBytes(input)
	if err != nil {
		t.Error(err)
	}
	result, err := ReadDatabaseBytes(data)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, input) {
		t.Errorf("expected %+v to be equal to %+v", result, input)
	}
}
