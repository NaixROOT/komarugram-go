package model

import (
	"strings"
	"testing"
)

func TestOutgoingValidation(t *testing.T) {
	for _, m := range []OutgoingMessage{{RandomID: 1}, {RandomID: 1, Text: strings.Repeat("😀", 2049)}, {RandomID: 1, Text: "Tasks", Tasks: []string{""}}} {
		if m.Validate() == nil {
			t.Fatalf("accepted invalid draft: %+v", m)
		}
	}
	if err := (OutgoingMessage{RandomID: 1, Text: "hello"}).Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestSavedSearchPriority(t *testing.T) {
	got := PreferSaved([]PickerItem{{ID: "saved"}, {ID: "shared"}}, []PickerItem{{ID: "global"}, {ID: "shared"}})
	if len(got) != 3 || got[0].ID != "saved" || got[1].ID != "shared" || got[2].ID != "global" {
		t.Fatalf("wrong order: %+v", got)
	}
}
