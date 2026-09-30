// SPDX-License-Identifier: Unlicense OR MIT

package miniapps

import (
	"context"
	"encoding/json"
	"time"

	"komarugram/pkg/miniapp"
)

// run is one running Mini App and what the runner does for it.
type run struct {
	r       *Runner
	app     App
	launch  Launch
	queryID int64
	theme   string
	seen    int
	// main and back are the last state of the buttons drawn in the page.
	main, back string
}

// loop serves the app until its window closes: it answers what the page
// asks, and keeps an inline launch alive.
func (a *run) loop() {
	tick := time.NewTicker(poll)
	defer tick.Stop()
	prolong := time.NewTicker(prolongEvery)
	defer prolong.Stop()
	for {
		select {
		case <-a.r.ctx.Done():
			return
		case <-prolong.C:
			if a.queryID != 0 {
				ctx, cancel := context.WithTimeout(a.r.ctx, time.Minute)
				_ = a.r.store.ProlongWebView(ctx, a.launch.Request, a.queryID)
				cancel()
			}
		case <-tick.C:
			if !a.app.Running() {
				return
			}
			events := a.app.Events()
			for ; a.seen < len(events); a.seen++ {
				if a.handle(events[a.seen]) {
					return
				}
			}
			if err := a.app.Err(); err != nil && !a.app.Running() {
				return
			}
		}
	}
}

// send tells the page something; a page that is gone is not an error worth
// telling.
func (a *run) send(eventType, data string) {
	ctx, cancel := context.WithTimeout(a.r.ctx, 10*time.Second)
	defer cancel()
	_ = a.app.Send(ctx, eventType, data)
}

// eval runs a script in the page.
func (a *run) eval(script string) {
	ctx, cancel := context.WithTimeout(a.r.ctx, 10*time.Second)
	defer cancel()
	_, _ = a.app.Eval(ctx, script)
}

// handle does what the page asked, and reports whether the app is done.
func (a *run) handle(e miniapp.Event) (done bool) {
	var data map[string]any
	if e.Data != "" {
		_ = json.Unmarshal([]byte(e.Data), &data)
	}
	str := func(key string) string { s, _ := data[key].(string); return s }
	flag := func(key string) bool { b, _ := data[key].(bool); return b }
	switch e.Type {
	case "web_app_close":
		return true
	case "web_app_request_viewport":
		a.send("viewport_changed", `{"height":720,"is_state_stable":true,"is_expanded":true}`)
	case "web_app_request_theme":
		a.send("theme_changed", `{"theme_params":`+a.theme+`}`)
	case "web_app_open_link":
		if link := str("url"); link != "" && a.r.hooks.Link != nil {
			a.r.hooks.Link(link)
		}
	case "web_app_open_tg_link":
		if path := str("path_full"); path != "" && a.r.hooks.Link != nil {
			a.r.hooks.Link("https://t.me" + path)
		}
	case "web_app_data_send":
		// A keyboard button's app hands its data to the bot and is done.
		text := str("data")
		ctx, cancel := context.WithTimeout(a.r.ctx, time.Minute)
		err := a.r.store.SendWebViewData(ctx, a.launch.Request.Bot, a.launch.Button, text)
		cancel()
		if err != nil && a.r.hooks.Failed != nil {
			a.r.hooks.Failed(err)
		}
		return true
	case "web_app_setup_main_button":
		a.setMain(data)
	case "web_app_setup_back_button":
		a.setBack(flag("is_visible"))
	case "web_app_invoke_custom_method":
		// Cloud storage and the like are Telegram's to serve; an app that asks
		// is told the method is not there rather than left waiting.
		a.send("custom_method_invoked", `{"req_id":`+jsonString(str("req_id"))+`,"error":"METHOD_INVALID"}`)
	case "web_app_request_write_access":
		a.send("write_access_requested", `{"status":"cancelled"}`)
	case "web_app_request_phone":
		a.send("phone_requested", `{"status":"cancelled"}`)
	case "web_app_open_popup":
		a.send("popup_closed", `{}`)
	case "web_app_read_text_from_clipboard":
		a.send("clipboard_text_received", `{"req_id":`+jsonString(str("req_id"))+`,"data":null}`)
	}
	return false
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// setMain draws the app's main button at the bottom of the page.
func (a *run) setMain(data map[string]any) {
	visible, _ := data["is_visible"].(bool)
	state := map[string]any{"is_visible": visible}
	for _, key := range []string{"text", "color", "text_color"} {
		if s, ok := data[key].(string); ok {
			state[key] = s
		}
	}
	for _, key := range []string{"is_active", "is_progress_visible"} {
		if b, ok := data[key].(bool); ok {
			state[key] = b
		}
	}
	js, _ := json.Marshal(state)
	if string(js) == a.main {
		return
	}
	a.main = string(js)
	a.eval(mainButtonScript + "(" + string(js) + ")")
}

// setBack draws the app's back button at the top of the page.
func (a *run) setBack(visible bool) {
	state := "hidden"
	if visible {
		state = "shown"
	}
	if state == a.back {
		return
	}
	a.back = state
	if visible {
		a.eval(backButtonScript + "(true)")
	} else {
		a.eval(backButtonScript + "(false)")
	}
}

// mainButtonScript is a function that draws, updates or removes the main
// button in the page; the button's press goes back as the SDK's own event.
const mainButtonScript = `(function (p) {
  var id = '__tg_main_button', el = document.getElementById(id);
  if (!p.is_visible) { if (el) el.remove(); document.body.style.paddingBottom = ''; return; }
  if (!el) {
    el = document.createElement('button');
    el.id = id;
    el.style.cssText = 'position:fixed;left:0;right:0;bottom:0;height:52px;z-index:2147483647;border:0;' +
      'font:600 15px system-ui,sans-serif;letter-spacing:.5px;text-transform:uppercase;cursor:pointer';
    el.onclick = function () { if (!el.disabled) window.Telegram.WebView.receiveEvent('main_button_pressed', null); };
    document.body.appendChild(el);
  }
  el.textContent = (p.text || '') + (p.is_progress_visible ? ' …' : '');
  el.style.background = p.color || '#2481cc';
  el.style.color = p.text_color || '#ffffff';
  el.disabled = p.is_active === false;
  el.style.opacity = el.disabled ? '0.5' : '1';
  document.body.style.paddingBottom = '52px';
})`

// backButtonScript is a function that draws or removes the back button.
const backButtonScript = `(function (visible) {
  var id = '__tg_back_button', el = document.getElementById(id);
  if (!visible) { if (el) el.remove(); return; }
  if (el) return;
  el = document.createElement('button');
  el.id = id;
  el.textContent = '←';
  el.setAttribute('aria-label', 'Back');
  el.style.cssText = 'position:fixed;left:8px;top:8px;width:40px;height:40px;z-index:2147483647;border:0;' +
    'border-radius:20px;background:rgba(0,0,0,.45);color:#fff;font:20px system-ui,sans-serif;cursor:pointer';
  el.onclick = function () { window.Telegram.WebView.receiveEvent('back_button_pressed', null); };
  document.body.appendChild(el);
})`
