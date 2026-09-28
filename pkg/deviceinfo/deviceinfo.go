// SPDX-License-Identifier: Unlicense OR MIT

// Package deviceinfo names the computer and its operating system the way
// Telegram Desktop does when it connects (lib_base's
// base/platform/base_platform_info.cpp and base/platform/*/base_info_*).
// Telegram shows both in the account's list of sessions.
//
// Windows and Linux are ported. Elsewhere the model is "Desktop" ("Mac" on
// macOS) and the system is left to the caller.
package deviceinfo

import (
	"runtime"
	"slices"
	"strings"
	"sync"
	"unicode/utf16"
)

// Model is the name of the computer (DeviceModelPretty): the product name
// the firmware gives when it is short enough, else its family and board,
// else "Desktop".
func Model() string { return modelOnce() }

// System is the operating system with what it runs on
// (SystemVersionPretty), such as "Windows 11 x64" or
// "Linux XFCE X11 glibc 2.39"; "" where it is not ported.
func System() string { return systemOnce() }

var (
	modelOnce  = sync.OnceValue(func() string { return finalizeModel(model(), runtime.GOOS == "darwin") })
	systemOnce = sync.OnceValue(system)
)

const (
	// maxModelLength is the longest name from the firmware that is used as
	// it is (kMaxDeviceModelLength).
	maxModelLength = 15
	// maxGoodModelLength bounds the long product names that are shortened
	// instead (kMaxGoodDeviceModelLength).
	maxGoodModelLength = 32
)

// qtLength is the length of s as Qt counts it, in UTF-16 code units.
func qtLength(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}

// modelOK reports whether a name from the firmware is short enough to be the
// model (IsDeviceModelOk).
func modelOK(model string) bool {
	return model != "" && qtLength(model) <= maxModelLength
}

// cleanAndSimplify turns control characters into spaces, trims the ends and
// collapses runs of white space into one space (base::CleanAndSimplify).
func cleanAndSimplify(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 32 {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

// simplifyModel is a name from the firmware without underscores, simplified
// (SimplifyDeviceModel).
func simplifyModel(s string) string {
	return cleanAndSimplify(strings.ReplaceAll(s, "_", ""))
}

// simplifyGoodModel drops the words of model that are in remove, compared in
// lower case, and stops before the word that would make it longer than
// maxGoodModelLength (SimplifyGoodDeviceModel).
func simplifyGoodModel(model string, remove ...string) string {
	result := ""
	for _, word := range strings.Split(model, " ") {
		switch {
		case slices.Contains(remove, strings.ToLower(word)):
		case result == "":
			result = word
		case qtLength(result)+qtLength(word)+1 > maxGoodModelLength:
			return result
		default:
			result += " " + word
		}
	}
	return result
}

// productModel is the product name when it names the computer by itself
// (ProductNameToDeviceModel): short ones as they are, HP's long ones
// without their generic words.
func productModel(product string) string {
	switch {
	case strings.HasPrefix(product, "HP "):
		return simplifyGoodModel(product, "notebook", "desktop", "mobile", "workstation", "pc")
	case modelOK(product):
		return product
	}
	return ""
}

// firmwareModel picks the model from the product name, the family and the
// board the firmware gives, each already simplified, as DeviceModelPretty
// does on Windows and Linux; "" when none of them will do.
func firmwareModel(product, family, board string) string {
	if m := productModel(product); m != "" {
		return m
	}
	familyBoard := simplifyModel(family + " " + board)
	switch {
	case modelOK(familyBoard):
		return familyBoard
	case modelOK(board):
		return board
	case modelOK(family):
		return family
	}
	return ""
}

// finalizeModel trims model and names an unknown computer "Desktop", or
// "Mac" on a Mac (FinalizeDeviceModel).
func finalizeModel(model string, mac bool) string {
	model = strings.TrimSpace(model)
	switch {
	case model != "":
		return model
	case mac:
		return "Mac"
	}
	return "Desktop"
}
