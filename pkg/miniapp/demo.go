// SPDX-License-Identifier: Unlicense OR MIT

package miniapp

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"time"
)

// assets holds the demo Mini App. telegram-web-app.js is Telegram's own SDK,
// fetched from https://telegram.org/js/telegram-web-app.js, and is here so the
// demo exercises the real thing rather than a stand-in.
//
//go:embed assets
var assets embed.FS

// Demo is a locally served Mini App and the launch parameters to open it with.
type Demo struct {
	URL      string
	Params   Params
	listener net.Listener
}

// ServeDemo publishes the bundled Mini App on loopback under a random path, so
// that other local processes cannot load it just by guessing the port.
func ServeDemo() (*Demo, error) {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}
	// Web storage is keyed by origin, and the origin of a loopback server is
	// its port. A port picked afresh on every launch would hand the demo a new
	// origin each time and hide the very difference the profile modes make, so
	// one is asked for by name, with any free port as a fallback.
	listener, err := net.Listen("tcp", "127.0.0.1:8730")
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		return nil, err
	}
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		_ = listener.Close()
		return nil, err
	}
	prefix := "/" + hex.EncodeToString(raw[:]) + "/"

	mux := http.NewServeMux()
	mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.FS(sub))))
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go server.Serve(listener) //nolint:errcheck // Serve ends with an error on close

	// A stand-in for what messages.requestWebView returns. The whole string is
	// escaped as one value: it is a query string of its own, and its inner
	// separators must not leak into the fragment.
	initData := url.QueryEscape(
		`query_id=AAkitchen&user={"id":42,"first_name":"Kitchen","username":"kitchen_user"}` +
			`&auth_date=1700000000&hash=demo`)

	return &Demo{
		URL: fmt.Sprintf("http://%s%sapp.html", listener.Addr(), prefix),
		Params: Params{
			InitData:    initData,
			Version:     "7.0",
			Platform:    "gio",
			ThemeParams: url.QueryEscape(`{"bg_color":"#fef7ff","text_color":"#1d1b20","hint_color":"#49454f","button_color":"#6750a4","button_text_color":"#ffffff"}`),
		},
		listener: listener,
	}, nil
}

// Close stops serving the demo.
func (d *Demo) Close() error { return d.listener.Close() }

// DarkTheme and LightTheme are the payloads the demo page sends back as
// theme_changed events.
const (
	DarkTheme  = `{theme_params:{bg_color:"#141218",text_color:"#e6e0e9",hint_color:"#cac4d0",button_color:"#d0bcff",button_text_color:"#381e72"}}`
	LightTheme = `{theme_params:{bg_color:"#fef7ff",text_color:"#1d1b20",hint_color:"#49454f",button_color:"#6750a4",button_text_color:"#ffffff"}}`
)
