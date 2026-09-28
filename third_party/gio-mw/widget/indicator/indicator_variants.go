// SPDX-License-Identifier: Unlicense OR MIT

package indicator

func Linear() *Indicator {
	return &Indicator{
		indeterminate: false,
		kind:          linearKind,
	}
}

func Circular() *Indicator {
	return &Indicator{
		indeterminate: false,
		kind:          circularKind,
	}
}
