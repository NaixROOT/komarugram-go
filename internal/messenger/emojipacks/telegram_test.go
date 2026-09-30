// SPDX-License-Identifier: Unlicense OR MIT

package emojipacks

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// miniList is a list in the format of Telegram Desktop's, small enough to
// follow: two categories that matter and six that are there to be eight.
const miniList = `"😀","☺️"
"👍","👍🏻","👍🏼","👍🏽","👍🏾","👍🏿"
"🤝","🤝🏻","🫱🏻‍🫲🏼","🫱🏻‍🫲🏽","🫱🏻‍🫲🏾","🫱🏻‍🫲🏿","🫱🏼‍🫲🏻","🤝🏼","🫱🏼‍🫲🏽","🫱🏼‍🫲🏾","🫱🏼‍🫲🏿","🫱🏽‍🫲🏻","🫱🏽‍🫲🏼","🤝🏽","🫱🏽‍🫲🏾","🫱🏽‍🫲🏿","🫱🏾‍🫲🏻","🫱🏾‍🫲🏼","🫱🏾‍🫲🏽","🤝🏾","🫱🏾‍🫲🏿","🫱🏿‍🫲🏻","🫱🏿‍🫲🏼","🫱🏿‍🫲🏽","🫱🏿‍🫲🏾","🤝🏿"

"🐶"

"🍏"

"⚽️"

"🚗"

"⌚️"

"❤️","😀"

"🏳️"

=========================================
"😀","☺️","👍","🤝"

=========================================
"👍"

"🤝"

=========================================
"🗨", "©️"
`

func TestTelegramOrder(t *testing.T) {
	order, err := TelegramOrder(miniList)
	if err != nil {
		t.Fatal(err)
	}
	var cells []string
	for _, sequences := range order {
		cells = append(cells, strings.Join(sequences, "="))
	}
	want := []string{"😀", "☺"}
	// An emoji with skin tones is followed by its five tones.
	want = append(want, "👍", "👍🏻", "👍🏼", "👍🏽", "👍🏾", "👍🏿")
	// Two people: the 25 pairs of tones, the pairs of one tone being the
	// emoji with that tone, which the pair spelled out is drawn as.
	want = append(want, "🤝")
	tones := []string{"🏻", "🏼", "🏽", "🏾", "🏿"}
	for _, first := range tones {
		for _, second := range tones {
			if first == second {
				want = append(want, "🤝"+first+"=🫱"+first+"‍🫲"+first)
			} else {
				want = append(want, "🫱"+first+"‍🫲"+second)
			}
		}
	}
	// The other categories, without U+FE0F, an emoji met again keeping the
	// cell it has; then the emoji outside the panel.
	want = append(want, "🐶", "🍏", "⚽", "🚗", "⌚", "❤", "🏳", "🗨", "©")
	if !reflect.DeepEqual(cells, want) {
		t.Errorf("order:\n%v\nwant:\n%v", cells, want)
	}
}

func TestTelegramOrderTurnsDownOtherFiles(t *testing.T) {
	for name, list := range map[string]string{
		"not the list":      "hello",
		"too few sections":  `"😀"` + "\n\n" + `"😀"`,
		"an unquoted entry": strings.Replace(miniList, `"🐶"`, `🐶`, 1),
		"a tone that is not an emoji's second character": strings.Replace(miniList, `"👍","👍🏻"`, `"👍","👍"`, 1),
		"a line of tones that is short":                  strings.Replace(miniList, `,"👍🏿"`, ``, 1),
	} {
		if _, err := TelegramOrder(list); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

// TestTelegramOrderOfTheRealList checks the order against the list and the
// sprites of Telegram Desktop's sources, when they are cloned beside the
// project: the emoji must fill the sprites to their last row.
func TestTelegramOrderOfTheRealList(t *testing.T) {
	root := filepath.Join("..", "..", "..", "tdesktop", "Telegram")
	list, err := os.ReadFile(filepath.Join(root, "lib_ui", "emoji.txt"))
	if err != nil {
		t.Skip("no clone of Telegram Desktop: ", err)
	}
	order, err := TelegramOrder(string(list))
	if err != nil {
		t.Fatal(err)
	}
	images, _ := filepath.Glob(filepath.Join(root, "Resources", "emoji", "emoji_*.webp"))
	if len(images) == 0 {
		t.Skip("no sprites in the clone")
	}
	catalog := t.TempDir()
	p, err := BuildSprites(catalog, Pack{ID: "test", Name: "Test"}, TelegramLayout, images, order)
	if err != nil {
		t.Fatal(err)
	}
	set, err := OpenSprites(filepath.Join(catalog, "test"), p)
	if err != nil {
		t.Fatal(err)
	}
	cells := 0
	for _, rows := range set.rows {
		cells += rows * TelegramLayout.Columns
	}
	if n := set.Count(); n > cells || n <= cells-TelegramLayout.Columns {
		t.Errorf("%d emoji for %d cells: they do not end in the last row", n, cells)
	}
	t.Logf("%d emoji in %d cells of %d sprites", set.Count(), cells, len(images))
	if order[0][0] != "😀" {
		t.Errorf("the first cell is %q", order[0][0])
	}
}
