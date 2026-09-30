// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of Mini Apps, which are this client's own.
func init() {
	for key, texts := range map[string][2]string{
		"miniapp.no_browser": {"Для Mini Apps нужен браузер на основе Chromium: укажите его в настройках, в разделе «Внешние интеграции».", "Mini Apps need a Chromium-based browser: choose one in the settings, under External integrations."},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
