// SPDX-License-Identifier: Unlicense OR MIT

package model

import "testing"

func TestInlineQuery(t *testing.T) {
	for text, want := range map[string][2]string{"@gif cats": {"gif", "cats"}, "@vid ": {"vid", ""}, "@wiki two words": {"wiki", "two words"}} {
		if u, q, ok := InlineQuery(text); !ok || u != want[0] || q != want[1] {
			t.Errorf("%q: %q %q %v", text, u, q, ok)
		}
	}
	for _, text := range []string{"@gif", "hi @gif cats", "@g cats", "@ cats", "/start"} {
		if _, _, ok := InlineQuery(text); ok {
			t.Errorf("%q is a query", text)
		}
	}
	if ItemSendKind(PickerItem{ResultID: "1"}) != SendInline || ItemSendKind(PickerItem{ResultID: "1", Media: Message{Kind: MessageGIF}}) != SendGIF {
		t.Error("inline results need the wrong right")
	}
}
