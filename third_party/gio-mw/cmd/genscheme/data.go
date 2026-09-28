// SPDX-License-Identifier: Unlicense OR MIT

package main

type HexData struct {
	Description    string        `json:"description"`
	Seed           string        `json:"seed"`
	CoreColors     HexCoreColors `json:"coreColors"`
	ExtendedColors []any         `json:"extendedColors"`
	Schemes        HexSchemes    `json:"schemes"`
	Palettes       HexPalettes   `json:"palettes"`
}

type HexCoreColors struct {
	Primary        string `json:"primary"`
	Secondary      string `json:"secondary"`
	Tertiary       string `json:"tertiary"`
	Error          string `json:"error"`
	Neutral        string `json:"neutral"`
	NeutralVariant string `json:"neutralVariant"`
}

type HexScheme struct {
	Primary                 string `json:"primary"`
	SurfaceTint             string `json:"surfaceTint"`
	OnPrimary               string `json:"onPrimary"`
	PrimaryContainer        string `json:"primaryContainer"`
	OnPrimaryContainer      string `json:"onPrimaryContainer"`
	Secondary               string `json:"secondary"`
	OnSecondary             string `json:"onSecondary"`
	SecondaryContainer      string `json:"secondaryContainer"`
	OnSecondaryContainer    string `json:"onSecondaryContainer"`
	Tertiary                string `json:"tertiary"`
	OnTertiary              string `json:"onTertiary"`
	TertiaryContainer       string `json:"tertiaryContainer"`
	OnTertiaryContainer     string `json:"onTertiaryContainer"`
	Error                   string `json:"error"`
	OnError                 string `json:"onError"`
	ErrorContainer          string `json:"errorContainer"`
	OnErrorContainer        string `json:"onErrorContainer"`
	Background              string `json:"background"`
	OnBackground            string `json:"onBackground"`
	Surface                 string `json:"surface"`
	OnSurface               string `json:"onSurface"`
	SurfaceVariant          string `json:"surfaceVariant"`
	OnSurfaceVariant        string `json:"onSurfaceVariant"`
	Outline                 string `json:"outline"`
	OutlineVariant          string `json:"outlineVariant"`
	Shadow                  string `json:"shadow"`
	Scrim                   string `json:"scrim"`
	InverseSurface          string `json:"inverseSurface"`
	InverseOnSurface        string `json:"inverseOnSurface"`
	InversePrimary          string `json:"inversePrimary"`
	PrimaryFixed            string `json:"primaryFixed"`
	OnPrimaryFixed          string `json:"onPrimaryFixed"`
	PrimaryFixedDim         string `json:"primaryFixedDim"`
	OnPrimaryFixedVariant   string `json:"onPrimaryFixedVariant"`
	SecondaryFixed          string `json:"secondaryFixed"`
	OnSecondaryFixed        string `json:"onSecondaryFixed"`
	SecondaryFixedDim       string `json:"secondaryFixedDim"`
	OnSecondaryFixedVariant string `json:"onSecondaryFixedVariant"`
	TertiaryFixed           string `json:"tertiaryFixed"`
	OnTertiaryFixed         string `json:"onTertiaryFixed"`
	TertiaryFixedDim        string `json:"tertiaryFixedDim"`
	OnTertiaryFixedVariant  string `json:"onTertiaryFixedVariant"`
	SurfaceDim              string `json:"surfaceDim"`
	SurfaceBright           string `json:"surfaceBright"`
	SurfaceContainerLowest  string `json:"surfaceContainerLowest"`
	SurfaceContainerLow     string `json:"surfaceContainerLow"`
	SurfaceContainer        string `json:"surfaceContainer"`
	SurfaceContainerHigh    string `json:"surfaceContainerHigh"`
	SurfaceContainerHighest string `json:"surfaceContainerHighest"`
}

type HexSchemes struct {
	Light               HexScheme `json:"light"`
	LightMediumContrast HexScheme `json:"light-medium-contrast"`
	LightHighContrast   HexScheme `json:"light-high-contrast"`
	Dark                HexScheme `json:"dark"`
	DarkMediumContrast  HexScheme `json:"dark-medium-contrast"`
	DarkHighContrast    HexScheme `json:"dark-high-contrast"`
}

type HexPalette struct {
	Num0   string `json:"0"`
	Num5   string `json:"5"`
	Num10  string `json:"10"`
	Num15  string `json:"15"`
	Num20  string `json:"20"`
	Num25  string `json:"25"`
	Num30  string `json:"30"`
	Num35  string `json:"35"`
	Num40  string `json:"40"`
	Num50  string `json:"50"`
	Num60  string `json:"60"`
	Num70  string `json:"70"`
	Num80  string `json:"80"`
	Num90  string `json:"90"`
	Num95  string `json:"95"`
	Num98  string `json:"98"`
	Num99  string `json:"99"`
	Num100 string `json:"100"`
}

type HexPalettes struct {
	Primary        HexPalette `json:"primary"`
	Secondary      HexPalette `json:"secondary"`
	Tertiary       HexPalette `json:"tertiary"`
	Neutral        HexPalette `json:"neutral"`
	NeutralVariant HexPalette `json:"neutral-variant"`
}
