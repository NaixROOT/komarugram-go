// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"os"
	"testing"

	"komarugram/internal/messenger/preferences"
)

// TestMain gives the themes of the tests the fonts the environment names
// (KOMARUGRAM_FONT and the others of internal/messenger/fonts), so that a
// render test shows a font without the settings.
func TestMain(m *testing.M) {
	applyFonts(preferences.Fonts{}, nil)
	os.Exit(m.Run())
}
