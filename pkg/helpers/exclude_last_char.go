// SPDX-License-Identifier: Unlicense

package helpers

func ExcludeLastChar(str string) string {
	length := len(str)
	if length > 0 {
		return str[:length-1]
	}
	return str
}
