// SPDX-License-Identifier: Unlicense OR MIT

// Package assets holds the files of this directory the client builds in.
package assets

import _ "embed"

// LogoRound is logo_round.png, the application's icon.
//
//go:embed logo_round.png
var LogoRound []byte
