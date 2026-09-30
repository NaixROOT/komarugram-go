// SPDX-License-Identifier: Unlicense OR MIT

package miniapps

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"komarugram/internal/messenger/model"
	"komarugram/pkg/miniapp"
)

type fakeStore struct {
	mu       sync.Mutex
	requests []model.WebViewRequest
	sent     []string
	prolongs int
	err      error
}

func (s *fakeStore) RequestWebView(_ context.Context, req model.WebViewRequest) (model.WebView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, req)
	if s.err != nil {
		return model.WebView{}, s.err
	}
	return model.WebView{URL: "https://app.example/x?a=1#tgWebAppData=query_id%3D1&tgWebAppStartParam=go", QueryID: 77}, nil
}
func (s *fakeStore) ProlongWebView(context.Context, model.WebViewRequest, int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prolongs++
	return nil
}
func (s *fakeStore) SendWebViewData(_ context.Context, bot int64, button, data string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, strings.Join([]string{itoa(bot), button, data}, "|"))
	return nil
}

// fakeApp is an app whose page a test speaks for.
type fakeApp struct {
	mu     sync.Mutex
	events []miniapp.Event
	sends  []string
	evals  []string
	closed bool
}

func (a *fakeApp) Events() []miniapp.Event {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]miniapp.Event(nil), a.events...)
}
func (a *fakeApp) Send(_ context.Context, t, data string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sends = append(a.sends, t+" "+data)
	return nil
}
func (a *fakeApp) Eval(_ context.Context, expr string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.evals = append(a.evals, expr)
	return "", nil
}
func (a *fakeApp) Running() bool { a.mu.Lock(); defer a.mu.Unlock(); return !a.closed }
func (a *fakeApp) Err() error    { return nil }
func (a *fakeApp) Close()        { a.mu.Lock(); a.closed = true; a.mu.Unlock() }
func (a *fakeApp) emit(t, data string) {
	a.mu.Lock()
	a.events = append(a.events, miniapp.Event{Type: t, Data: data, At: time.Now()})
	a.mu.Unlock()
}
func (a *fakeApp) got() (sends, evals []string, closed bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.sends...), append([]string(nil), a.evals...), a.closed
}

type harness struct {
	t      *testing.T
	store  *fakeStore
	app    *fakeApp
	r      *Runner
	mu     sync.Mutex
	links  []string
	failed []error
	url    string
	params miniapp.Params
	prof   miniapp.Profile
}

func newHarness(t *testing.T) *harness {
	h := &harness{t: t, store: &fakeStore{}, app: &fakeApp{}}
	h.r = New(h.store, Hooks{
		Link:   func(u string) { h.mu.Lock(); h.links = append(h.links, u); h.mu.Unlock() },
		Failed: func(err error) { h.mu.Lock(); h.failed = append(h.failed, err); h.mu.Unlock() },
	}, func() miniapp.Storage { return miniapp.PerApp }, "/profiles")
	h.r.open = func(_ context.Context, u string, p miniapp.Params, prof miniapp.Profile) (App, error) {
		h.mu.Lock()
		h.url, h.params, h.prof = u, p, prof
		h.mu.Unlock()
		return h.app, nil
	}
	t.Cleanup(h.r.Close)
	return h
}

func (h *harness) launch(kind model.WebViewKind) {
	h.r.Open(Launch{Request: model.WebViewRequest{Kind: kind, Chat: 9, Bot: 55, URL: "https://app.example/x"}, Button: "Open", Account: "acc"}, `{"bg_color":"#fff"}`)
}

