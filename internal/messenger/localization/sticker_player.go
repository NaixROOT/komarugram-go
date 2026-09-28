// SPDX-License-Identifier: Unlicense OR MIT

package localization

func init() {
	for k, v := range map[string]string{
		"title":     "Внутренний плеер WebM-стикеров",
		"wasm":      "WASM-песочница",
		"mp4_title": "Внутренний плеер MP4-анимаций",
		"mp4_hint":  "GIF и анимированные аватары. Декодер для песочницы загружается с GitHub при первом использовании.",
	} {
		russian["sticker_player."+k] = v
	}
	for k, v := range map[string]string{
		"title":     "Internal WebM sticker player",
		"wasm":      "WASM sandbox",
		"mp4_title": "Internal MP4 animation player",
		"mp4_hint":  "GIFs and animated avatars. The sandbox's decoder is downloaded from GitHub when first used.",
	} {
		english["sticker_player."+k] = v
	}
}
