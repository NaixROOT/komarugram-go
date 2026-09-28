// SPDX-License-Identifier: Unlicense OR MIT

package account

import (
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// Identity is who the client says it is when it connects: the application's
// api_id and api_hash, and the device, system and version reported in
// initConnection. The server ties an authorization to the api_id it was
// created with and shows the rest in the account's list of sessions.
type Identity struct {
	APIID   int
	APIHash string
	Device  telegram.DeviceConfig
}

// TDesktopWindows is the official Telegram Desktop for 64-bit Windows, as
// version 7.2.9 introduces itself (Telegram/SourceFiles/mtproto/
// session_private.cpp): the version gets " x64", the language pack is
// "tdesktop", and the time zone travels as tz_offset in params.
//
// A session created by Telegram Desktop has to be used under this identity:
// the server ends an authorization that shows up under another api_id.
var TDesktopWindows = Identity{
	APIID:   2040,
	APIHash: "b18441a1ff607e10a989891a5462e627",
	Device: telegram.DeviceConfig{
		// On Windows the model comes from the BIOS: manufacturer and product.
		DeviceModel:    "MSI MS-7C56",
		SystemVersion:  "Windows 10",
		AppVersion:     "7.2.9 x64",
		SystemLangCode: "ru-RU",
		LangPack:       "tdesktop",
		LangCode:       "ru",
	},
}

// withTimeZone returns i with the local time zone in the initConnection
// params, the way Telegram Desktop sends it.
func (i Identity) withTimeZone(now time.Time) Identity {
	_, offset := now.Zone()
	i.Device.Params = &tg.JSONObject{Value: []tg.JSONObjectValue{{
		Key:   "tz_offset",
		Value: &tg.JSONNumber{Value: float64(tzOffset(offset))},
	}}}
	return i
}

// tzOffset rounds a UTC offset in seconds as Telegram Desktop does: into
// [-12h, +14h] and to a quarter of an hour.
func tzOffset(seconds int) int {
	for seconds < -12*3600 {
		seconds += 24 * 3600
	}
	for seconds > 14*3600 {
		seconds -= 24 * 3600
	}
	sign := 1
	if seconds < 0 {
		sign, seconds = -1, -seconds
	}
	return sign * (seconds + 450) / 900 * 900
}
