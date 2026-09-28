// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"errors"
	"strings"
	"testing"
)

func TestProfileEditValidate(t *testing.T) {
	ok := ProfileEdit{FirstName: "Ада", Username: "ada_lovelace", Bio: "Математик"}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		edit ProfileEdit
		want error
	}{
		"no first name":       {ProfileEdit{FirstName: "  "}, ErrNameInvalid},
		"long last name":      {ProfileEdit{FirstName: "A", LastName: strings.Repeat("я", MaxNameLength+1)}, ErrNameInvalid},
		"long bio":            {ProfileEdit{FirstName: "A", Bio: strings.Repeat("я", MaxBioLength+1)}, ErrBioTooLong},
		"short username":      {ProfileEdit{FirstName: "A", Username: "ada"}, ErrUsernameInvalid},
		"digit first":         {ProfileEdit{FirstName: "A", Username: "1adaaa"}, ErrUsernameInvalid},
		"trailing _":          {ProfileEdit{FirstName: "A", Username: "adaaa_"}, ErrUsernameInvalid},
		"not latin":           {ProfileEdit{FirstName: "A", Username: "адааааа"}, ErrUsernameInvalid},
		"no username is fine": {ProfileEdit{FirstName: "A"}, nil},
	} {
		if err := c.edit.Validate(); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, want %v", name, err, c.want)
		}
	}
}