func (h *harness) wait(what string, ok func() bool) {
	h.t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			h.t.Fatalf("timed out: %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

// The link Telegram answers with is opened with its fragment kept, and what
// it left out added.
func TestLaunchParams(t *testing.T) {
	base, p := LaunchParams("https://app.example/x?a=1#tgWebAppData=query_id%3D1&tgWebAppStartParam=go&tgWebAppVersion=8.0", `{"a":1}`)
	if base != "https://app.example/x?a=1" {
		t.Fatalf("base %q", base)
	}
	if p.InitData != "query_id%3D1" || p.Version != "8.0" || p.Platform != "tdesktop" || p.ThemeParams != "%7B%22a%22%3A1%7D" || p.Extra != "tgWebAppStartParam=go" {
		t.Fatalf("params %+v", p)
	}
	base, p = LaunchParams("https://app.example/", "")
	if base != "https://app.example/" || p.Version != "7.0" || p.InitData != "" || p.Extra != "" {
		t.Fatalf("without a fragment: %q %+v", base, p)
	}
}

// The runner asks Telegram, opens the page under the bot's own profile, and
// answers what the page asks.
func TestRunnerServesTheApp(t *testing.T) {
	h := newHarness(t)
	h.launch(model.WebViewInline)
	h.wait("the app to open", func() bool { h.mu.Lock(); defer h.mu.Unlock(); return h.url != "" })
	h.mu.Lock()
	if h.url != "https://app.example/x?a=1" || h.params.InitData != "query_id%3D1" || h.params.Extra != "tgWebAppStartParam=go" {
		t.Errorf("opened %q %+v", h.url, h.params)
	}
	if h.prof.Storage != miniapp.PerApp || h.prof.App != "55" || h.prof.Account != "acc" || h.prof.Root != "/profiles" {
		t.Errorf("profile %+v", h.prof)
	}
	h.mu.Unlock()
	if req := h.store.requests[0]; req.Kind != model.WebViewInline || req.Chat != 9 || req.Bot != 55 || req.Theme != `{"bg_color":"#fff"}` {
		t.Errorf("asked %+v", req)
	}

	h.app.emit("web_app_request_theme", "")
	h.app.emit("web_app_open_link", `{"url":"https://example.org/page"}`)
	h.app.emit("web_app_open_tg_link", `{"path_full":"/somebot?start=1"}`)
	h.app.emit("web_app_invoke_custom_method", `{"req_id":"r1","method":"getStorageValues","params":{}}`)
	h.wait("the answers", func() bool { s, _, _ := h.app.got(); return len(s) >= 2 })
	h.wait("the links", func() bool { h.mu.Lock(); defer h.mu.Unlock(); return len(h.links) == 2 })
	sends, _, closed := h.app.got()
	if closed {
		t.Fatal("the app was closed")
	}
	want := []string{`theme_changed {"theme_params":{"bg_color":"#fff"}}`, `custom_method_invoked {"req_id":"r1","error":"METHOD_INVALID"}`}
	for i, w := range want {
		if sends[i] != w {
			t.Errorf("sent %q, want %q", sends[i], w)
		}
	}
	if h.links[0] != "https://example.org/page" || h.links[1] != "https://t.me/somebot?start=1" {
		t.Errorf("links %v", h.links)
	}
}

// What a keyboard button's app hands over goes to its bot, and closes it.
func TestRunnerSendsData(t *testing.T) {
	h := newHarness(t)
	h.launch(model.WebViewSimple)
	h.wait("the app to open", func() bool { h.mu.Lock(); defer h.mu.Unlock(); return h.url != "" })
	h.app.emit("web_app_data_send", `{"data":"hello"}`)
	h.wait("the app to close", func() bool { _, _, c := h.app.got(); return c })
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	if len(h.store.sent) != 1 || h.store.sent[0] != "55|Open|hello" {
		t.Fatalf("sent %v", h.store.sent)
	}
}

// The app closes itself.
func TestRunnerCloses(t *testing.T) {
	h := newHarness(t)
	h.launch(model.WebViewMenu)
	h.wait("the app to open", func() bool { h.mu.Lock(); defer h.mu.Unlock(); return h.url != "" })
	h.app.emit("web_app_close", "")
	h.wait("the app to close", func() bool { _, _, c := h.app.got(); return c })
}

// The main and back buttons are drawn in the page, the main one with the
// state the app set; hiding it takes it away.
func TestRunnerDrawsButtons(t *testing.T) {
	h := newHarness(t)
	h.launch(model.WebViewInline)
	h.wait("the app to open", func() bool { h.mu.Lock(); defer h.mu.Unlock(); return h.url != "" })
	h.app.emit("web_app_setup_main_button", `{"is_visible":true,"is_active":true,"text":"Pay 5","color":"#123456","text_color":"#ffffff"}`)
	h.app.emit("web_app_setup_back_button", `{"is_visible":true}`)
	h.wait("both buttons", func() bool { _, e, _ := h.app.got(); return len(e) == 2 })
	_, evals, _ := h.app.got()
	if !strings.Contains(evals[0], `"text":"Pay 5"`) || !strings.Contains(evals[0], "main_button_pressed") {
		t.Errorf("main button script %q", evals[0])
	}
	if !strings.Contains(evals[1], "back_button_pressed") || !strings.HasSuffix(evals[1], "(true)") {
		t.Errorf("back button script %q", evals[1])
	}
	// The same state again draws nothing new.
	h.app.emit("web_app_setup_main_button", `{"is_visible":true,"is_active":true,"text":"Pay 5","color":"#123456","text_color":"#ffffff"}`)
	h.app.emit("web_app_setup_main_button", `{"is_visible":false}`)
	h.wait("the button to go", func() bool { _, e, _ := h.app.got(); return len(e) == 3 })
	_, evals, _ = h.app.got()
	if !strings.HasSuffix(evals[2], `({"is_visible":false})`) {
		t.Errorf("hiding script %q", evals[2])
	}
}

// A refused request is told, and no browser opens.
func TestRunnerFailure(t *testing.T) {
	h := newHarness(t)
	h.store.err = errors.New("BOT_INVALID")
	h.launch(model.WebViewInline)
	h.wait("the failure", func() bool { h.mu.Lock(); defer h.mu.Unlock(); return len(h.failed) == 1 })
	if h.url != "" {
		t.Fatal("a browser opened for a refused request")
	}
}
