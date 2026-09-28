// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"testing"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

func TestPremiumFromServerConfig(t *testing.T) {
	config := &tg.JSONObject{Value: []tg.JSONObjectValue{
		{Key: "about_length_limit_default", Value: &tg.JSONNumber{Value: 80}},
		{Key: "about_length_limit_premium", Value: &tg.JSONNumber{Value: 160}},
		{Key: "premium_purchase_blocked", Value: &tg.JSONBool{Value: false}},
		{Key: "premium_bot_username", Value: &tg.JSONString{Value: "PremiumBot"}},
		{Key: "emojies_send_dice", Value: &tg.JSONArray{Value: []tg.JSONValueClass{&tg.JSONString{Value: "🎲"}}}},
	}}
	values, ok := jsonValue(config).(map[string]any)
	if !ok {
		t.Fatal("the configuration is not an object")
	}
	p := model.PremiumFromConfig(false, values)
	if p.Limit("about_length_limit") != 80 || !p.Purchasable || p.Bot != "PremiumBot" {
		t.Fatalf("without Premium: %+v", p)
	}
	p.Active = true
	if p.Limit("about_length_limit") != 160 {
		t.Fatalf("with Premium the bio limit is %d", p.Limit("about_length_limit"))
	}
	// A key the server did not send keeps Telegram Desktop's value.
	if p.Limit("dialog_filters_limit") != 30 {
		t.Fatalf("fallback folder limit %d", p.Limit("dialog_filters_limit"))
	}
}

func TestPurchaseBlockedUnlessConfigured(t *testing.T) {
	if model.PremiumFromConfig(false, nil).Purchasable {
		t.Fatal("Premium offered without the server allowing it")
	}
}

func TestEmojiStatusExpires(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	if got := emojiStatus(&tg.EmojiStatus{DocumentID: 7}, now); got != 7 {
		t.Errorf("permanent status %d", got)
	}
	if got := emojiStatus(&tg.EmojiStatus{DocumentID: 7, Until: int(now.Add(-time.Minute).Unix())}, now); got != 0 {
		t.Errorf("expired status %d", got)
	}
	if got := emojiStatus(&tg.EmojiStatusCollectible{DocumentID: 9}, now); got != 9 {
		t.Errorf("collectible status %d", got)
	}
	if got := emojiStatus(&tg.EmojiStatusEmpty{}, now); got != 0 {
		t.Errorf("empty status %d", got)
	}
}
