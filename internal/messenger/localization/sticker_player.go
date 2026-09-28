// SPDX-License-Identifier: Unlicense OR MIT

package localization

func init() {
	for k, v := range map[string]string{
		"title": "Внутренний плеер WebM-стикеров",
		"wasm":  "WASM-песочница",
	} {
		russian["sticker_player."+k] = v
	}
	for k, v := range map[string]string{
		"title": "Internal WebM sticker player",
		"wasm":  "WASM sandbox",
	} {
		english["sticker_player."+k] = v
	}
}
