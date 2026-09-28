// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"fmt"
	"gio-mw/token"
	"strings"

	"gioui.org/layout"
	"gioui.org/text"
)

const ThemeNamespace = "gio-mw.wdk.Theme"

func InitMaterialThemeInContext(gtx layout.Context, theme *token.Theme) {
	gtx.Values[ThemeNamespace] = theme
}

func GetMaterialTheme(gtx layout.Context) *token.Theme {
	if v, ok := gtx.Values[ThemeNamespace]; ok {
		if typed, ok := v.(*token.Theme); ok {
			return typed
		}
		panic(fmt.Errorf("%s is not of type *token.Theme", ThemeNamespace))
	}
	panic(fmt.Errorf("%s not set", ThemeNamespace))
}

func GetTextShaper(gtx layout.Context) *text.Shaper {
	theme := GetMaterialTheme(gtx)
	return theme.TextShaper
}

func InitWidgetThemeInTheme[T any](materialTheme *token.Theme, key string, v *T) {
	materialTheme.Widgets[key] = v
}

func GetWidgetTheme[T any](materialTheme *token.Theme, namespace string) *T {
	if v, ok := materialTheme.Widgets[namespace]; ok {
		if typed, ok := v.(*T); ok {
			return typed
		}
		panic(fmt.Errorf("%s is not of type *%T", namespace, new(T)))
	}
	return nil
}

func ClearWidgetTheme(materialTheme *token.Theme, namespace string) {
	delete(materialTheme.Widgets, namespace)
}

func ClearWidgetThemesByPrefix(materialTheme *token.Theme, prefix string) {
	for k := range materialTheme.Widgets {
		if strings.HasPrefix(k, prefix) {
			delete(materialTheme.Widgets, k)
		}
	}
}
