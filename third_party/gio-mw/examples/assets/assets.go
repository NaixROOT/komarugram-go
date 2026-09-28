// SPDX-License-Identifier: Unlicense OR MIT

package assets

import (
	"embed"
)

//go:embed *.png
var exampleAssets embed.FS

func GetExampleAssets() *embed.FS {
	return &exampleAssets
}
