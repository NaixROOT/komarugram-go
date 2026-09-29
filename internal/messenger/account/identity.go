// SPDX-License-Identifier: Unlicense OR MIT

package account

import (
	"os"
	"runtime"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"komarugram/pkg/deviceinfo"
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

// tdesktopVersion is the version of Telegram Desktop the client introduces
// itself as.
const tdesktopVersion = "7.2.9"

// TDesktop is the official Telegram Desktop on this computer, as version
// 7.2.9 introduces itself (Telegram/SourceFiles/mtproto/session_private.cpp,
// main/main_account.cpp): the device and the system as lib_base names them
// (see deviceinfo), the version with the processor (see appVersion), the
// language pack "tdesktop", and the time zone as tz_offset in params.
//
// A session created by Telegram Desktop has to be used under this identity:
// the server ends an authorization that shows up under another api_id.
func TDesktop() Identity {
	_, flatpak := os.Stat("/.flatpak-info")
	_, snap := os.LookupEnv("SNAP")
	return Identity{
		APIID:   2040,
		APIHash: "b18441a1ff607e10a989891a5462e627",
		Device: telegram.DeviceConfig{
			DeviceModel:    deviceinfo.Model(),
			SystemVersion:  deviceinfo.System(),
			AppVersion:     appVersion(tdesktopVersion, runtime.GOOS, runtime.GOARCH, flatpak == nil, snap),
			SystemLangCode: "ru-RU",
			LangPack:       "tdesktop",
			LangCode:       "ru",
		},
	}
}

// appVersion is version as Telegram Desktop sends it (ComputeAppVersion):
// with " x64" in a 64-bit Windows build, as it is in the other x86-64 builds
// and in 32-bit Windows, with the processor as Qt names it elsewhere; then
// the sandbox it runs in (KSandbox).
func appVersion(version, goos, goarch string, flatpak, snap bool) string {
	switch {
	case goos == "windows" && goarch == "amd64":
		version += " x64"
	case goos == "windows" && goarch == "386", goarch == "amd64":
	default:
		version += " " + qtArchitecture(goarch)
	}
	switch {
	case flatpak:
		version += " Flatpak"
	case snap:
		version += " Snap"
	}
	return version
}

// qtArchitecture is QSysInfo::buildCpuArchitecture for a GOARCH.
func qtArchitecture(goarch string) string {
	switch goarch {
	case "386":
		return "i386"
	case "loong64":
		return "loongarch64"
	case "mipsle":
		return "mips"
	case "mips64le":
		return "mips64"
	case "ppc64", "ppc64le":
		return "power64"
	}
	return goarch
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
