// SPDX-License-Identifier: Unlicense OR MIT

package model

// PremiumLimit is a limit that Telegram Premium raises. The server sets both
// values in the client configuration (help.getAppConfig) as Key+"_default"
// and Key+"_premium".
type PremiumLimit struct {
	Key              string
	Default, Premium int
}

// PremiumLimits are the limits the client knows, in the order the Premium
// page lists them, with the values Telegram Desktop falls back on when the
// server's configuration does not name them.
var PremiumLimits = []PremiumLimit{
	{"channels_limit", 500, 1000},
	{"channels_public_limit", 10, 20},
	{"dialog_filters_limit", 10, 30},
	{"dialog_filters_chats_limit", 100, 200},
	{"dialogs_pinned_limit", 5, 10},
	{"dialogs_folder_pinned_limit", 100, 200},
	{"saved_dialogs_pinned_limit", 5, 100},
	{"chatlists_joined_limit", 2, 20},
	{"saved_gifs_limit", 200, 400},
	{"stickers_faved_limit", 5, 10},
	{"about_length_limit", 70, 140},
	{"caption_length_limit", 1024, 2048},
	{"message_length_limit", 4096, 8192},
	{"upload_max_fileparts", 4000, 8000},
}

// UploadPartSize is the size of one part of an upload, which
// upload_max_fileparts counts: 4000 parts are 2 GB.
const UploadPartSize = 512 << 10

// Premium is what Telegram Premium means for the account.
type Premium struct {
	// Active is set when the account has Premium.
	Active bool
	// Purchasable is set when the server lets this client offer Premium;
	// Telegram Desktop offers it only if the configuration says so.
	Purchasable bool
	// Bot sells Premium, by username; empty when the server names none.
	Bot string
	// Limits are PremiumLimits with the server's values.
	Limits []PremiumLimit
}

// Limit returns the account's value of the limit key, with or without
// Premium as it has it; 0 for a key that is not known.
func (p Premium) Limit(key string) int {
	limits := p.Limits
	if limits == nil {
		limits = PremiumLimits
	}
	for _, l := range limits {
		if l.Key == key {
			if p.Active {
				return l.Premium
			}
			return l.Default
		}
	}
	return 0
}

// PremiumSource is a Store that knows the account's Premium.
type PremiumSource interface {
	Premium() Premium
}

// PremiumFromConfig reads the client configuration, flattened to keys and
// numbers, bools and strings, into Premium. Keys the server does not send
// keep their fallback values.
func PremiumFromConfig(active bool, config map[string]any) Premium {
	p := Premium{Active: active, Limits: make([]PremiumLimit, len(PremiumLimits))}
	copy(p.Limits, PremiumLimits)
	for i, l := range p.Limits {
		if v, ok := config[l.Key+"_default"].(float64); ok && v > 0 {
			p.Limits[i].Default = int(v)
		}
		if v, ok := config[l.Key+"_premium"].(float64); ok && v > 0 {
			p.Limits[i].Premium = int(v)
		}
	}
	if blocked, ok := config["premium_purchase_blocked"].(bool); ok {
		p.Purchasable = !blocked
	}
	p.Bot, _ = config["premium_bot_username"].(string)
	return p
}
