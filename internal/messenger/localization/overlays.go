// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of the settings of the overlays (the floating composer, the context
// menus and the toasts), which are this client's own.
func init() {
	for key, texts := range map[string][2]string{
		"settings.overlays":                   {"Оверлеи", "Overlays"},
		"settings.overlays_transparency":      {"Прозрачность оверлеев", "Overlay transparency"},
		"settings.overlays_transparency_hint": {"Панели, под которыми включено размытие, пропускают столько размытого фона; остальные непрозрачны.", "Overlays that blur what is behind them let this much of it show through; the others are opaque."},
		"settings.menus_blur":                 {"Размытие под контекстными меню", "Blur behind the context menus"},
		"settings.toasts_blur":                {"Размытие под уведомлениями", "Blur behind the toasts"},
		"settings.overlays_classic":           {"У классической панели ввода нет фона под собой: её размытие включается вместе с плавающей.", "The classic composer has no content behind it: its blur applies once it floats."},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
