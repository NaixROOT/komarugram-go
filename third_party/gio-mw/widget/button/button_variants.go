// SPDX-License-Identifier: Unlicense OR MIT

package button

func Elevated() *Button {
	return &Button{
		bKind: elevatedKind,
	}
}

func Filled() *Button {
	return &Button{
		bKind: filledKind,
	}
}

func FilledTonal() *Button {
	return &Button{
		bKind: filledTonalKind,
	}
}

func Outlined() *Button {
	return &Button{
		bKind: outlinedKind,
	}
}

func Text() *Button {
	return &Button{
		bKind: textKind,
	}
}
